package postgres

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"cybertoolkit/backend/internal/domain"
)

type Store struct {
	pool          *pgxpool.Pool
	redis         *redis.Client
	adminUsername string
	adminEmail    string
	adminPassword string
	adminName     string
}

func NewStore(databaseURL, redisURL, adminUsername, adminEmail, adminPassword, adminName string) (*Store, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	redisOpts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	rdb := redis.NewClient(redisOpts)
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	s := &Store{
		pool:          pool,
		redis:         rdb,
		adminUsername: adminUsername,
		adminEmail:    adminEmail,
		adminPassword: adminPassword,
		adminName:     adminName,
	}

	if err := s.migrate(ctx); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	if err := s.seed(ctx); err != nil {
		return nil, fmt.Errorf("seed: %w", err)
	}

	if err := s.ensureAdmin(ctx); err != nil {
		return nil, fmt.Errorf("ensure admin: %w", err)
	}

	return s, nil
}

// ensureAdmin guarantees the admin user configured via env vars exists with
// the configured username, email, password and name on every startup.
func (s *Store) ensureAdmin(ctx context.Context) error {
	username := strings.TrimSpace(s.adminUsername)
	password := s.adminPassword
	if username == "" || password == "" {
		return nil
	}

	email := strings.TrimSpace(s.adminEmail)
	name := strings.TrimSpace(s.adminName)
	if name == "" {
		name = "Admin"
	}

	now := time.Now().UTC()
	tag, err := s.pool.Exec(ctx,
		`UPDATE users SET email = $1, password_hash = $2, display_name = $3, role = 'admin', is_active = true, updated_at = $4
		 WHERE username = $5`,
		email, hashPassword(password), name, now, username,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() > 0 {
		return nil
	}

	_, err = s.pool.Exec(ctx,
		`INSERT INTO users (username, email, password_hash, display_name, role, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, 'admin', $5, $5)`,
		username, email, hashPassword(password), name, now,
	)
	return err
}

func (s *Store) migrate(ctx context.Context) error {
	schemaPath := "sql/schema.sql"
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		schemaPath = "../sql/schema.sql"
	}

	data, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("read schema: %w", err)
	}

	// Execute schema statements one by one to avoid protocol issues
	statements := splitSQLStatements(string(data))
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := s.pool.Exec(ctx, stmt); err != nil {
			// Ignore "already exists" errors
			msg := strings.ToLower(err.Error())
			if strings.Contains(msg, "already exists") {
				continue
			}
			return fmt.Errorf("exec schema: %w", err)
		}
	}

	// Migration: add 'type' column to tool_submissions if it doesn't exist
	var hasTypeCol int
	_ = s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM information_schema.columns WHERE table_name = 'tool_submissions' AND column_name = 'type'`).Scan(&hasTypeCol)
	if hasTypeCol == 0 {
		if _, err := s.pool.Exec(ctx, `ALTER TABLE tool_submissions ADD COLUMN type varchar(50) not null default 'tool'`); err != nil {
			fmt.Fprintf(os.Stderr, "Migration warning: could not add type column: %v\n", err)
		} else {
			fmt.Println("Migration: added 'type' column to tool_submissions")
		}
	}

	// Migration: add comment for 'type' column if missing
	var hasTypeComment int
	_ = s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM pg_col_description WHERE objoid = 'tool_submissions'::regclass AND objsubid = (SELECT ordinal_position FROM information_schema.columns WHERE table_name = 'tool_submissions' AND column_name = 'type')`).Scan(&hasTypeComment)
	if hasTypeComment == 0 {
		if _, err := s.pool.Exec(ctx, `COMMENT ON COLUMN tool_submissions.type IS '投稿类型：tool / feature / other'`); err != nil {
			fmt.Fprintf(os.Stderr, "Migration warning: could not add type column comment: %v\n", err)
		}
	}

	return nil
}

func splitSQLStatements(sql string) []string {
	var stmts []string
	var current strings.Builder
	lines := strings.Split(sql, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			continue
		}
		current.WriteString(line)
		current.WriteString("\n")
		if strings.HasSuffix(trimmed, ";") {
			stmts = append(stmts, current.String())
			current.Reset()
		}
	}
	if current.Len() > 0 {
		stmts = append(stmts, current.String())
	}
	return stmts
}

func (s *Store) seed(ctx context.Context) error {
	var count int
	if err := s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM categories").Scan(&count); err != nil {
		// Table might not exist yet (shouldn't happen after migrate)
		return nil
	}
	if count > 0 {
		return nil
	}

	now := time.Now().UTC()
	publishedAt := now.Add(-24 * time.Hour)

	categories := []domain.Category{
		{Slug: "network-scanning", Name: "Network Scanning", Description: "Discover hosts, ports, services and exposed assets.", Icon: "Radar", SortOrder: 1, IsVisible: true, CreatedAt: now, UpdatedAt: now},
		{Slug: "vulnerability-assessment", Name: "Vulnerability Assessment", Description: "Identify vulnerabilities, weak configurations and risks.", Icon: "ShieldAlert", SortOrder: 2, IsVisible: true, CreatedAt: now, UpdatedAt: now},
		{Slug: "penetration-testing", Name: "Penetration Testing", Description: "Validate attack paths and defensive controls.", Icon: "Crosshair", SortOrder: 3, IsVisible: true, CreatedAt: now, UpdatedAt: now},
		{Slug: "osint", Name: "OSINT", Description: "Collect intelligence from public data sources.", Icon: "Eye", SortOrder: 4, IsVisible: true, CreatedAt: now, UpdatedAt: now},
	}

	tags := []domain.Tag{
		{Slug: "port-scan", Name: "Port Scan", CreatedAt: now},
		{Slug: "host-discovery", Name: "Host Discovery", CreatedAt: now},
		{Slug: "automation", Name: "Automation", CreatedAt: now},
		{Slug: "web-security", Name: "Web Security", CreatedAt: now},
		{Slug: "asset-discovery", Name: "Asset Discovery", CreatedAt: now},
	}

	tools := []domain.Tool{
		{Slug: "nmap", Name: "Nmap", ShortDescription: "Industry-standard network discovery and audit tool.", LongDescription: "Nmap supports host discovery, port scanning, service identification and scriptable security checks.", CategoryID: "", Difficulty: domain.DifficultyIntermediate, Icon: "Radar", Featured: true, Status: domain.ToolStatusPublished, WebsiteURL: "https://nmap.org", GitHubURL: "https://github.com/nmap/nmap", ViewCount: 1200, FavoriteCount: 230, PublishedAt: &publishedAt, CreatedAt: now, UpdatedAt: now},
		{Slug: "nuclei", Name: "Nuclei", ShortDescription: "Fast template-driven vulnerability scanner.", LongDescription: "Nuclei uses YAML templates to scan for CVEs, exposed panels, misconfigurations and fingerprints.", CategoryID: "", Difficulty: domain.DifficultyIntermediate, Icon: "Atom", Featured: true, Status: domain.ToolStatusPublished, WebsiteURL: "https://nuclei.projectdiscovery.io", GitHubURL: "https://github.com/projectdiscovery/nuclei", ViewCount: 980, FavoriteCount: 180, PublishedAt: &publishedAt, CreatedAt: now, UpdatedAt: now},
		{Slug: "burpsuite", Name: "Burp Suite", ShortDescription: "Core platform for web application security testing.", LongDescription: "Burp Suite combines proxying, replay, scanning and traffic analysis for web testing workflows.", CategoryID: "", Difficulty: domain.DifficultyIntermediate, Icon: "Bug", Featured: true, Status: domain.ToolStatusPublished, WebsiteURL: "https://portswigger.net/burp", GitHubURL: "https://github.com/portswigger", ViewCount: 870, FavoriteCount: 150, PublishedAt: &publishedAt, CreatedAt: now, UpdatedAt: now},
		{Slug: "shodan", Name: "Shodan", ShortDescription: "Search engine for exposed devices and internet-facing services.", LongDescription: "Shodan helps discover exposed services, fingerprints and assets visible on the public internet.", CategoryID: "", Difficulty: domain.DifficultyBeginner, Icon: "Eye", Featured: true, Status: domain.ToolStatusPublished, WebsiteURL: "https://www.shodan.io", GitHubURL: "https://github.com/achillean/shodan-python", ViewCount: 1100, FavoriteCount: 210, PublishedAt: &publishedAt, CreatedAt: now, UpdatedAt: now},
	}

	toolTags := map[string][]string{
		"nmap":      {"port-scan", "host-discovery"},
		"nuclei":    {"automation", "asset-discovery"},
		"burpsuite": {"web-security"},
		"shodan":    {"asset-discovery"},
	}

	users := []struct {
		Username    string
		Email       string
		Password    string
		DisplayName string
		Role        string
	}{
		{Username: s.adminUsername, Email: s.adminEmail, Password: s.adminPassword, DisplayName: s.adminName, Role: "admin"},
		{Username: "editor", Email: "editor@cybertoolkit.local", Password: "editor123456", DisplayName: "Editor Demo", Role: "editor"},
		{Username: "viewer", Email: "viewer@cybertoolkit.local", Password: "viewer123456", DisplayName: "Viewer Demo", Role: "viewer"},
	}

	// Insert categories and collect IDs by slug
	catIDs := make(map[string]string)
	for _, c := range categories {
		var id string
		err := s.pool.QueryRow(ctx,
			`INSERT INTO categories (slug, name, description, icon, sort_order, is_visible, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id::text`,
			c.Slug, c.Name, c.Description, c.Icon, c.SortOrder, c.IsVisible, c.CreatedAt, c.UpdatedAt,
		).Scan(&id)
		if err != nil {
			return fmt.Errorf("seed category %s: %w", c.Slug, err)
		}
		catIDs[c.Slug] = id
	}

	// Insert tags and collect IDs by slug
	tagIDs := make(map[string]string)
	for _, t := range tags {
		var id string
		err := s.pool.QueryRow(ctx,
			`INSERT INTO tags (slug, name, created_at) VALUES ($1, $2, $3) RETURNING id::text`,
			t.Slug, t.Name, t.CreatedAt,
		).Scan(&id)
		if err != nil {
			return fmt.Errorf("seed tag %s: %w", t.Slug, err)
		}
		tagIDs[t.Slug] = id
	}

	// Insert tools and collect IDs by slug
	toolIDs := make(map[string]string)
	for i := range tools {
		t := &tools[i]
		switch t.Slug {
		case "nmap", "nuclei":
			t.CategoryID = catIDs["network-scanning"]
		case "burpsuite":
			t.CategoryID = catIDs["penetration-testing"]
		case "shodan":
			t.CategoryID = catIDs["osint"]
		}
		var id string
		err := s.pool.QueryRow(ctx,
			`INSERT INTO tools (slug, name, short_description, long_description, category_id, difficulty, icon, featured, status, website_url, github_url, view_count, favorite_count, published_at, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16) RETURNING id::text`,
			t.Slug, t.Name, t.ShortDescription, t.LongDescription, t.CategoryID, t.Difficulty, t.Icon, t.Featured, t.Status, t.WebsiteURL, t.GitHubURL, t.ViewCount, t.FavoriteCount, t.PublishedAt, t.CreatedAt, t.UpdatedAt,
		).Scan(&id)
		if err != nil {
			return fmt.Errorf("seed tool %s: %w", t.Slug, err)
		}
		toolIDs[t.Slug] = id
	}

	// Insert tool_tags
	for toolSlug, tagSlugs := range toolTags {
		for _, tagSlug := range tagSlugs {
			_, err := s.pool.Exec(ctx,
				`INSERT INTO tool_tags (tool_id, tag_id) VALUES ($1, $2)`,
				toolIDs[toolSlug], tagIDs[tagSlug],
			)
			if err != nil {
				return fmt.Errorf("seed tool_tag %s-%s: %w", toolSlug, tagSlug, err)
			}
		}
	}

	// Insert users
	for _, u := range users {
		_, err := s.pool.Exec(ctx,
			`INSERT INTO users (username, email, password_hash, display_name, role, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $6)`,
			u.Username, u.Email, hashPassword(u.Password), u.DisplayName, u.Role, now,
		)
		if err != nil {
			return fmt.Errorf("seed user %s: %w", u.Email, err)
		}
	}

	return nil
}

func (s *Store) Stats() map[string]int {
	ctx := context.Background()
	var toolCount, categoryCount, featuredCount int

	_ = s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM tools").Scan(&toolCount)
	_ = s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM categories").Scan(&categoryCount)
	_ = s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM tools WHERE status = 'published' AND featured = true").Scan(&featuredCount)

	return map[string]int{
		"toolCount":     toolCount,
		"categoryCount": categoryCount,
		"featuredCount": featuredCount,
	}
}

func (s *Store) ListCategories(visibleOnly bool) []domain.Category {
	ctx := context.Background()
	query := `SELECT id::text, slug, name, description, icon, sort_order, is_visible, created_at, updated_at FROM categories ORDER BY sort_order`
	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var out []domain.Category
	for rows.Next() {
		var c domain.Category
		if err := rows.Scan(&c.ID, &c.Slug, &c.Name, &c.Description, &c.Icon, &c.SortOrder, &c.IsVisible, &c.CreatedAt, &c.UpdatedAt); err != nil {
			continue
		}
		if visibleOnly && !c.IsVisible {
			continue
		}
		out = append(out, c)
	}
	return out
}

func (s *Store) CategoryByID(id string) (domain.Category, bool) {
	ctx := context.Background()
	var c domain.Category
	err := s.pool.QueryRow(ctx,
		`SELECT id::text, slug, name, description, icon, sort_order, is_visible, created_at, updated_at FROM categories WHERE id = $1`,
		id,
	).Scan(&c.ID, &c.Slug, &c.Name, &c.Description, &c.Icon, &c.SortOrder, &c.IsVisible, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return domain.Category{}, false
	}
	return c, true
}

func (s *Store) CreateCategory(input domain.Category) domain.Category {
	ctx := context.Background()
	now := time.Now().UTC()
	var id string
	_ = s.pool.QueryRow(ctx,
		`INSERT INTO categories (slug, name, description, icon, sort_order, is_visible, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $7) RETURNING id::text`,
		input.Slug, input.Name, input.Description, input.Icon, input.SortOrder, input.IsVisible, now,
	).Scan(&id)
	input.ID = id
	input.CreatedAt = now
	input.UpdatedAt = now
	return input
}

func (s *Store) UpdateCategory(id string, update domain.Category) (domain.Category, error) {
	ctx := context.Background()
	var c domain.Category
	err := s.pool.QueryRow(ctx,
		`UPDATE categories SET slug=$1, name=$2, description=$3, icon=$4, sort_order=$5, is_visible=$6, updated_at=now()
		 WHERE id=$7 RETURNING id::text, slug, name, description, icon, sort_order, is_visible, created_at, updated_at`,
		update.Slug, update.Name, update.Description, update.Icon, update.SortOrder, update.IsVisible, id,
	).Scan(&c.ID, &c.Slug, &c.Name, &c.Description, &c.Icon, &c.SortOrder, &c.IsVisible, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return domain.Category{}, errors.New("category not found")
	}
	return c, nil
}

func (s *Store) ListTags() []domain.Tag {
	ctx := context.Background()
	rows, err := s.pool.Query(ctx, `SELECT id::text, slug, name, created_at FROM tags ORDER BY name`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var out []domain.Tag
	for rows.Next() {
		var t domain.Tag
		if err := rows.Scan(&t.ID, &t.Slug, &t.Name, &t.CreatedAt); err != nil {
			continue
		}
		out = append(out, t)
	}
	return out
}

func (s *Store) ListTools(filters domain.ToolFilters, admin bool) ([]domain.Tool, int) {
	ctx := context.Background()

	where := []string{"1=1"}
	args := []any{}
	argIdx := 1

	if !admin {
		where = append(where, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, "published")
		argIdx++
	}
	if admin && filters.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, filters.Status)
		argIdx++
	}
	if filters.Query != "" {
		where = append(where, fmt.Sprintf("(name ILIKE $%d OR short_description ILIKE $%d OR long_description ILIKE $%d)", argIdx, argIdx, argIdx))
		args = append(args, "%"+filters.Query+"%")
		argIdx++
	}
	if filters.Category != "" {
		where = append(where, fmt.Sprintf("category_id = (SELECT id FROM categories WHERE slug = $%d)", argIdx))
		args = append(args, filters.Category)
		argIdx++
	}
	if filters.Difficulty != "" {
		where = append(where, fmt.Sprintf("difficulty = $%d", argIdx))
		args = append(args, filters.Difficulty)
		argIdx++
	}
	if filters.Featured != nil {
		where = append(where, fmt.Sprintf("featured = $%d", argIdx))
		args = append(args, *filters.Featured)
		argIdx++
	}
	if filters.Tag != "" {
		where = append(where, fmt.Sprintf("id IN (SELECT tool_id FROM tool_tags WHERE tag_id = (SELECT id FROM tags WHERE slug = $%d))", argIdx))
		args = append(args, filters.Tag)
		argIdx++
	}

	query := `SELECT id::text, slug, name, short_description, long_description, category_id::text, difficulty, icon, featured, status, website_url, github_url, view_count, favorite_count, published_at, created_at, updated_at FROM tools WHERE ` + strings.Join(where, " AND ")

	switch filters.Sort {
	case "name":
		query += " ORDER BY name"
	case "popular":
		query += " ORDER BY view_count DESC"
	default:
		query += " ORDER BY created_at DESC"
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0
	}
	defer rows.Close()

	var all []domain.Tool
	for rows.Next() {
		var t domain.Tool
		if err := rows.Scan(&t.ID, &t.Slug, &t.Name, &t.ShortDescription, &t.LongDescription, &t.CategoryID, &t.Difficulty, &t.Icon, &t.Featured, &t.Status, &t.WebsiteURL, &t.GitHubURL, &t.ViewCount, &t.FavoriteCount, &t.PublishedAt, &t.CreatedAt, &t.UpdatedAt); err != nil {
			continue
		}
		all = append(all, t)
	}

	page := filters.Page
	if page < 1 {
		page = 1
	}
	pageSize := filters.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	total := len(all)
	start := (page - 1) * pageSize
	if start >= total {
		return []domain.Tool{}, total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return all[start:end], total
}

func (s *Store) GetToolBySlug(slug string, admin bool) (domain.Tool, bool) {
	ctx := context.Background()
	var t domain.Tool
	var statusFilter string
	if admin {
		statusFilter = "1=1"
	} else {
		statusFilter = "status = 'published'"
	}
	err := s.pool.QueryRow(ctx,
		`SELECT id::text, slug, name, short_description, long_description, category_id::text, difficulty, icon, featured, status, website_url, github_url, view_count, favorite_count, published_at, created_at, updated_at FROM tools WHERE slug = $1 AND `+statusFilter,
		slug,
	).Scan(&t.ID, &t.Slug, &t.Name, &t.ShortDescription, &t.LongDescription, &t.CategoryID, &t.Difficulty, &t.Icon, &t.Featured, &t.Status, &t.WebsiteURL, &t.GitHubURL, &t.ViewCount, &t.FavoriteCount, &t.PublishedAt, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return domain.Tool{}, false
	}
	return t, true
}

func (s *Store) GetToolByID(id string) (domain.Tool, bool) {
	ctx := context.Background()
	var t domain.Tool
	err := s.pool.QueryRow(ctx,
		`SELECT id::text, slug, name, short_description, long_description, category_id::text, difficulty, icon, featured, status, website_url, github_url, view_count, favorite_count, published_at, created_at, updated_at FROM tools WHERE id = $1`,
		id,
	).Scan(&t.ID, &t.Slug, &t.Name, &t.ShortDescription, &t.LongDescription, &t.CategoryID, &t.Difficulty, &t.Icon, &t.Featured, &t.Status, &t.WebsiteURL, &t.GitHubURL, &t.ViewCount, &t.FavoriteCount, &t.PublishedAt, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return domain.Tool{}, false
	}
	return t, true
}

func (s *Store) CreateTool(input domain.Tool) domain.Tool {
	ctx := context.Background()
	now := time.Now().UTC()
	var publishedAt interface{}
	if input.Status == domain.ToolStatusPublished {
		publishedAt = now
	} else {
		publishedAt = nil
	}
	var id string
	_ = s.pool.QueryRow(ctx,
		`INSERT INTO tools (slug, name, short_description, long_description, category_id, difficulty, icon, featured, status, website_url, github_url, published_at, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $13) RETURNING id::text`,
		input.Slug, input.Name, input.ShortDescription, input.LongDescription, input.CategoryID, input.Difficulty, input.Icon, input.Featured, input.Status, input.WebsiteURL, input.GitHubURL, publishedAt, now,
	).Scan(&id)
	input.ID = id
	input.CreatedAt = now
	input.UpdatedAt = now
	if input.Status == domain.ToolStatusPublished {
		input.PublishedAt = &now
	}
	return input
}

func (s *Store) UpdateTool(id string, update domain.Tool) (domain.Tool, error) {
	ctx := context.Background()
	var t domain.Tool
	now := time.Now().UTC()
	var publishedAt interface{}
	if update.Status == domain.ToolStatusPublished {
		publishedAt = now
	} else {
		publishedAt = update.PublishedAt
	}
	err := s.pool.QueryRow(ctx,
		`UPDATE tools SET slug=$1, name=$2, short_description=$3, long_description=$4, category_id=$5, difficulty=$6, icon=$7, featured=$8, status=$9, website_url=$10, github_url=$11, published_at=$12, updated_at=$13
		 WHERE id=$14 RETURNING id::text, slug, name, short_description, long_description, category_id::text, difficulty, icon, featured, status, website_url, github_url, view_count, favorite_count, published_at, created_at, updated_at`,
		update.Slug, update.Name, update.ShortDescription, update.LongDescription, update.CategoryID, update.Difficulty, update.Icon, update.Featured, update.Status, update.WebsiteURL, update.GitHubURL, publishedAt, now, id,
	).Scan(&t.ID, &t.Slug, &t.Name, &t.ShortDescription, &t.LongDescription, &t.CategoryID, &t.Difficulty, &t.Icon, &t.Featured, &t.Status, &t.WebsiteURL, &t.GitHubURL, &t.ViewCount, &t.FavoriteCount, &t.PublishedAt, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return domain.Tool{}, errors.New("tool not found")
	}
	return t, nil
}

func (s *Store) ArchiveTool(id string) error {
	ctx := context.Background()
	tag, err := s.pool.Exec(ctx,
		`UPDATE tools SET status='archived', updated_at=now() WHERE id=$1`,
		id,
	)
	if err != nil || tag.RowsAffected() == 0 {
		return errors.New("tool not found")
	}
	return nil
}

func (s *Store) TagsForTool(toolID string) []domain.Tag {
	ctx := context.Background()
	rows, err := s.pool.Query(ctx,
		`SELECT t.id::text, t.slug, t.name, t.created_at FROM tags t JOIN tool_tags tt ON t.id = tt.tag_id WHERE tt.tool_id = $1 ORDER BY t.name`,
		toolID,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var out []domain.Tag
	for rows.Next() {
		var t domain.Tag
		if err := rows.Scan(&t.ID, &t.Slug, &t.Name, &t.CreatedAt); err != nil {
			continue
		}
		out = append(out, t)
	}
	return out
}

func (s *Store) RelatedTools(categoryID, exceptToolID string, limit int) []domain.Tool {
	ctx := context.Background()
	rows, err := s.pool.Query(ctx,
		`SELECT id::text, slug, name, short_description, long_description, category_id::text, difficulty, icon, featured, status, website_url, github_url, view_count, favorite_count, published_at, created_at, updated_at FROM tools WHERE category_id=$1 AND id != $2 AND status='published' LIMIT $3`,
		categoryID, exceptToolID, limit,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var out []domain.Tool
	for rows.Next() {
		var t domain.Tool
		if err := rows.Scan(&t.ID, &t.Slug, &t.Name, &t.ShortDescription, &t.LongDescription, &t.CategoryID, &t.Difficulty, &t.Icon, &t.Featured, &t.Status, &t.WebsiteURL, &t.GitHubURL, &t.ViewCount, &t.FavoriteCount, &t.PublishedAt, &t.CreatedAt, &t.UpdatedAt); err != nil {
			continue
		}
		out = append(out, t)
		if len(out) == limit {
			break
		}
	}
	return out
}

func (s *Store) ReplaceToolTags(toolID string, tagNames []string) {
	ctx := context.Background()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return
	}
	defer tx.Rollback(ctx)

	_, _ = tx.Exec(ctx, `DELETE FROM tool_tags WHERE tool_id = $1`, toolID)

	for _, name := range dedupeStrings(tagNames) {
		var tagID string
		err := tx.QueryRow(ctx,
			`INSERT INTO tags (slug, name, created_at) VALUES ($1, $2, $3)
			 ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name RETURNING id::text`,
			slugify(name), name, time.Now().UTC(),
		).Scan(&tagID)
		if err != nil {
			continue
		}
		_, _ = tx.Exec(ctx, `INSERT INTO tool_tags (tool_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, toolID, tagID)
	}

	_ = tx.Commit(ctx)
}

func (s *Store) CreateSubmission(submission domain.Submission) domain.Submission {
	ctx := context.Background()
	now := time.Now().UTC()
	var id string

	submittedBy := nilIfEmpty(submission.SubmittedBy)
	toolID := nilIfEmpty(submission.ToolID)

	err := s.pool.QueryRow(ctx,
		`INSERT INTO tool_submissions (type, submitted_by, tool_id, submitter_email, payload, status, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id::text`,
		submission.Type, submittedBy, toolID, submission.SubmitterEmail, submission.Payload, submission.Status, now,
	).Scan(&id)
	if err != nil {
		// Log error but don't panic — caller should check for empty ID
		fmt.Fprintf(os.Stderr, "CreateSubmission error: %v\n", err)
	}
	submission.ID = id
	submission.CreatedAt = now
	return submission
}

func (s *Store) SubmissionsByUser(userID string) ([]domain.Submission, int) {
	ctx := context.Background()

	var total int
	_ = s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM tool_submissions WHERE submitted_by = $1`, userID).Scan(&total)

	rows, err := s.pool.Query(ctx,
		`SELECT id::text, type, COALESCE(submitted_by::text, ''), COALESCE(tool_id::text, ''), COALESCE(submitter_email, ''), payload, status, COALESCE(reviewer_id::text, ''), COALESCE(review_note, ''), created_at, reviewed_at FROM tool_submissions WHERE submitted_by = $1 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, total
	}
	defer rows.Close()

	var out []domain.Submission
	for rows.Next() {
		var sub domain.Submission
		if err := rows.Scan(&sub.ID, &sub.Type, &sub.SubmittedBy, &sub.ToolID, &sub.SubmitterEmail, &sub.Payload, &sub.Status, &sub.ReviewerID, &sub.ReviewNote, &sub.CreatedAt, &sub.ReviewedAt); err != nil {
			continue
		}
		out = append(out, sub)
	}
	return out, total
}

const sessionTTL = 24 * time.Hour

func sessKey(token string) string    { return "sess:" + token }
func refreshKey(token string) string { return "refresh:" + token }
func userSessKey(userID string) string { return "user_sessions:" + userID }

func (s *Store) Authenticate(account, password, ip, userAgent string) (string, string, domain.User, error) {
	ctx := context.Background()
	var user domain.User

	// If account contains '@', treat as email; otherwise treat as username
	var query string
	if strings.Contains(account, "@") {
		query = `SELECT id::text, username, email, display_name, password_hash, role, is_active, last_login_at, created_at FROM users WHERE email = $1 AND is_active = true`
	} else {
		query = `SELECT id::text, username, email, display_name, password_hash, role, is_active, last_login_at, created_at FROM users WHERE username = $1 AND is_active = true`
	}

	err := s.pool.QueryRow(ctx, query, account).Scan(&user.ID, &user.Username, &user.Email, &user.DisplayName, &user.PasswordHash, &user.Role, &user.IsActive, &user.LastLoginAt, &user.CreatedAt)
	if err != nil {
		return "", "", domain.User{}, errors.New("invalid username/email or password")
	}
	if user.PasswordHash != hashPassword(password) {
		return "", "", domain.User{}, errors.New("invalid username/email or password")
	}

	accessToken, err := randomToken()
	if err != nil {
		return "", "", domain.User{}, err
	}
	refreshToken, err := randomToken()
	if err != nil {
		return "", "", domain.User{}, err
	}

	now := time.Now().UTC()
	_, _ = s.pool.Exec(ctx, `UPDATE users SET last_login_at = $1 WHERE id = $2`, now, user.ID)

	if err := s.writeSession(ctx, accessToken, refreshToken, user, ip, userAgent, now); err != nil {
		return "", "", domain.User{}, err
	}

	return accessToken, refreshToken, publicUser(user), nil
}

func (s *Store) Register(username, email, password, displayName, ip, userAgent string) (string, string, domain.User, error) {
	ctx := context.Background()

	var exists int
	_ = s.pool.QueryRow(ctx, `SELECT 1 FROM users WHERE username = $1`, username).Scan(&exists)
	if exists == 1 {
		return "", "", domain.User{}, errors.New("username already taken")
	}

	exists = 0
	_ = s.pool.QueryRow(ctx, `SELECT 1 FROM users WHERE email = $1`, email).Scan(&exists)
	if exists == 1 {
		return "", "", domain.User{}, errors.New("email already registered")
	}

	now := time.Now().UTC()
	var id string
	err := s.pool.QueryRow(ctx,
		`INSERT INTO users (username, email, password_hash, display_name, role, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, 'viewer', $5, $5) RETURNING id::text`,
		username, email, hashPassword(password), displayName, now,
	).Scan(&id)
	if err != nil {
		return "", "", domain.User{}, err
	}

	user := domain.User{
		ID: id, Username: username, Email: email, DisplayName: displayName,
		Role: "viewer", IsActive: true, CreatedAt: now,
	}

	accessToken, err := randomToken()
	if err != nil {
		return "", "", domain.User{}, err
	}
	refreshToken, err := randomToken()
	if err != nil {
		return "", "", domain.User{}, err
	}

	if err := s.writeSession(ctx, accessToken, refreshToken, user, ip, userAgent, now); err != nil {
		return "", "", domain.User{}, err
	}

	return accessToken, refreshToken, publicUser(user), nil
}

// writeSession stores session data in Redis and adds the token to the user's session set.
func (s *Store) writeSession(ctx context.Context, accessToken, refreshToken string, user domain.User, ip, userAgent string, now time.Time) error {
	expiresAt := now.Add(sessionTTL)
	pipe := s.redis.Pipeline()

	pipe.HSet(ctx, sessKey(accessToken), map[string]interface{}{
		"user_id":           user.ID,
		"user_email":        user.Email,
		"user_display_name": user.DisplayName,
		"user_role":         user.Role,
		"refresh_token":     refreshToken,
		"ip_address":        ip,
		"user_agent":        userAgent,
		"created_at":        now.Format(time.RFC3339),
		"last_active_at":    now.Format(time.RFC3339),
		"expires_at":        expiresAt.Format(time.RFC3339),
	})
	pipe.Expire(ctx, sessKey(accessToken), sessionTTL)
	pipe.Set(ctx, refreshKey(refreshToken), accessToken, sessionTTL)
	pipe.SAdd(ctx, userSessKey(user.ID), accessToken)

	_, err := pipe.Exec(ctx)
	return err
}

func (s *Store) ValidateSession(token string) (domain.User, bool) {
	ctx := context.Background()

	data, err := s.redis.HGetAll(ctx, sessKey(token)).Result()
	if err != nil || len(data) == 0 {
		return domain.User{}, false
	}

	// Check expiry (Redis TTL handles it, but guard against clock skew)
	expiresAt, _ := time.Parse(time.RFC3339, data["expires_at"])
	if time.Now().UTC().After(expiresAt) {
		return domain.User{}, false
	}

	// Update last_active_at asynchronously
	go s.redis.HSet(context.Background(), sessKey(token), "last_active_at", time.Now().UTC().Format(time.RFC3339))

	// Always fetch user from PG so role/isActive changes are reflected immediately
	userID := data["user_id"]
	var user domain.User
	err = s.pool.QueryRow(ctx,
		`SELECT id::text, username, email, display_name, role, is_active, created_at FROM users WHERE id = $1 AND is_active = true`,
		userID,
	).Scan(&user.ID, &user.Username, &user.Email, &user.DisplayName, &user.Role, &user.IsActive, &user.CreatedAt)
	if err != nil {
		return domain.User{}, false
	}
	return publicUser(user), true
}

func (s *Store) DeleteSession(token string) {
	ctx := context.Background()
	// Get refresh token before deleting
	rt, _ := s.redis.HGet(ctx, sessKey(token), "refresh_token").Result()
	userID, _ := s.redis.HGet(ctx, sessKey(token), "user_id").Result()

	pipe := s.redis.Pipeline()
	pipe.Del(ctx, sessKey(token))
	if rt != "" {
		pipe.Del(ctx, refreshKey(rt))
	}
	if userID != "" {
		pipe.SRem(ctx, userSessKey(userID), token)
	}
	_, _ = pipe.Exec(ctx)
}

func (s *Store) RefreshSession(refreshToken, ip, userAgent string) (string, string, domain.User, error) {
	ctx := context.Background()

	// Look up access token from refresh token
	oldAccessToken, err := s.redis.Get(ctx, refreshKey(refreshToken)).Result()
	if err != nil {
		return "", "", domain.User{}, errors.New("invalid refresh token")
	}

	// Get session data
	data, err := s.redis.HGetAll(ctx, sessKey(oldAccessToken)).Result()
	if err != nil || len(data) == 0 {
		return "", "", domain.User{}, errors.New("invalid refresh token")
	}

	expiresAt, _ := time.Parse(time.RFC3339, data["expires_at"])
	if time.Now().UTC().After(expiresAt) {
		return "", "", domain.User{}, errors.New("refresh token expired")
	}

	userID := data["user_id"]
	var user domain.User
	err = s.pool.QueryRow(ctx,
		`SELECT id::text, username, email, display_name, role, is_active, created_at FROM users WHERE id = $1 AND is_active = true`,
		userID,
	).Scan(&user.ID, &user.Username, &user.Email, &user.DisplayName, &user.Role, &user.IsActive, &user.CreatedAt)
	if err != nil {
		return "", "", domain.User{}, errors.New("user not found")
	}

	newAccessToken, err := randomToken()
	if err != nil {
		return "", "", domain.User{}, err
	}
	newRefreshToken, err := randomToken()
	if err != nil {
		return "", "", domain.User{}, err
	}

	now := time.Now().UTC()

	// Delete old session
	pipe := s.redis.Pipeline()
	pipe.Del(ctx, sessKey(oldAccessToken))
	pipe.Del(ctx, refreshKey(refreshToken))
	pipe.SRem(ctx, userSessKey(userID), oldAccessToken)
	if _, err := pipe.Exec(ctx); err != nil {
		return "", "", domain.User{}, err
	}

	// Write new session
	if err := s.writeSession(ctx, newAccessToken, newRefreshToken, user, ip, userAgent, now); err != nil {
		return "", "", domain.User{}, err
	}

	return newAccessToken, newRefreshToken, publicUser(user), nil
}

func (s *Store) ListUserSessions(userID string) ([]domain.Session, error) {
	ctx := context.Background()

	tokens, err := s.redis.SMembers(ctx, userSessKey(userID)).Result()
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	var sessions []domain.Session
	for _, token := range tokens {
		data, err := s.redis.HGetAll(ctx, sessKey(token)).Result()
		if err != nil || len(data) == 0 {
			// Token expired/missing — clean up the set entry
			s.redis.SRem(ctx, userSessKey(userID), token)
			continue
		}
		expiresAt, _ := time.Parse(time.RFC3339, data["expires_at"])
		if now.After(expiresAt) {
			s.redis.SRem(ctx, userSessKey(userID), token)
			continue
		}
		createdAt, _ := time.Parse(time.RFC3339, data["created_at"])
		lastActiveAt, _ := time.Parse(time.RFC3339, data["last_active_at"])
		sessions = append(sessions, domain.Session{
			AccessToken:     token,
			UserID:          userID,
			UserEmail:       data["user_email"],
			UserDisplayName: data["user_display_name"],
			UserRole:        data["user_role"],
			IPAddress:       data["ip_address"],
			UserAgent:       data["user_agent"],
			CreatedAt:       createdAt,
			LastActiveAt:    lastActiveAt,
			ExpiresAt:       expiresAt,
		})
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].LastActiveAt.After(sessions[j].LastActiveAt)
	})
	return sessions, nil
}

func (s *Store) RevokeUserSessions(userID string, exceptToken string) (int, error) {
	ctx := context.Background()

	tokens, err := s.redis.SMembers(ctx, userSessKey(userID)).Result()
	if err != nil {
		return 0, err
	}

	count := 0
	for _, token := range tokens {
		if token == exceptToken {
			continue
		}
		rt, _ := s.redis.HGet(ctx, sessKey(token), "refresh_token").Result()
		pipe := s.redis.Pipeline()
		pipe.Del(ctx, sessKey(token))
		if rt != "" {
			pipe.Del(ctx, refreshKey(rt))
		}
		pipe.SRem(ctx, userSessKey(userID), token)
		if _, err := pipe.Exec(ctx); err == nil {
			count++
		}
	}
	return count, nil
}

// ── Admin: session management ──

func (s *Store) ListSessions(page, pageSize int) ([]domain.Session, int) {
	ctx := context.Background()
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// SCAN all session keys
	var keys []string
	var cursor uint64
	for {
		batch, nextCursor, err := s.redis.Scan(ctx, cursor, "sess:*", 100).Result()
		if err != nil {
			break
		}
		keys = append(keys, batch...)
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	now := time.Now().UTC()
	var sessions []domain.Session
	for _, key := range keys {
		data, err := s.redis.HGetAll(ctx, key).Result()
		if err != nil || len(data) == 0 {
			continue
		}
		expiresAt, _ := time.Parse(time.RFC3339, data["expires_at"])
		if now.After(expiresAt) {
			continue // skip expired (TTL may not have fired yet)
		}
		createdAt, _ := time.Parse(time.RFC3339, data["created_at"])
		lastActiveAt, _ := time.Parse(time.RFC3339, data["last_active_at"])
		accessToken := strings.TrimPrefix(key, "sess:")

		sessions = append(sessions, domain.Session{
			AccessToken:     accessToken,
			UserID:          data["user_id"],
			UserEmail:       data["user_email"],
			UserDisplayName: data["user_display_name"],
			UserRole:        data["user_role"],
			IPAddress:       data["ip_address"],
			UserAgent:       data["user_agent"],
			CreatedAt:       createdAt,
			LastActiveAt:    lastActiveAt,
			ExpiresAt:       expiresAt,
		})
	}

	// Sort by LastActiveAt descending
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].LastActiveAt.After(sessions[j].LastActiveAt)
	})

	total := len(sessions)
	start := (page - 1) * pageSize
	if start >= total {
		return []domain.Session{}, total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return sessions[start:end], total
}

func (s *Store) RevokeSession(accessToken string) error {
	ctx := context.Background()

	data, err := s.redis.HGetAll(ctx, sessKey(accessToken)).Result()
	if err != nil || len(data) == 0 {
		return errors.New("session not found")
	}

	rt := data["refresh_token"]
	userID := data["user_id"]

	pipe := s.redis.Pipeline()
	pipe.Del(ctx, sessKey(accessToken))
	if rt != "" {
		pipe.Del(ctx, refreshKey(rt))
	}
	if userID != "" {
		pipe.SRem(ctx, userSessKey(userID), accessToken)
	}
	_, err = pipe.Exec(ctx)
	return err
}

func (s *Store) UpdateSessionLastActive(token, ip, userAgent string) {
	// No-op: ValidateSession now updates last_active_at asynchronously.
}

func (s *Store) FindUserByEmail(email string) (domain.User, bool) {
	ctx := context.Background()
	var user domain.User
	err := s.pool.QueryRow(ctx,
		`SELECT id::text, username, email, display_name, role, is_active, created_at FROM users WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.Username, &user.Email, &user.DisplayName, &user.Role, &user.IsActive, &user.CreatedAt)
	if err != nil {
		return domain.User{}, false
	}
	return publicUser(user), true
}

func (s *Store) UpdateUserProfile(userID string, displayName string) (domain.User, error) {
	ctx := context.Background()
	var user domain.User
	err := s.pool.QueryRow(ctx,
		`UPDATE users SET display_name = $1, updated_at = now() WHERE id = $2 AND is_active = true
		 RETURNING id::text, username, email, display_name, role, is_active, created_at`,
		displayName, userID,
	).Scan(&user.ID, &user.Username, &user.Email, &user.DisplayName, &user.Role, &user.IsActive, &user.CreatedAt)
	if err != nil {
		return domain.User{}, errors.New("user not found")
	}
	return publicUser(user), nil
}

func (s *Store) UpdateUserPassword(userID, currentPassword, newPassword string) error {
	ctx := context.Background()
	var hash string
	err := s.pool.QueryRow(ctx,
		`SELECT password_hash FROM users WHERE id = $1 AND is_active = true`,
		userID,
	).Scan(&hash)
	if err != nil {
		return errors.New("user not found")
	}
	if hash != hashPassword(currentPassword) {
		return errors.New("current password is incorrect")
	}
	if hashPassword(newPassword) == hash {
		return errors.New("new password must differ from current password")
	}
	_, err = s.pool.Exec(ctx,
		`UPDATE users SET password_hash = $1, updated_at = now() WHERE id = $2`,
		hashPassword(newPassword), userID,
	)
	if err != nil {
		return errors.New("failed to update password")
	}
	return nil
}

func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

func randomToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func publicUser(user domain.User) domain.User {
	user.PasswordHash = ""
	return user
}

func dedupeStrings(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	out := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		key := strings.ToLower(item)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, item)
	}
	return out
}

func slugify(input string) string {
	input = strings.ToLower(strings.TrimSpace(input))
	input = strings.ReplaceAll(input, " ", "-")
	input = strings.ReplaceAll(input, "/", "-")
	input = strings.ReplaceAll(input, "_", "-")
	return input
}

// ── Admin: user management ──

func (s *Store) ListUsers(page, pageSize int) ([]domain.User, int) {
	ctx := context.Background()
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	var total int
	_ = s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&total)

	offset := (page - 1) * pageSize
	rows, err := s.pool.Query(ctx,
		`SELECT id::text, username, email, display_name, role, is_active, last_login_at, created_at FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		pageSize, offset,
	)
	if err != nil {
		return nil, total
	}
	defer rows.Close()

	var out []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.DisplayName, &u.Role, &u.IsActive, &u.LastLoginAt, &u.CreatedAt); err != nil {
			continue
		}
		out = append(out, publicUser(u))
	}
	return out, total
}

func (s *Store) UpdateUserRole(userID, role string) (domain.User, error) {
	ctx := context.Background()
	var user domain.User
	err := s.pool.QueryRow(ctx,
		`UPDATE users SET role = $1, updated_at = now() WHERE id = $2
		 RETURNING id::text, username, email, display_name, role, is_active, created_at`,
		role, userID,
	).Scan(&user.ID, &user.Username, &user.Email, &user.DisplayName, &user.Role, &user.IsActive, &user.CreatedAt)
	if err != nil {
		return domain.User{}, errors.New("user not found")
	}
	return publicUser(user), nil
}

func (s *Store) SetUserActive(userID string, active bool) error {
	ctx := context.Background()
	tag, err := s.pool.Exec(ctx,
		`UPDATE users SET is_active = $1, updated_at = now() WHERE id = $2`,
		active, userID,
	)
	if err != nil || tag.RowsAffected() == 0 {
		return errors.New("user not found")
	}
	if !active {
		_, _ = s.RevokeUserSessions(userID, "")
	}
	return nil
}

// ── Admin: submission management ──

func (s *Store) ListSubmissions(status string, page, pageSize int) ([]domain.Submission, int) {
	ctx := context.Background()
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	where := "1=1"
	args := []any{}
	if status != "" {
		where = "status = $1"
		args = append(args, status)
	}

	var total int
	_ = s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM tool_submissions WHERE `+where, args...).Scan(&total)

	offset := (page - 1) * pageSize
	argIdx := len(args) + 1
	query := fmt.Sprintf(`SELECT id::text, type, COALESCE(submitted_by::text, ''), COALESCE(tool_id::text, ''), COALESCE(submitter_email, ''), payload, status, COALESCE(reviewer_id::text, ''), COALESCE(review_note, ''), created_at, reviewed_at FROM tool_submissions WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)
	args = append(args, pageSize, offset)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, total
	}
	defer rows.Close()

	var out []domain.Submission
	for rows.Next() {
		var sub domain.Submission
		if err := rows.Scan(&sub.ID, &sub.Type, &sub.SubmittedBy, &sub.ToolID, &sub.SubmitterEmail, &sub.Payload, &sub.Status, &sub.ReviewerID, &sub.ReviewNote, &sub.CreatedAt, &sub.ReviewedAt); err != nil {
			continue
		}
		out = append(out, sub)
	}
	return out, total
}

func (s *Store) ReviewSubmission(id, reviewerID, status, note string) (domain.Submission, error) {
	ctx := context.Background()
	now := time.Now().UTC()
	var sub domain.Submission
	err := s.pool.QueryRow(ctx,
		`UPDATE tool_submissions SET status = $1, reviewer_id = $2, review_note = $3, reviewed_at = $4
		 WHERE id = $5
		 RETURNING id::text, type, COALESCE(submitted_by::text, ''), COALESCE(tool_id::text, ''), COALESCE(submitter_email, ''), payload, status, COALESCE(reviewer_id::text, ''), COALESCE(review_note, ''), created_at, reviewed_at`,
		status, reviewerID, note, now, id,
	).Scan(&sub.ID, &sub.Type, &sub.SubmittedBy, &sub.ToolID, &sub.SubmitterEmail, &sub.Payload, &sub.Status, &sub.ReviewerID, &sub.ReviewNote, &sub.CreatedAt, &sub.ReviewedAt)
	if err != nil {
		return domain.Submission{}, errors.New("submission not found")
	}
	return sub, nil
}

// ── Admin: dashboard stats ──

func (s *Store) AdminStats() map[string]int {
	ctx := context.Background()
	var toolCount, publishedToolCount, draftToolCount, categoryCount, userCount, activeUserCount, pendingSubmissionCount, submissionCount int

	_ = s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM tools`).Scan(&toolCount)
	_ = s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM tools WHERE status = 'published'`).Scan(&publishedToolCount)
	_ = s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM tools WHERE status = 'draft'`).Scan(&draftToolCount)
	_ = s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM categories`).Scan(&categoryCount)
	_ = s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&userCount)
	_ = s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE is_active = true`).Scan(&activeUserCount)
	_ = s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM tool_submissions WHERE status = 'pending'`).Scan(&pendingSubmissionCount)
	_ = s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM tool_submissions`).Scan(&submissionCount)

	return map[string]int{
		"toolCount":              toolCount,
		"publishedToolCount":     publishedToolCount,
		"draftToolCount":         draftToolCount,
		"categoryCount":          categoryCount,
		"userCount":              userCount,
		"activeUserCount":        activeUserCount,
		"pendingSubmissionCount": pendingSubmissionCount,
		"submissionCount":        submissionCount,
	}
}

// ── Admin: audit logs ──

func (s *Store) ListAuditLogs(page, pageSize int) ([]domain.AuditLog, int) {
	ctx := context.Background()
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	var total int
	_ = s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM audit_logs`).Scan(&total)

	offset := (page - 1) * pageSize
	rows, err := s.pool.Query(ctx,
		`SELECT id::text, COALESCE(user_id::text, ''), action, resource_type, COALESCE(resource_id::text, ''), before_data, after_data, created_at FROM audit_logs ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		pageSize, offset,
	)
	if err != nil {
		return nil, total
	}
	defer rows.Close()

	var out []domain.AuditLog
	for rows.Next() {
		var log domain.AuditLog
		if err := rows.Scan(&log.ID, &log.UserID, &log.Action, &log.ResourceType, &log.ResourceID, &log.BeforeData, &log.AfterData, &log.CreatedAt); err != nil {
			continue
		}
		out = append(out, log)
	}
	return out, total
}

func (s *Store) CreateAuditLog(log domain.AuditLog) {
	ctx := context.Background()
	_, _ = s.pool.Exec(ctx,
		`INSERT INTO audit_logs (user_id, action, resource_type, resource_id, before_data, after_data, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		nilIfEmpty(log.UserID), log.Action, log.ResourceType, nilIfEmpty(log.ResourceID), log.BeforeData, log.AfterData, time.Now().UTC(),
	)
}

func nilIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

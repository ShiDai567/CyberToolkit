package postgres

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"cybertoolkit/backend/internal/domain"
)

type Store struct {
	pool             *pgxpool.Pool
	adminEmail       string
	adminPassword    string
	adminDisplayName string
}

func NewStore(databaseURL, adminEmail, adminPassword, adminDisplayName string) (*Store, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	s := &Store{
		pool:             pool,
		adminEmail:       adminEmail,
		adminPassword:    adminPassword,
		adminDisplayName: adminDisplayName,
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
// the configured password and admin role on every startup. This avoids the
// previous behaviour where the admin was only created during the initial
// empty-database seed and could not be changed by editing .env later.
func (s *Store) ensureAdmin(ctx context.Context) error {
	email := strings.TrimSpace(s.adminEmail)
	password := s.adminPassword
	if email == "" || password == "" {
		return nil
	}

	displayName := strings.TrimSpace(s.adminDisplayName)
	if displayName == "" {
		displayName = "Admin"
	}

	now := time.Now().UTC()
	tag, err := s.pool.Exec(ctx,
		`UPDATE users SET password_hash = $1, display_name = $2, role = 'admin', is_active = true, updated_at = $3
		 WHERE email = $4`,
		hashPassword(password), displayName, now, email,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() > 0 {
		return nil
	}

	_, err = s.pool.Exec(ctx,
		`INSERT INTO users (email, password_hash, display_name, role, created_at, updated_at)
		 VALUES ($1, $2, $3, 'admin', $4, $4)`,
		email, hashPassword(password), displayName, now,
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
		Email       string
		Password    string
		DisplayName string
		Role        string
	}{
		{Email: "editor@cybertoolkit.local", Password: "editor123456", DisplayName: "Editor Demo", Role: "editor"},
		{Email: "viewer@cybertoolkit.local", Password: "viewer123456", DisplayName: "Viewer Demo", Role: "viewer"},
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
			`INSERT INTO users (email, password_hash, display_name, role, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $5)`,
			u.Email, hashPassword(u.Password), u.DisplayName, u.Role, now,
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
	_ = s.pool.QueryRow(ctx,
		`INSERT INTO tool_submissions (type, submitted_by, tool_id, submitter_email, payload, status, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id::text`,
		submission.Type, submission.SubmittedBy, submission.ToolID, submission.SubmitterEmail, submission.Payload, submission.Status, now,
	).Scan(&id)
	submission.ID = id
	submission.CreatedAt = now
	return submission
}

func (s *Store) Authenticate(email, password string) (string, string, domain.User, error) {
	ctx := context.Background()
	var user domain.User
	err := s.pool.QueryRow(ctx,
		`SELECT id::text, email, display_name, password_hash, role, is_active, last_login_at, created_at FROM users WHERE email = $1 AND is_active = true`,
		email,
	).Scan(&user.ID, &user.Email, &user.DisplayName, &user.PasswordHash, &user.Role, &user.IsActive, &user.LastLoginAt, &user.CreatedAt)
	if err != nil {
		return "", "", domain.User{}, errors.New("invalid email or password")
	}
	if user.PasswordHash != hashPassword(password) {
		return "", "", domain.User{}, errors.New("invalid email or password")
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
	public := publicUser(user)
	expiresAt := now.Add(24 * time.Hour)
	_, _ = s.pool.Exec(ctx,
		`UPDATE users SET last_login_at = $1 WHERE id = $2`, now, user.ID,
	)
	_, _ = s.pool.Exec(ctx,
		`INSERT INTO sessions (access_token, user_id, refresh_token, expires_at) VALUES ($1, $2, $3, $4)`,
		accessToken, user.ID, refreshToken, expiresAt,
	)
	return accessToken, refreshToken, public, nil
}

func (s *Store) Register(email, password, displayName string) (string, string, domain.User, error) {
	ctx := context.Background()

	var exists int
	_ = s.pool.QueryRow(ctx, `SELECT 1 FROM users WHERE email = $1`, email).Scan(&exists)
	if exists == 1 {
		return "", "", domain.User{}, errors.New("email already registered")
	}

	now := time.Now().UTC()
	var id string
	err := s.pool.QueryRow(ctx,
		`INSERT INTO users (email, password_hash, display_name, role, created_at, updated_at)
		 VALUES ($1, $2, $3, 'viewer', $4, $4) RETURNING id::text`,
		email, hashPassword(password), displayName, now,
	).Scan(&id)
	if err != nil {
		return "", "", domain.User{}, err
	}

	user := domain.User{
		ID:          id,
		Email:       email,
		DisplayName: displayName,
		Role:        "viewer",
		IsActive:    true,
		CreatedAt:   now,
	}

	accessToken, err := randomToken()
	if err != nil {
		return "", "", domain.User{}, err
	}
	refreshToken, err := randomToken()
	if err != nil {
		return "", "", domain.User{}, err
	}

	public := publicUser(user)
	expiresAt := now.Add(24 * time.Hour)
	_, _ = s.pool.Exec(ctx,
		`INSERT INTO sessions (access_token, user_id, refresh_token, expires_at) VALUES ($1, $2, $3, $4)`,
		accessToken, user.ID, refreshToken, expiresAt,
	)
	return accessToken, refreshToken, public, nil
}

func (s *Store) ValidateSession(token string) (domain.User, bool) {
	ctx := context.Background()
	// Clean up expired sessions
	_, _ = s.pool.Exec(ctx, `DELETE FROM sessions WHERE expires_at < now()`)

	var user domain.User
	err := s.pool.QueryRow(ctx,
		`SELECT u.id::text, u.email, u.display_name, u.role, u.is_active, u.created_at FROM sessions s JOIN users u ON s.user_id = u.id WHERE s.access_token = $1 AND s.expires_at > now()`,
		token,
	).Scan(&user.ID, &user.Email, &user.DisplayName, &user.Role, &user.IsActive, &user.CreatedAt)
	if err != nil {
		return domain.User{}, false
	}
	return publicUser(user), true
}

func (s *Store) DeleteSession(token string) {
	ctx := context.Background()
	_, _ = s.pool.Exec(ctx, `DELETE FROM sessions WHERE access_token = $1`, token)
}

func (s *Store) RefreshSession(refreshToken string) (string, string, domain.User, error) {
	ctx := context.Background()
	// Clean up expired sessions
	_, _ = s.pool.Exec(ctx, `DELETE FROM sessions WHERE expires_at < now()`)

	var user domain.User
	err := s.pool.QueryRow(ctx,
		`SELECT u.id::text, u.email, u.display_name, u.role, u.is_active, u.created_at FROM sessions s JOIN users u ON s.user_id = u.id WHERE s.refresh_token = $1 AND s.expires_at > now()`,
		refreshToken,
	).Scan(&user.ID, &user.Email, &user.DisplayName, &user.Role, &user.IsActive, &user.CreatedAt)
	if err != nil {
		return "", "", domain.User{}, errors.New("invalid refresh token")
	}

	newAccessToken, err := randomToken()
	if err != nil {
		return "", "", domain.User{}, err
	}
	newRefreshToken, err := randomToken()
	if err != nil {
		return "", "", domain.User{}, err
	}

	// Delete old session and create new one in a transaction
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return "", "", domain.User{}, err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `DELETE FROM sessions WHERE refresh_token = $1`, refreshToken)
	if err != nil {
		return "", "", domain.User{}, err
	}

	expiresAt := time.Now().UTC().Add(24 * time.Hour)
	_, err = tx.Exec(ctx,
		`INSERT INTO sessions (access_token, user_id, refresh_token, expires_at) VALUES ($1, $2, $3, $4)`,
		newAccessToken, user.ID, newRefreshToken, expiresAt,
	)
	if err != nil {
		return "", "", domain.User{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", "", domain.User{}, err
	}

	return newAccessToken, newRefreshToken, publicUser(user), nil
}

func (s *Store) FindUserByEmail(email string) (domain.User, bool) {
	ctx := context.Background()
	var user domain.User
	err := s.pool.QueryRow(ctx,
		`SELECT id::text, email, display_name, role, is_active, created_at FROM users WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.Email, &user.DisplayName, &user.Role, &user.IsActive, &user.CreatedAt)
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
		 RETURNING id::text, email, display_name, role, is_active, created_at`,
		displayName, userID,
	).Scan(&user.ID, &user.Email, &user.DisplayName, &user.Role, &user.IsActive, &user.CreatedAt)
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

func (s *Store) RevokeUserSessions(userID string, exceptToken string) (int, error) {
	ctx := context.Background()
	var (
		tag pgconn.CommandTag
		err error
	)
	if exceptToken == "" {
		tag, err = s.pool.Exec(ctx,
			`DELETE FROM sessions WHERE user_id = $1`, userID,
		)
	} else {
		tag, err = s.pool.Exec(ctx,
			`DELETE FROM sessions WHERE user_id = $1 AND access_token <> $2`, userID, exceptToken,
		)
	}
	if err != nil {
		return 0, err
	}
	return int(tag.RowsAffected()), nil
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
		`SELECT id::text, email, display_name, role, is_active, last_login_at, created_at FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		pageSize, offset,
	)
	if err != nil {
		return nil, total
	}
	defer rows.Close()

	var out []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Email, &u.DisplayName, &u.Role, &u.IsActive, &u.LastLoginAt, &u.CreatedAt); err != nil {
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
		 RETURNING id::text, email, display_name, role, is_active, created_at`,
		role, userID,
	).Scan(&user.ID, &user.Email, &user.DisplayName, &user.Role, &user.IsActive, &user.CreatedAt)
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
		_, _ = s.pool.Exec(ctx, `DELETE FROM sessions WHERE user_id = $1`, userID)
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

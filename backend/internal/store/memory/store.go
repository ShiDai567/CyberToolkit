package memory

import (
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	"cybertoolkit/backend/internal/domain"
)

type Store struct {
	mu           sync.RWMutex
	categories   []domain.Category
	tags         []domain.Tag
	tools        []domain.Tool
	toolTags     []domain.ToolTag
	submissions  []domain.Submission
	adminUser    domain.AdminUser
	nextCategory int
	nextTag      int
	nextTool     int
	nextSubmit   int
}

func NewStore() *Store {
	now := time.Now().UTC()
	publishedAt := now.Add(-24 * time.Hour)

	categories := []domain.Category{
		{ID: "cat_1", Slug: "network-scanning", Name: "Network Scanning", Description: "Discover hosts, ports, services and exposed assets.", Icon: "Radar", SortOrder: 1, IsVisible: true, CreatedAt: now, UpdatedAt: now},
		{ID: "cat_2", Slug: "vulnerability-assessment", Name: "Vulnerability Assessment", Description: "Identify vulnerabilities, weak configurations and risks.", Icon: "ShieldAlert", SortOrder: 2, IsVisible: true, CreatedAt: now, UpdatedAt: now},
		{ID: "cat_3", Slug: "penetration-testing", Name: "Penetration Testing", Description: "Validate attack paths and defensive controls.", Icon: "Crosshair", SortOrder: 3, IsVisible: true, CreatedAt: now, UpdatedAt: now},
		{ID: "cat_4", Slug: "osint", Name: "OSINT", Description: "Collect intelligence from public data sources.", Icon: "Eye", SortOrder: 4, IsVisible: true, CreatedAt: now, UpdatedAt: now},
	}

	tags := []domain.Tag{
		{ID: "tag_1", Slug: "port-scan", Name: "Port Scan", CreatedAt: now},
		{ID: "tag_2", Slug: "host-discovery", Name: "Host Discovery", CreatedAt: now},
		{ID: "tag_3", Slug: "automation", Name: "Automation", CreatedAt: now},
		{ID: "tag_4", Slug: "web-security", Name: "Web Security", CreatedAt: now},
		{ID: "tag_5", Slug: "asset-discovery", Name: "Asset Discovery", CreatedAt: now},
	}

	tools := []domain.Tool{
		{ID: "tool_1", Slug: "nmap", Name: "Nmap", ShortDescription: "Industry-standard network discovery and audit tool.", LongDescription: "Nmap supports host discovery, port scanning, service identification and scriptable security checks.", CategoryID: "cat_1", Difficulty: domain.DifficultyIntermediate, Icon: "Radar", Featured: true, Status: domain.ToolStatusPublished, WebsiteURL: "https://nmap.org", GitHubURL: "https://github.com/nmap/nmap", ViewCount: 1200, FavoriteCount: 230, PublishedAt: &publishedAt, CreatedAt: now, UpdatedAt: now},
		{ID: "tool_2", Slug: "nuclei", Name: "Nuclei", ShortDescription: "Fast template-driven vulnerability scanner.", LongDescription: "Nuclei uses YAML templates to scan for CVEs, exposed panels, misconfigurations and fingerprints.", CategoryID: "cat_2", Difficulty: domain.DifficultyIntermediate, Icon: "Atom", Featured: true, Status: domain.ToolStatusPublished, WebsiteURL: "https://nuclei.projectdiscovery.io", GitHubURL: "https://github.com/projectdiscovery/nuclei", ViewCount: 980, FavoriteCount: 180, PublishedAt: &publishedAt, CreatedAt: now, UpdatedAt: now},
		{ID: "tool_3", Slug: "burpsuite", Name: "Burp Suite", ShortDescription: "Core platform for web application security testing.", LongDescription: "Burp Suite combines proxying, replay, scanning and traffic analysis for web testing workflows.", CategoryID: "cat_3", Difficulty: domain.DifficultyIntermediate, Icon: "Bug", Featured: true, Status: domain.ToolStatusPublished, WebsiteURL: "https://portswigger.net/burp", ViewCount: 870, FavoriteCount: 150, PublishedAt: &publishedAt, CreatedAt: now, UpdatedAt: now},
		{ID: "tool_4", Slug: "shodan", Name: "Shodan", ShortDescription: "Search engine for exposed devices and internet-facing services.", LongDescription: "Shodan helps discover exposed services, fingerprints and assets visible on the public internet.", CategoryID: "cat_4", Difficulty: domain.DifficultyBeginner, Icon: "Eye", Featured: true, Status: domain.ToolStatusPublished, WebsiteURL: "https://www.shodan.io", ViewCount: 1100, FavoriteCount: 210, PublishedAt: &publishedAt, CreatedAt: now, UpdatedAt: now},
	}

	toolTags := []domain.ToolTag{
		{ToolID: "tool_1", TagID: "tag_1"},
		{ToolID: "tool_1", TagID: "tag_2"},
		{ToolID: "tool_2", TagID: "tag_3"},
		{ToolID: "tool_2", TagID: "tag_5"},
		{ToolID: "tool_3", TagID: "tag_4"},
		{ToolID: "tool_4", TagID: "tag_5"},
	}

	return &Store{
		categories:  categories,
		tags:        tags,
		tools:       tools,
		toolTags:    toolTags,
		submissions: []domain.Submission{},
		adminUser: domain.AdminUser{
			ID:          "user_admin",
			Email:       "admin@cybertoolkit.local",
			DisplayName: "Admin",
			Role:        "admin",
		},
		nextCategory: len(categories) + 1,
		nextTag:      len(tags) + 1,
		nextTool:     len(tools) + 1,
		nextSubmit:   1,
	}
}

func (s *Store) Stats() map[string]int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	featured := 0
	for _, tool := range s.tools {
		if tool.Status == domain.ToolStatusPublished && tool.Featured {
			featured++
		}
	}

	return map[string]int{
		"toolCount":     len(s.tools),
		"categoryCount": len(s.categories),
		"featuredCount": featured,
	}
}

func (s *Store) ListCategories(visibleOnly bool) []domain.Category {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]domain.Category, 0, len(s.categories))
	for _, category := range s.categories {
		if visibleOnly && !category.IsVisible {
			continue
		}
		out = append(out, category)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].SortOrder < out[j].SortOrder })
	return out
}

func (s *Store) CategoryByID(id string) (domain.Category, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, category := range s.categories {
		if category.ID == id {
			return category, true
		}
	}
	return domain.Category{}, false
}

func (s *Store) CreateCategory(input domain.Category) domain.Category {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	input.ID = "cat_" + itoa(s.nextCategory)
	input.CreatedAt = now
	input.UpdatedAt = now
	s.nextCategory++
	s.categories = append(s.categories, input)
	return input
}

func (s *Store) UpdateCategory(id string, update domain.Category) (domain.Category, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.categories {
		if s.categories[i].ID != id {
			continue
		}
		update.ID = s.categories[i].ID
		update.CreatedAt = s.categories[i].CreatedAt
		update.UpdatedAt = time.Now().UTC()
		s.categories[i] = update
		return update, nil
	}
	return domain.Category{}, errors.New("category not found")
}

func (s *Store) ListTags() []domain.Tag {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]domain.Tag, len(s.tags))
	copy(out, s.tags)
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (s *Store) ListTools(filters domain.ToolFilters, admin bool) ([]domain.Tool, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

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

	filtered := make([]domain.Tool, 0, len(s.tools))
	for _, tool := range s.tools {
		if !admin && tool.Status != domain.ToolStatusPublished {
			continue
		}
		if admin && filters.Status != "" && string(tool.Status) != filters.Status {
			continue
		}
		if filters.Query != "" && !containsTool(tool, filters.Query) {
			continue
		}
		if filters.Category != "" && !s.matchCategory(tool.CategoryID, filters.Category) {
			continue
		}
		if filters.Difficulty != "" && string(tool.Difficulty) != filters.Difficulty {
			continue
		}
		if filters.Tag != "" && !s.matchTag(tool.ID, filters.Tag) {
			continue
		}
		if filters.Featured != nil && tool.Featured != *filters.Featured {
			continue
		}
		filtered = append(filtered, tool)
	}

	switch filters.Sort {
	case "name":
		sort.Slice(filtered, func(i, j int) bool { return filtered[i].Name < filtered[j].Name })
	case "popular":
		sort.Slice(filtered, func(i, j int) bool { return filtered[i].ViewCount > filtered[j].ViewCount })
	default:
		sort.Slice(filtered, func(i, j int) bool { return filtered[i].CreatedAt.After(filtered[j].CreatedAt) })
	}

	total := len(filtered)
	start := (page - 1) * pageSize
	if start >= total {
		return []domain.Tool{}, total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	out := make([]domain.Tool, end-start)
	copy(out, filtered[start:end])
	return out, total
}

func (s *Store) GetToolBySlug(slug string, admin bool) (domain.Tool, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, tool := range s.tools {
		if tool.Slug != slug {
			continue
		}
		if !admin && tool.Status != domain.ToolStatusPublished {
			return domain.Tool{}, false
		}
		return tool, true
	}
	return domain.Tool{}, false
}

func (s *Store) GetToolByID(id string) (domain.Tool, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, tool := range s.tools {
		if tool.ID == id {
			return tool, true
		}
	}
	return domain.Tool{}, false
}

func (s *Store) CreateTool(input domain.Tool) domain.Tool {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	input.ID = "tool_" + itoa(s.nextTool)
	input.CreatedAt = now
	input.UpdatedAt = now
	if input.Status == domain.ToolStatusPublished {
		input.PublishedAt = &now
	}
	s.nextTool++
	s.tools = append(s.tools, input)
	return input
}

func (s *Store) UpdateTool(id string, update domain.Tool) (domain.Tool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.tools {
		if s.tools[i].ID != id {
			continue
		}
		update.ID = s.tools[i].ID
		update.CreatedAt = s.tools[i].CreatedAt
		update.UpdatedAt = time.Now().UTC()
		if update.Status == domain.ToolStatusPublished && update.PublishedAt == nil {
			now := time.Now().UTC()
			update.PublishedAt = &now
		}
		s.tools[i] = update
		return update, nil
	}
	return domain.Tool{}, errors.New("tool not found")
}

func (s *Store) ArchiveTool(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.tools {
		if s.tools[i].ID == id {
			s.tools[i].Status = domain.ToolStatusArchived
			s.tools[i].UpdatedAt = time.Now().UTC()
			return nil
		}
	}
	return errors.New("tool not found")
}

func (s *Store) TagsForTool(toolID string) []domain.Tag {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tagIDs := make(map[string]struct{})
	for _, tt := range s.toolTags {
		if tt.ToolID == toolID {
			tagIDs[tt.TagID] = struct{}{}
		}
	}
	out := make([]domain.Tag, 0, len(tagIDs))
	for _, tag := range s.tags {
		if _, ok := tagIDs[tag.ID]; ok {
			out = append(out, tag)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (s *Store) RelatedTools(categoryID, exceptToolID string, limit int) []domain.Tool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]domain.Tool, 0, limit)
	for _, tool := range s.tools {
		if tool.Status != domain.ToolStatusPublished || tool.CategoryID != categoryID || tool.ID == exceptToolID {
			continue
		}
		out = append(out, tool)
		if len(out) == limit {
			break
		}
	}
	return out
}

func (s *Store) ReplaceToolTags(toolID string, tagNames []string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	kept := s.toolTags[:0]
	for _, item := range s.toolTags {
		if item.ToolID != toolID {
			kept = append(kept, item)
		}
	}
	s.toolTags = kept

	for _, name := range dedupeStrings(tagNames) {
		tag := s.ensureTagLocked(name)
		s.toolTags = append(s.toolTags, domain.ToolTag{ToolID: toolID, TagID: tag.ID})
	}
}

func (s *Store) CreateSubmission(submission domain.Submission) domain.Submission {
	s.mu.Lock()
	defer s.mu.Unlock()

	submission.ID = "sub_" + itoa(s.nextSubmit)
	submission.CreatedAt = time.Now().UTC()
	s.nextSubmit++
	s.submissions = append(s.submissions, submission)
	return submission
}

func (s *Store) AdminUser() domain.AdminUser {
	return s.adminUser
}

func (s *Store) matchCategory(categoryID, categorySlug string) bool {
	for _, category := range s.categories {
		if category.ID == categoryID && category.Slug == categorySlug {
			return true
		}
	}
	return false
}

func (s *Store) matchTag(toolID, tagSlug string) bool {
	for _, tt := range s.toolTags {
		if tt.ToolID != toolID {
			continue
		}
		for _, tag := range s.tags {
			if tag.ID == tt.TagID && tag.Slug == tagSlug {
				return true
			}
		}
	}
	return false
}

func (s *Store) ensureTagLocked(name string) domain.Tag {
	slug := slugify(name)
	for _, tag := range s.tags {
		if tag.Slug == slug {
			return tag
		}
	}

	tag := domain.Tag{ID: "tag_" + itoa(s.nextTag), Slug: slug, Name: name, CreatedAt: time.Now().UTC()}
	s.nextTag++
	s.tags = append(s.tags, tag)
	return tag
}

func containsTool(tool domain.Tool, query string) bool {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return true
	}

	return strings.Contains(strings.ToLower(tool.Name), q) ||
		strings.Contains(strings.ToLower(tool.ShortDescription), q) ||
		strings.Contains(strings.ToLower(tool.LongDescription), q)
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

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := make([]byte, 0, 12)
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

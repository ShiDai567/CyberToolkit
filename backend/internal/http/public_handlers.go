package http

import (
	"net/http"
	"strconv"
	"strings"

	"cybertoolkit/backend/internal/domain"
)

func (api *API) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"}, nil)
}

func (api *API) handleHome(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", nil)
		return
	}

	categories := api.store.ListCategories(true)
	tools, _ := api.store.ListTools(domain.ToolFilters{Featured: boolPtr(true), Page: 1, PageSize: 6}, false)

	categoryData := make([]map[string]any, 0, len(categories))
	for _, category := range categories {
		categoryData = append(categoryData, map[string]any{
			"id":          category.Slug,
			"name":        category.Name,
			"description": category.Description,
			"icon":        category.Icon,
			"toolCount":   api.categoryToolCount(category.ID),
		})
	}

	featured := make([]map[string]any, 0, len(tools))
	for _, tool := range tools {
		featured = append(featured, api.toToolCard(tool))
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"stats":         api.store.Stats(),
		"featuredTools": featured,
		"categories":    categoryData,
	}, nil)
}

func (api *API) handleCategories(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", nil)
		return
	}

	includeCounts := r.URL.Query().Get("includeCounts") == "true"
	categories := api.store.ListCategories(true)
	data := make([]map[string]any, 0, len(categories))
	for _, category := range categories {
		item := map[string]any{
			"id":          category.Slug,
			"name":        category.Name,
			"description": category.Description,
			"icon":        category.Icon,
		}
		if includeCounts {
			item["toolCount"] = api.categoryToolCount(category.ID)
		}
		data = append(data, item)
	}

	writeJSON(w, http.StatusOK, data, nil)
}

func (api *API) handleTags(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", nil)
		return
	}

	tags := api.store.ListTags()
	data := make([]map[string]any, 0, len(tags))
	for _, tag := range tags {
		data = append(data, map[string]any{
			"id":   tag.Slug,
			"name": tag.Name,
		})
	}
	writeJSON(w, http.StatusOK, data, nil)
}

func (api *API) handleTools(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", nil)
		return
	}

	featuredQuery := r.URL.Query().Get("featured")
	var featured *bool
	if featuredQuery != "" {
		value := featuredQuery == "true"
		featured = &value
	}

	filters := domain.ToolFilters{
		Query:      r.URL.Query().Get("q"),
		Category:   r.URL.Query().Get("category"),
		Difficulty: r.URL.Query().Get("difficulty"),
		Tag:        r.URL.Query().Get("tag"),
		Featured:   featured,
		Page:       parseIntDefault(r.URL.Query().Get("page"), 1),
		PageSize:   parseIntDefault(r.URL.Query().Get("pageSize"), 20),
		Sort:       r.URL.Query().Get("sort"),
	}

	tools, total := api.store.ListTools(filters, false)
	data := make([]map[string]any, 0, len(tools))
	for _, tool := range tools {
		data = append(data, api.toToolCard(tool))
	}

	page := filters.Page
	if page < 1 {
		page = 1
	}
	pageSize := filters.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	totalPages := 0
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}

	writeJSON(w, http.StatusOK, data, meta{
		"page":       page,
		"pageSize":   pageSize,
		"total":      total,
		"totalPages": totalPages,
	})
}

func (api *API) handleToolDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", nil)
		return
	}

	slug := strings.TrimPrefix(r.URL.Path, "/api/v1/tools/")
	if slug == "" {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "tool not found", nil)
		return
	}

	tool, ok := api.store.GetToolBySlug(slug, false)
	if !ok {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "tool not found", nil)
		return
	}

	category, _ := api.store.CategoryByID(tool.CategoryID)
	tags := api.store.TagsForTool(tool.ID)
	related := api.store.RelatedTools(tool.CategoryID, tool.ID, 3)

	tagNames := make([]string, 0, len(tags))
	for _, tag := range tags {
		tagNames = append(tagNames, tag.Name)
	}

	relatedPayload := make([]map[string]any, 0, len(related))
	for _, item := range related {
		relatedPayload = append(relatedPayload, map[string]any{
			"id":   item.Slug,
			"name": item.Name,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"id":              tool.Slug,
		"name":            tool.Name,
		"description":     tool.ShortDescription,
		"longDescription": tool.LongDescription,
		"category": map[string]any{
			"id":          category.Slug,
			"name":        category.Name,
			"description": category.Description,
		},
		"difficulty": tool.Difficulty,
		"icon":       tool.Icon,
		"featured":   tool.Featured,
		"tags":       tagNames,
		"links": []map[string]any{
			{"type": "website", "label": "Official Website", "url": tool.WebsiteURL},
			{"type": "github", "label": "Source Repository", "url": tool.GitHubURL},
		},
		"relatedTools": relatedPayload,
	}, nil)
}

func (api *API) handleSubmissions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", nil)
		return
	}

	var request struct {
		Type           string         `json:"type"`
		SubmitterEmail string         `json:"submitterEmail"`
		Payload        map[string]any `json:"payload"`
	}
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body", nil)
		return
	}
	if request.Type == "" || len(request.Payload) == 0 {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "type and payload are required", nil)
		return
	}

	submission := domain.Submission{
		Type:           request.Type,
		SubmitterEmail: request.SubmitterEmail,
		Payload:        request.Payload,
		Status:         domain.SubmissionStatusPending,
	}

	// If authenticated, associate submission with the user
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		token := strings.TrimPrefix(auth, "Bearer ")
		if user, ok := api.store.ValidateSession(token); ok {
			submission.SubmittedBy = user.ID
			if submission.SubmitterEmail == "" {
				submission.SubmitterEmail = user.Email
			}
		}
	}

	submission = api.store.CreateSubmission(submission)
	if submission.ID == "" {
		writeError(w, http.StatusInternalServerError, "SUBMISSION_FAILED", "failed to create submission", nil)
		return
	}

	writeJSON(w, http.StatusCreated, submission, nil)
}

func (api *API) handleMySubmissions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", nil)
		return
	}

	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing bearer token", nil)
		return
	}
	token := strings.TrimPrefix(auth, "Bearer ")
	user, ok := api.store.ValidateSession(token)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid token", nil)
		return
	}

	submissions, total := api.store.SubmissionsByUser(user.ID)
	data := make([]map[string]any, 0, len(submissions))
	for _, sub := range submissions {
		data = append(data, map[string]any{
			"id":             sub.ID,
			"type":           sub.Type,
			"submitterEmail": sub.SubmitterEmail,
			"payload":        sub.Payload,
			"status":         sub.Status,
			"reviewNote":     sub.ReviewNote,
			"createdAt":      sub.CreatedAt,
			"reviewedAt":     sub.ReviewedAt,
		})
	}

	page := 1
	pageSize := 20
	totalPages := 0
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}

	writeJSON(w, http.StatusOK, data, meta{
		"page":       page,
		"pageSize":   pageSize,
		"total":      total,
		"totalPages": totalPages,
	})
}

func (api *API) toToolCard(tool domain.Tool) map[string]any {
	category, _ := api.store.CategoryByID(tool.CategoryID)
	tags := api.store.TagsForTool(tool.ID)
	tagNames := make([]string, 0, len(tags))
	for _, tag := range tags {
		tagNames = append(tagNames, tag.Name)
	}

	return map[string]any{
		"id":          tool.Slug,
		"name":        tool.Name,
		"description": tool.ShortDescription,
		"category": map[string]any{
			"id":   category.Slug,
			"name": category.Name,
		},
		"difficulty": tool.Difficulty,
		"icon":       tool.Icon,
		"featured":   tool.Featured,
		"tags":       tagNames,
		"links": map[string]any{
			"website": tool.WebsiteURL,
			"github":  tool.GitHubURL,
		},
	}
}

func (api *API) categoryToolCount(categoryID string) int {
	tools, _ := api.store.ListTools(domain.ToolFilters{Page: 1, PageSize: 100}, false)
	count := 0
	for _, tool := range tools {
		if tool.CategoryID == categoryID {
			count++
		}
	}
	return count
}

func parseIntDefault(value string, fallback int) int {
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func boolPtr(v bool) *bool {
	return &v
}

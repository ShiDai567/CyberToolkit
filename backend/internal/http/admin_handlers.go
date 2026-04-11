package http

import (
	"net/http"
	"strings"

	"cybertoolkit/backend/internal/domain"
)

func (api *API) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", nil)
		return
	}

	var request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body", nil)
		return
	}

	accessToken, refreshToken, user, err := api.store.Authenticate(request.Email, request.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", err.Error(), nil)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
		"user":         user,
	}, nil)
}

func (api *API) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", nil)
		return
	}

	var request struct {
		Email       string `json:"email"`
		Password    string `json:"password"`
		DisplayName string `json:"displayName"`
	}
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body", nil)
		return
	}
	if request.Email == "" || request.Password == "" || request.DisplayName == "" {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "email, password and displayName are required", nil)
		return
	}

	accessToken, refreshToken, user, err := api.store.Register(request.Email, request.Password, request.DisplayName)
	if err != nil {
		writeError(w, http.StatusConflict, "REGISTER_ERROR", err.Error(), nil)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
		"user":         user,
	}, nil)
}

func (api *API) handleMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", nil)
		return
	}

	auth := r.Header.Get("Authorization")
	token := strings.TrimPrefix(auth, "Bearer ")
	user, ok := api.store.ValidateSession(token)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid token", nil)
		return
	}

	writeJSON(w, http.StatusOK, user, nil)
}

func (api *API) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", nil)
		return
	}

	auth := r.Header.Get("Authorization")
	token := strings.TrimPrefix(auth, "Bearer ")
	api.store.DeleteSession(token)
	writeJSON(w, http.StatusOK, map[string]bool{"loggedOut": true}, nil)
}

func (api *API) handleAdminMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", nil)
		return
	}

	auth := r.Header.Get("Authorization")
	token := strings.TrimPrefix(auth, "Bearer ")
	user, ok := api.store.ValidateSession(token)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid token", nil)
		return
	}
	writeJSON(w, http.StatusOK, user, nil)
}

func (api *API) handleAdminCategories(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, api.store.ListCategories(false), nil)
	case http.MethodPost:
		var request struct {
			Slug        string `json:"slug"`
			Name        string `json:"name"`
			Description string `json:"description"`
			Icon        string `json:"icon"`
			SortOrder   int    `json:"sortOrder"`
			IsVisible   bool   `json:"isVisible"`
		}
		if err := decodeJSON(r, &request); err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body", nil)
			return
		}
		if request.Slug == "" || request.Name == "" {
			writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "slug and name are required", nil)
			return
		}
		category := api.store.CreateCategory(domain.Category{
			Slug:        request.Slug,
			Name:        request.Name,
			Description: request.Description,
			Icon:        request.Icon,
			SortOrder:   request.SortOrder,
			IsVisible:   request.IsVisible,
		})
		writeJSON(w, http.StatusCreated, category, nil)
	default:
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", nil)
	}
}

func (api *API) handleAdminCategoryByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/categories/")
	if id == "" {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "category not found", nil)
		return
	}

	switch r.Method {
	case http.MethodPatch:
		existing, ok := api.store.CategoryByID(id)
		if !ok {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "category not found", nil)
			return
		}
		var request struct {
			Slug        string `json:"slug"`
			Name        string `json:"name"`
			Description string `json:"description"`
			Icon        string `json:"icon"`
			SortOrder   int    `json:"sortOrder"`
			IsVisible   *bool  `json:"isVisible"`
		}
		if err := decodeJSON(r, &request); err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body", nil)
			return
		}
		if request.Slug != "" {
			existing.Slug = request.Slug
		}
		if request.Name != "" {
			existing.Name = request.Name
		}
		if request.Description != "" {
			existing.Description = request.Description
		}
		if request.Icon != "" {
			existing.Icon = request.Icon
		}
		if request.SortOrder != 0 {
			existing.SortOrder = request.SortOrder
		}
		if request.IsVisible != nil {
			existing.IsVisible = *request.IsVisible
		}
		category, err := api.store.UpdateCategory(id, existing)
		if err != nil {
			writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error(), nil)
			return
		}
		writeJSON(w, http.StatusOK, category, nil)
	case http.MethodDelete:
		existing, ok := api.store.CategoryByID(id)
		if !ok {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "category not found", nil)
			return
		}
		existing.IsVisible = false
		category, _ := api.store.UpdateCategory(id, existing)
		writeJSON(w, http.StatusOK, category, nil)
	default:
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", nil)
	}
}

func (api *API) handleAdminTools(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		filters := domain.ToolFilters{
			Query:      r.URL.Query().Get("q"),
			Status:     r.URL.Query().Get("status"),
			Page:       parseIntDefault(r.URL.Query().Get("page"), 1),
			PageSize:   parseIntDefault(r.URL.Query().Get("pageSize"), 20),
			Category:   r.URL.Query().Get("category"),
			Difficulty: r.URL.Query().Get("difficulty"),
			Sort:       r.URL.Query().Get("sort"),
		}
		tools, total := api.store.ListTools(filters, true)
		pageSize := filters.PageSize
		if pageSize < 1 {
			pageSize = 20
		}
		writeJSON(w, http.StatusOK, tools, meta{
			"page":       max(filters.Page, 1),
			"pageSize":   pageSize,
			"total":      total,
			"totalPages": (total + pageSize - 1) / pageSize,
		})
	case http.MethodPost:
		var request struct {
			Slug             string   `json:"slug"`
			Name             string   `json:"name"`
			ShortDescription string   `json:"shortDescription"`
			LongDescription  string   `json:"longDescription"`
			CategoryID       string   `json:"categoryId"`
			Difficulty       string   `json:"difficulty"`
			Icon             string   `json:"icon"`
			Featured         bool     `json:"featured"`
			Status           string   `json:"status"`
			WebsiteURL       string   `json:"websiteUrl"`
			GitHubURL        string   `json:"githubUrl"`
			Tags             []string `json:"tags"`
		}
		if err := decodeJSON(r, &request); err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body", nil)
			return
		}
		if request.Slug == "" || request.Name == "" || request.CategoryID == "" || request.WebsiteURL == "" {
			writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "missing required fields", nil)
			return
		}
		tool := api.store.CreateTool(domain.Tool{
			Slug:             request.Slug,
			Name:             request.Name,
			ShortDescription: request.ShortDescription,
			LongDescription:  request.LongDescription,
			CategoryID:       request.CategoryID,
			Difficulty:       domain.Difficulty(request.Difficulty),
			Icon:             request.Icon,
			Featured:         request.Featured,
			Status:           domain.ToolStatus(request.Status),
			WebsiteURL:       request.WebsiteURL,
			GitHubURL:        request.GitHubURL,
		})
		api.store.ReplaceToolTags(tool.ID, request.Tags)
		writeJSON(w, http.StatusCreated, tool, nil)
	default:
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", nil)
	}
}

func (api *API) handleAdminToolByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/tools/")
	if id == "" {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "tool not found", nil)
		return
	}

	switch r.Method {
	case http.MethodGet:
		tool, ok := api.store.GetToolByID(id)
		if !ok {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "tool not found", nil)
			return
		}
		writeJSON(w, http.StatusOK, tool, nil)
	case http.MethodPatch:
		existing, ok := api.store.GetToolByID(id)
		if !ok {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "tool not found", nil)
			return
		}
		var request struct {
			Slug             string   `json:"slug"`
			Name             string   `json:"name"`
			ShortDescription string   `json:"shortDescription"`
			LongDescription  string   `json:"longDescription"`
			CategoryID       string   `json:"categoryId"`
			Difficulty       string   `json:"difficulty"`
			Icon             string   `json:"icon"`
			Featured         *bool    `json:"featured"`
			Status           string   `json:"status"`
			WebsiteURL       string   `json:"websiteUrl"`
			GitHubURL        string   `json:"githubUrl"`
			Tags             []string `json:"tags"`
		}
		if err := decodeJSON(r, &request); err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body", nil)
			return
		}
		if request.Slug != "" {
			existing.Slug = request.Slug
		}
		if request.Name != "" {
			existing.Name = request.Name
		}
		if request.ShortDescription != "" {
			existing.ShortDescription = request.ShortDescription
		}
		if request.LongDescription != "" {
			existing.LongDescription = request.LongDescription
		}
		if request.CategoryID != "" {
			existing.CategoryID = request.CategoryID
		}
		if request.Difficulty != "" {
			existing.Difficulty = domain.Difficulty(request.Difficulty)
		}
		if request.Icon != "" {
			existing.Icon = request.Icon
		}
		if request.Featured != nil {
			existing.Featured = *request.Featured
		}
		if request.Status != "" {
			existing.Status = domain.ToolStatus(request.Status)
		}
		if request.WebsiteURL != "" {
			existing.WebsiteURL = request.WebsiteURL
		}
		if request.GitHubURL != "" {
			existing.GitHubURL = request.GitHubURL
		}
		tool, err := api.store.UpdateTool(id, existing)
		if err != nil {
			writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error(), nil)
			return
		}
		if request.Tags != nil {
			api.store.ReplaceToolTags(id, request.Tags)
		}
		writeJSON(w, http.StatusOK, tool, nil)
	case http.MethodDelete:
		if err := api.store.ArchiveTool(id); err != nil {
			writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error(), nil)
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"archived": true}, nil)
	default:
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", nil)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

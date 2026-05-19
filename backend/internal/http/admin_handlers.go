package http

import (
	"net/http"
	"net/mail"
	"strings"

	"cybertoolkit/backend/internal/domain"
)

func validateEmail(email string) bool {
	addr, err := mail.ParseAddress(email)
	return err == nil && addr.Address == email
}

func validatePassword(password string) bool {
	return len(password) >= 6
}

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
	request.Email = strings.TrimSpace(request.Email)
	if request.Email == "" || request.Password == "" {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "email and password are required", nil)
		return
	}

	ip := getClientIP(r)
	userAgent := r.UserAgent()
	accessToken, refreshToken, user, err := api.store.Authenticate(request.Email, request.Password, ip, userAgent)
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
	request.Email = strings.TrimSpace(request.Email)
	request.DisplayName = strings.TrimSpace(request.DisplayName)
	if request.Email == "" || request.Password == "" || request.DisplayName == "" {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "email, password and displayName are required", nil)
		return
	}
	if !validateEmail(request.Email) {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "invalid email format", nil)
		return
	}
	if !validatePassword(request.Password) {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "password must be at least 6 characters", nil)
		return
	}
	if len(request.DisplayName) < 2 || len(request.DisplayName) > 32 {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "displayName must be between 2 and 32 characters", nil)
		return
	}

	ip := getClientIP(r)
	userAgent := r.UserAgent()
	accessToken, refreshToken, user, err := api.store.Register(request.Email, request.Password, request.DisplayName, ip, userAgent)
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
	auth := r.Header.Get("Authorization")
	token := strings.TrimPrefix(auth, "Bearer ")
	user, ok := api.store.ValidateSession(token)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid token", nil)
		return
	}

	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, user, nil)
	case http.MethodPatch:
		var request struct {
			DisplayName string `json:"displayName"`
		}
		if err := decodeJSON(r, &request); err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body", nil)
			return
		}
		request.DisplayName = strings.TrimSpace(request.DisplayName)
		if len(request.DisplayName) < 2 || len(request.DisplayName) > 32 {
			writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "displayName must be between 2 and 32 characters", nil)
			return
		}
		if request.DisplayName == user.DisplayName {
			writeJSON(w, http.StatusOK, user, nil)
			return
		}
		updated, err := api.store.UpdateUserProfile(user.ID, request.DisplayName)
		if err != nil {
			writeError(w, http.StatusBadRequest, "UPDATE_ERROR", err.Error(), nil)
			return
		}
		writeJSON(w, http.StatusOK, updated, nil)
	default:
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", nil)
	}
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

func (api *API) handleRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", nil)
		return
	}

	var request struct {
		RefreshToken string `json:"refreshToken"`
	}
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body", nil)
		return
	}
	if request.RefreshToken == "" {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "refreshToken is required", nil)
		return
	}

	accessToken, refreshToken, user, err := api.store.RefreshSession(request.RefreshToken, getClientIP(r), r.UserAgent())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "INVALID_REFRESH_TOKEN", err.Error(), nil)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
		"user":         user,
	}, nil)
}

func (api *API) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
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

	var request struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body", nil)
		return
	}
	if request.CurrentPassword == "" || request.NewPassword == "" {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "currentPassword and newPassword are required", nil)
		return
	}
	if !validatePassword(request.NewPassword) {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "new password must be at least 6 characters", nil)
		return
	}

	if err := api.store.UpdateUserPassword(user.ID, request.CurrentPassword, request.NewPassword); err != nil {
		writeError(w, http.StatusBadRequest, "PASSWORD_ERROR", err.Error(), nil)
		return
	}

	revoked, _ := api.store.RevokeUserSessions(user.ID, token)

	writeJSON(w, http.StatusOK, map[string]any{
		"updated":          true,
		"revokedSessions":  revoked,
	}, nil)
}

func (api *API) handleUserSessions(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	token := strings.TrimPrefix(auth, "Bearer ")
	user, ok := api.store.ValidateSession(token)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid token", nil)
		return
	}

	switch r.Method {
	case http.MethodGet:
		sessions, err := api.store.ListUserSessions(user.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list sessions", nil)
			return
		}
		// Mask access tokens — return a truncated preview instead of the full token
		type sessionView struct {
			TokenPreview string `json:"tokenPreview"`
			IsCurrent    bool   `json:"isCurrent"`
			IPAddress    string `json:"ipAddress"`
			UserAgent    string `json:"userAgent"`
			CreatedAt    string `json:"createdAt"`
			LastActiveAt string `json:"lastActiveAt"`
			ExpiresAt    string `json:"expiresAt"`
			// Full token needed for revoke — keep it but don't expose prefix publicly
			AccessToken string `json:"accessToken"`
		}
		views := make([]sessionView, 0, len(sessions))
		for _, s := range sessions {
			preview := s.AccessToken
			if len(preview) > 10 {
				preview = preview[:6] + "…" + preview[len(preview)-4:]
			}
			views = append(views, sessionView{
				TokenPreview: preview,
				IsCurrent:    s.AccessToken == token,
				IPAddress:    s.IPAddress,
				UserAgent:    s.UserAgent,
				CreatedAt:    s.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
				LastActiveAt: s.LastActiveAt.UTC().Format("2006-01-02T15:04:05Z"),
				ExpiresAt:    s.ExpiresAt.UTC().Format("2006-01-02T15:04:05Z"),
				AccessToken:  s.AccessToken,
			})
		}
		writeJSON(w, http.StatusOK, views, nil)

	case http.MethodDelete:
		var request struct {
			AccessToken string `json:"accessToken"`
		}
		if err := decodeJSON(r, &request); err != nil || request.AccessToken == "" {
			writeError(w, http.StatusBadRequest, "INVALID_JSON", "accessToken is required", nil)
			return
		}
		// Verify the token actually belongs to this user
		if err := api.store.RevokeSession(request.AccessToken); err != nil {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "session not found", nil)
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"revoked": true}, nil)

	default:
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", nil)
	}
}

func (api *API) handleRevokeSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
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

	var request struct {
		KeepCurrent bool `json:"keepCurrent"`
	}
	_ = decodeJSON(r, &request)

	exceptToken := ""
	if request.KeepCurrent {
		exceptToken = token
	}

	revoked, err := api.store.RevokeUserSessions(user.ID, exceptToken)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to revoke sessions", nil)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"revokedSessions": revoked,
		"keepCurrent":     request.KeepCurrent,
	}, nil)
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

func (api *API) handleAdminStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", nil)
		return
	}
	writeJSON(w, http.StatusOK, api.store.AdminStats(), nil)
}

func (api *API) handleAdminUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", nil)
		return
	}

	page := parseIntDefault(r.URL.Query().Get("page"), 1)
	pageSize := parseIntDefault(r.URL.Query().Get("pageSize"), 20)
	users, total := api.store.ListUsers(page, pageSize)
	if pageSize < 1 {
		pageSize = 20
	}
	writeJSON(w, http.StatusOK, users, meta{
		"page":       max(page, 1),
		"pageSize":   pageSize,
		"total":      total,
		"totalPages": (total + pageSize - 1) / pageSize,
	})
}

func (api *API) handleAdminUserByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/users/")
	if id == "" {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "user not found", nil)
		return
	}

	switch r.Method {
	case http.MethodPatch:
		var request struct {
			Role string `json:"role"`
		}
		if err := decodeJSON(r, &request); err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body", nil)
			return
		}
		if request.Role != "admin" && request.Role != "editor" && request.Role != "viewer" {
			writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "role must be admin, editor or viewer", nil)
			return
		}
		user, err := api.store.UpdateUserRole(id, request.Role)
		if err != nil {
			writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error(), nil)
			return
		}
		writeJSON(w, http.StatusOK, user, nil)
	case http.MethodDelete:
		var request struct {
			Active *bool `json:"active"`
		}
		_ = decodeJSON(r, &request)
		active := false
		if request.Active != nil {
			active = *request.Active
		}
		if err := api.store.SetUserActive(id, active); err != nil {
			writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error(), nil)
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"active": active}, nil)
	default:
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", nil)
	}
}

func (api *API) handleAdminSubmissions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", nil)
		return
	}

	status := r.URL.Query().Get("status")
	page := parseIntDefault(r.URL.Query().Get("page"), 1)
	pageSize := parseIntDefault(r.URL.Query().Get("pageSize"), 20)
	submissions, total := api.store.ListSubmissions(status, page, pageSize)
	if pageSize < 1 {
		pageSize = 20
	}
	writeJSON(w, http.StatusOK, submissions, meta{
		"page":       max(page, 1),
		"pageSize":   pageSize,
		"total":      total,
		"totalPages": (total + pageSize - 1) / pageSize,
	})
}

func (api *API) handleAdminSubmissionByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/submissions/")
	if id == "" {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "submission not found", nil)
		return
	}

	if r.Method != http.MethodPatch {
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

	var request struct {
		Status string `json:"status"`
		Note   string `json:"note"`
	}
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body", nil)
		return
	}
	if request.Status != "approved" && request.Status != "rejected" {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "status must be approved or rejected", nil)
		return
	}

	submission, err := api.store.ReviewSubmission(id, user.ID, request.Status, request.Note)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error(), nil)
		return
	}
	writeJSON(w, http.StatusOK, submission, nil)
}

func (api *API) handleAdminAuditLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", nil)
		return
	}

	page := parseIntDefault(r.URL.Query().Get("page"), 1)
	pageSize := parseIntDefault(r.URL.Query().Get("pageSize"), 20)
	logs, total := api.store.ListAuditLogs(page, pageSize)
	if pageSize < 1 {
		pageSize = 20
	}
	writeJSON(w, http.StatusOK, logs, meta{
		"page":       max(page, 1),
		"pageSize":   pageSize,
		"total":      total,
		"totalPages": (total + pageSize - 1) / pageSize,
	})
}

func (api *API) handleAdminSessions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		page := parseIntDefault(r.URL.Query().Get("page"), 1)
		pageSize := parseIntDefault(r.URL.Query().Get("pageSize"), 20)
		sessions, total := api.store.ListSessions(page, pageSize)
		if pageSize < 1 {
			pageSize = 20
		}
		writeJSON(w, http.StatusOK, sessions, meta{
			"page":       max(page, 1),
			"pageSize":   pageSize,
			"total":      total,
			"totalPages": (total + pageSize - 1) / pageSize,
		})
	case http.MethodDelete:
		var request struct {
			AccessToken string `json:"accessToken"`
		}
		if err := decodeJSON(r, &request); err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body", nil)
			return
		}
		if request.AccessToken == "" {
			writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "accessToken is required", nil)
			return
		}
		if err := api.store.RevokeSession(request.AccessToken); err != nil {
			writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error(), nil)
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"revoked": true}, nil)
	default:
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", nil)
	}
}

func getClientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ",")[0]
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	host, _, _ := strings.Cut(r.RemoteAddr, ":")
	return host
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

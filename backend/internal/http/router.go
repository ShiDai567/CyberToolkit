package http

import (
	"net/http"
	"strings"

	"cybertoolkit/backend/internal/config"
	"cybertoolkit/backend/internal/store/memory"
)

type API struct {
	config config.Config
	store  *memory.Store
}

func NewRouter(cfg config.Config, store *memory.Store) http.Handler {
	api := &API{config: cfg, store: store}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/health", api.handleHealth)
	mux.HandleFunc("/api/v1/home", api.handleHome)
	mux.HandleFunc("/api/v1/categories", api.handleCategories)
	mux.HandleFunc("/api/v1/tags", api.handleTags)
	mux.HandleFunc("/api/v1/tools", api.handleTools)
	mux.HandleFunc("/api/v1/tools/", api.handleToolDetail)
	mux.HandleFunc("/api/v1/submissions", api.handleSubmissions)
	mux.HandleFunc("/api/v1/admin/auth/login", api.handleAdminLogin)
	mux.HandleFunc("/api/v1/admin/me", api.requireAdmin(api.handleAdminMe))
	mux.HandleFunc("/api/v1/admin/categories", api.requireAdmin(api.handleAdminCategories))
	mux.HandleFunc("/api/v1/admin/categories/", api.requireAdmin(api.handleAdminCategoryByID))
	mux.HandleFunc("/api/v1/admin/tools", api.requireAdmin(api.handleAdminTools))
	mux.HandleFunc("/api/v1/admin/tools/", api.requireAdmin(api.handleAdminToolByID))

	return withJSON(mux)
}

func withJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		next.ServeHTTP(w, r)
	})
}

func (api *API) requireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing bearer token", nil)
			return
		}
		token := strings.TrimPrefix(auth, "Bearer ")
		if token != api.config.AdminToken {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid token", nil)
			return
		}
		next(w, r)
	}
}

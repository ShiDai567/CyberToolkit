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
	mux.HandleFunc("/api/v1/auth/login", api.handleLogin)
	mux.HandleFunc("/api/v1/auth/register", api.handleRegister)
	mux.HandleFunc("/api/v1/auth/me", api.requireAuth(api.handleMe))
	mux.HandleFunc("/api/v1/auth/logout", api.requireAuth(api.handleLogout))
	mux.HandleFunc("/api/v1/auth/refresh", api.handleRefresh)
	mux.HandleFunc("/api/v1/admin/auth/login", api.handleLogin)
	mux.HandleFunc("/api/v1/admin/me", api.requireAdmin(api.handleAdminMe))
	mux.HandleFunc("/api/v1/admin/categories", api.requireAdmin(api.handleAdminCategories))
	mux.HandleFunc("/api/v1/admin/categories/", api.requireAdmin(api.handleAdminCategoryByID))
	mux.HandleFunc("/api/v1/admin/tools", api.requireAdmin(api.handleAdminTools))
	mux.HandleFunc("/api/v1/admin/tools/", api.requireAdmin(api.handleAdminToolByID))

	return withCORS(withJSON(mux), cfg.CORSOrigins)
}

func withJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		next.ServeHTTP(w, r)
	})
}

func withCORS(next http.Handler, allowedOrigins []string) http.Handler {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	allowAll := false
	for _, origin := range allowedOrigins {
		if origin == "*" {
			allowAll = true
		} else {
			allowed[origin] = struct{}{}
		}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			if allowAll {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
			} else if _, ok := allowed[origin]; ok {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
			}
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "600")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (api *API) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing bearer token", nil)
			return
		}
		token := strings.TrimPrefix(auth, "Bearer ")
		if _, ok := api.store.ValidateSession(token); !ok {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid token", nil)
			return
		}
		next(w, r)
	}
}

func (api *API) requireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
		if user.Role != "admin" {
			writeError(w, http.StatusForbidden, "FORBIDDEN", "admin access required", nil)
			return
		}
		next(w, r)
	}
}

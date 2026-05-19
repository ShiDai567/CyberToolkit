package http

import (
	"log"
	"net/http"
	"strings"
	"time"

	"cybertoolkit/backend/internal/config"
	"cybertoolkit/backend/internal/store"
)

type API struct {
	config config.Config
	store  store.Store
}

func NewRouter(cfg config.Config, store store.Store) http.Handler {
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
	mux.HandleFunc("/api/v1/auth/password", api.requireAuth(api.handleChangePassword))
	mux.HandleFunc("/api/v1/auth/sessions/revoke", api.requireAuth(api.handleRevokeSessions))
	mux.HandleFunc("/api/v1/auth/sessions", api.requireAuth(api.handleUserSessions))
	mux.HandleFunc("/api/v1/admin/auth/login", api.handleLogin)
	mux.HandleFunc("/api/v1/admin/me", api.requireAdmin(api.handleAdminMe))
	mux.HandleFunc("/api/v1/admin/categories", api.requireAdmin(api.handleAdminCategories))
	mux.HandleFunc("/api/v1/admin/categories/", api.requireAdmin(api.handleAdminCategoryByID))
	mux.HandleFunc("/api/v1/admin/tools", api.requireAdmin(api.handleAdminTools))
	mux.HandleFunc("/api/v1/admin/tools/", api.requireAdmin(api.handleAdminToolByID))
	mux.HandleFunc("/api/v1/admin/stats", api.requireAdmin(api.handleAdminStats))
	mux.HandleFunc("/api/v1/admin/users", api.requireAdmin(api.handleAdminUsers))
	mux.HandleFunc("/api/v1/admin/users/", api.requireAdmin(api.handleAdminUserByID))
	mux.HandleFunc("/api/v1/admin/submissions", api.requireAdmin(api.handleAdminSubmissions))
	mux.HandleFunc("/api/v1/admin/submissions/", api.requireAdmin(api.handleAdminSubmissionByID))
	mux.HandleFunc("/api/v1/admin/audit-logs", api.requireAdmin(api.handleAdminAuditLogs))
	mux.HandleFunc("/api/v1/admin/sessions", api.requireAdmin(api.handleAdminSessions))

	return withCORS(withJSON(withLogger(mux)), cfg.CORSOrigins)
}

func withLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("[%s] %s %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
		log.Printf("[%s] %s completed in %v", r.Method, r.URL.Path, time.Since(start))
	})
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

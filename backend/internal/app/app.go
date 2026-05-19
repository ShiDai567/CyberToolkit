package app

import (
	"net/http"

	"cybertoolkit/backend/internal/config"
	httpapi "cybertoolkit/backend/internal/http"
	"cybertoolkit/backend/internal/store/postgres"
)

type Server struct {
	Addr    string
	Handler http.Handler
}

func New() (*Server, error) {
	cfg := config.Load()
	store, err := postgres.NewStore(cfg.DatabaseURL, cfg.RedisURL, cfg.AdminUsername, cfg.AdminEmail, cfg.AdminPassword, cfg.AdminName)
	if err != nil {
		return nil, err
	}
	handler := httpapi.NewRouter(cfg, store)

	return &Server{
		Addr:    cfg.Addr,
		Handler: handler,
	}, nil
}

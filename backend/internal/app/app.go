package app

import (
	"net/http"

	"cybertoolkit/backend/internal/config"
	httpapi "cybertoolkit/backend/internal/http"
	"cybertoolkit/backend/internal/store/memory"
)

type Server struct {
	Addr    string
	Handler http.Handler
}

func New() (*Server, error) {
	cfg := config.Load()
	store := memory.NewStore()
	handler := httpapi.NewRouter(cfg, store)

	return &Server{
		Addr:    cfg.Addr,
		Handler: handler,
	}, nil
}

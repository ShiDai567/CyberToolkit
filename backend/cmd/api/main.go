package main

import (
	"log"
	"net/http"

	"cybertoolkit/backend/internal/app"
)

func main() {
	server, err := app.New()
	if err != nil {
		log.Fatalf("init app: %v", err)
	}

	log.Printf("cybertoolkit backend listening on %s", server.Addr)
	if err := http.ListenAndServe(server.Addr, server.Handler); err != nil {
		log.Fatalf("listen: %v", err)
	}
}

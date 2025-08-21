package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi"
	chimiddleware "github.com/go-chi/chi/middleware"
	"github.com/greenbrown932/fire-pmaas/pkg/api"
	firemiddleware "github.com/greenbrown932/fire-pmaas/pkg/middleware"
)

func main() {
	r := chi.NewRouter()

	// Middleware
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)

	if err := firemiddleware.InitOIDC(); err != nil {
		log.Fatalf("Failed to initialize OIDC: %v", err)
	}

	api.RegisterRoutes(r)

	log.Println("Starting server on port :8000")
	if err := http.ListenAndServe(":8000", r); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

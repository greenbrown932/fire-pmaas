package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/greenbrown932/fire-pmaas/pkg/api"
)

func main() {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	api.RegisterRoutes(r)

	log.Println("Starting server on :8000")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Error starting server :%v", err)
	}
}

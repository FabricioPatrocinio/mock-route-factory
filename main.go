package main

import (
	_ "embed"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"

	"github.com/FabricioPatrocinio/mock-route-factory/internal/config"
	"github.com/FabricioPatrocinio/mock-route-factory/internal/db"
	"github.com/FabricioPatrocinio/mock-route-factory/internal/handler"
	"github.com/FabricioPatrocinio/mock-route-factory/internal/migrate"
	"github.com/FabricioPatrocinio/mock-route-factory/internal/repo"
)

//go:embed openapi/openapi.yaml
var swaggerSpec []byte

func main() {
	_ = godotenv.Load() // .env is optional; missing file is not an error

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	writer, reader, err := db.Open(cfg)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer writer.Close()
	defer reader.Close()

	if err := migrate.Run(writer, cfg.DBSchema); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	mockRepo := repo.New(writer, reader, cfg.DBSchema)

	mux := chi.NewRouter()
	mux.Use(middleware.RequestID)
	mux.Use(middleware.RealIP)
	mux.Use(middleware.Logger)
	mux.Use(middleware.Recoverer)

	mux.Get("/health", handler.Health)
	mux.Get("/swagger", handler.SwaggerUI)
	mux.Get("/swagger/openapi.yaml", handler.SwaggerSpec(swaggerSpec))

	adminH := handler.NewAdmin(mockRepo, cfg.PublicBaseURL)
	mux.Route("/admin", func(r chi.Router) {
		r.Post("/mocks", adminH.Upsert)
		r.Get("/mocks", adminH.List)
		r.Get("/mocks/curls", adminH.Curls)
		r.Delete("/mocks", adminH.Delete)
	})

	// Catch-all: resolve every other method+path against stored mocks.
	dynH := handler.NewDynamic(mockRepo)
	mux.NotFound(dynH.Handle)
	mux.MethodNotAllowed(dynH.Handle)

	addr := ":" + cfg.Port
	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

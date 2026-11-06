package http

import (
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"geotrack/backend/internal/config"
	customMiddleware "geotrack/backend/internal/middleware"
)

// NewRouter sets up the Chi HTTP multiplexer and registers all middlewares & routes
func NewRouter(
	cfg *config.Config,
	healthHandler *HealthHandler,
	entityHandler *EntityHandler,
	activityHandler *ActivityLogHandler,
	dashboardHandler *DashboardHandler,
) *chi.Mux {
	r := chi.NewRouter()

	// Deterministic security & observability pipeline
	r.Use(customMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(StructuredLogger())
	r.Use(customMiddleware.Recoverer)
	r.Use(customMiddleware.SecurityHeaders)
	r.Use(customMiddleware.CORS(cfg.App.AllowedOrigins))

	rateLimitPerMin := 120
	rateBurst := 60
	if cfg.App.Env == "development" {
		rateLimitPerMin = 600
		rateBurst = 200
	}
	r.Use(customMiddleware.RateLimit(rateLimitPerMin, rateBurst))
	r.Use(customMiddleware.BodyLimit(1 << 20)) // 1 MB limit
	r.Use(chiMiddleware.Timeout(60 * time.Second))

	// API Route groups
	r.Route("/api/v1", func(v1 chi.Router) {
		if healthHandler != nil {
			v1.Get("/health", healthHandler.Check)
		}

		if entityHandler != nil {
			v1.Route("/entities", func(er chi.Router) {
				er.Get("/", entityHandler.List)
				er.Post("/", entityHandler.Create)
				er.Get("/{id}", entityHandler.GetByID)
				er.Put("/{id}", entityHandler.Update)
				er.Delete("/{id}", entityHandler.Delete)
			})
		}

		if activityHandler != nil {
			v1.Route("/activities", func(ar chi.Router) {
				ar.Get("/", activityHandler.List)
				ar.Get("/{id}", activityHandler.GetByID)
			})
		}

		if dashboardHandler != nil {
			v1.Route("/dashboard", func(dr chi.Router) {
				dr.Get("/stats", dashboardHandler.GetStats)
			})
		}
	})

	// Swagger UI & OpenAPI Specification routes
	RegisterSwaggerRoutes(r, "docs/swagger.json")

	return r
}

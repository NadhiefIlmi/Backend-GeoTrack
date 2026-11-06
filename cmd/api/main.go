package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"geotrack/backend/internal/config"
	deliveryHTTP "geotrack/backend/internal/delivery/http"
	"geotrack/backend/internal/repository/postgres"
	"geotrack/backend/internal/usecase"
	"geotrack/backend/pkg/logger"
)

func main() {
	// 1. Load Application Configuration
	cfg, err := config.LoadConfig(".")
	if err != nil {
		fmt.Printf("FATAL: Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// 2. Initialize Structured Logger
	logger.Init(cfg.App.Env)
	logger.Info().
		Str("app", cfg.App.Name).
		Str("env", cfg.App.Env).
		Int("port", cfg.App.Port).
		Strs("allowed_origins", cfg.App.AllowedOrigins).
		Msg("Configuration loaded and logger initialized")

	// 3. Connect to PostgreSQL Connection Pool
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var dbPool *pgxpool.Pool
	dbPool, err = postgres.NewPool(ctx, cfg.Database)
	if err != nil {
		logger.Warn().Err(err).Msg("Database connection failed. Continuing in offline mode.")
	} else {
		defer dbPool.Close()
	}

	// 4. Initialize Repositories & Usecases
	var entityHandler *deliveryHTTP.EntityHandler
	var activityHandler *deliveryHTTP.ActivityLogHandler
	var dashboardHandler *deliveryHTTP.DashboardHandler

	if dbPool != nil {
		activityRepo := postgres.NewActivityLogRepository(dbPool)
		activityUsecase := usecase.NewActivityLogUsecase(activityRepo)
		activityHandler = deliveryHTTP.NewActivityLogHandler(activityUsecase)

		entityRepo := postgres.NewEntityRepository(dbPool)
		entityUsecase := usecase.NewEntityUsecase(entityRepo, activityRepo)
		entityHandler = deliveryHTTP.NewEntityHandler(entityUsecase)

		dashboardRepo := postgres.NewDashboardRepository(dbPool)
		dashboardUsecase := usecase.NewDashboardUsecase(dashboardRepo, activityRepo)
		dashboardHandler = deliveryHTTP.NewDashboardHandler(dashboardUsecase)
	}

	// 5. Initialize Delivery Handlers
	healthHandler := deliveryHTTP.NewHealthHandler(cfg.App.Env, "1.0.0", dbPool)

	// 6. Setup Router
	router := deliveryHTTP.NewRouter(cfg, healthHandler, entityHandler, activityHandler, dashboardHandler)

	// 7. Construct HTTP Server
	serverAddr := fmt.Sprintf(":%d", cfg.App.Port)
	server := &http.Server{
		Addr:         serverAddr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 8. Start Server in Background Goroutine
	serverErrors := make(chan error, 1)
	go func() {
		logger.Info().Str("address", serverAddr).Msg("GeoTrack API Server is listening")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	// 9. Setup OS Signal Listener for Graceful Shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		logger.Fatal().Err(err).Msg("Server crashed unexpectedly")

	case sig := <-shutdown:
		logger.Info().Str("signal", sig.String()).Msg("Graceful shutdown initiated")

		// 10. Create Graceful Shutdown Deadline
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error().Err(err).Msg("Server forced to shutdown after timeout")
			if err := server.Close(); err != nil {
				logger.Error().Err(err).Msg("Error closing server")
			}
		}

		logger.Info().Msg("GeoTrack API Server stopped cleanly")
	}
}

package main

import (
	"context"
	"fmt"
	"os"

	"geotrack/backend/internal/config"
	"geotrack/backend/internal/repository/postgres"
	"geotrack/backend/pkg/logger"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run cmd/migrate/main.go [up|down|seed]")
		os.Exit(1)
	}

	command := os.Args[1]

	cfg, err := config.LoadConfig(".")
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	logger.Init(cfg.App.Env)

	ctx := context.Background()
	pool, err := postgres.NewPool(ctx, cfg.Database)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to database for migration")
	}
	defer pool.Close()

	switch command {
	case "up":
		logger.Info().Msg("Starting migration up...")
		if err := postgres.MigrateUp(ctx, pool, "migrations"); err != nil {
			logger.Fatal().Err(err).Msg("Migration up failed")
		}
		logger.Info().Msg("All migrations applied successfully!")

	case "down":
		logger.Info().Msg("Starting migration down...")
		if err := postgres.MigrateDown(ctx, pool, "migrations"); err != nil {
			logger.Fatal().Err(err).Msg("Migration down failed")
		}
		logger.Info().Msg("All migrations rolled back successfully!")

	case "seed":
		logger.Info().Msg("Starting seeder...")
		if err := postgres.SeedData(ctx, pool, "migrations/seeds"); err != nil {
			logger.Fatal().Err(err).Msg("Seeder execution failed")
		}
		logger.Info().Msg("All seeds executed successfully!")

	default:
		fmt.Printf("Unknown command: %s. Supported: up, down, seed\n", command)
		os.Exit(1)
	}
}


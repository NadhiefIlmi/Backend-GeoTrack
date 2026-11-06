package postgres

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"geotrack/backend/pkg/logger"
)

// EnsureSchemaMigrationsTable creates the schema_migrations tracking table if it doesn't exist
func EnsureSchemaMigrationsTable(ctx context.Context, pool *pgxpool.Pool) error {
	query := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);`
	_, err := pool.Exec(ctx, query)
	return err
}

// MigrateUp executes all pending .up.sql files in chronological order
func MigrateUp(ctx context.Context, pool *pgxpool.Pool, migrationsDir string) error {
	if err := EnsureSchemaMigrationsTable(ctx, pool); err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.up.sql"))
	if err != nil {
		return fmt.Errorf("failed to scan migration files: %w", err)
	}
	sort.Strings(files)

	for _, file := range files {
		baseName := filepath.Base(file)
		version := strings.TrimSuffix(baseName, ".up.sql")

		// Check if already applied
		var exists bool
		err := pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", version).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check migration status for %s: %w", version, err)
		}

		if exists {
			logger.Info().Str("version", version).Msg("Migration already applied, skipping")
			continue
		}

		// Read and execute SQL file
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to start migration transaction: %w", err)
		}

		if _, err := tx.Exec(ctx, string(content)); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("failed to execute migration %s: %w", version, err)
		}

		if _, err := tx.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", version); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("failed to record migration %s: %w", version, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit migration transaction: %w", err)
		}

		logger.Info().Str("version", version).Msg("Migration applied successfully")
	}

	return nil
}

// MigrateDown rolls back all applied migrations in reverse order
func MigrateDown(ctx context.Context, pool *pgxpool.Pool, migrationsDir string) error {
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.down.sql"))
	if err != nil {
		return fmt.Errorf("failed to scan down migration files: %w", err)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(files)))

	for _, file := range files {
		baseName := filepath.Base(file)
		version := strings.TrimSuffix(baseName, ".down.sql")

		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read down migration file %s: %w", file, err)
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to start transaction: %w", err)
		}

		if _, err := tx.Exec(ctx, string(content)); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("failed to rollback migration %s: %w", version, err)
		}

		if _, err := tx.Exec(ctx, "DELETE FROM schema_migrations WHERE version = $1", version); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("failed to delete migration record %s: %w", version, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit rollback transaction: %w", err)
		}

		logger.Info().Str("version", version).Msg("Migration rolled back successfully")
	}

	return nil
}

// SeedData executes seeder SQL files in migrations/seeds/
func SeedData(ctx context.Context, pool *pgxpool.Pool, seedsDir string) error {
	files, err := filepath.Glob(filepath.Join(seedsDir, "*.sql"))
	if err != nil {
		return fmt.Errorf("failed to scan seed files: %w", err)
	}
	sort.Strings(files)

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read seed file %s: %w", file, err)
		}

		if _, err := pool.Exec(ctx, string(content)); err != nil {
			return fmt.Errorf("failed to execute seed %s: %w", filepath.Base(file), err)
		}

		logger.Info().Str("file", filepath.Base(file)).Msg("Seed data executed successfully")
	}

	return nil
}


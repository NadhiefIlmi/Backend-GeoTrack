package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"geotrack/backend/internal/config"
	"geotrack/backend/pkg/logger"
)

// NewPool creates a production-ready PostgreSQL connection pool using pgxpool
func NewPool(ctx context.Context, cfg config.DBConfig) (*pgxpool.Pool, error) {
	var connStr string
	if strings.TrimSpace(cfg.URL) != "" {
		connStr = strings.TrimSpace(cfg.URL)
	} else if cfg.Password != "" {
		connStr = fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s?sslmode=%s",
			cfg.User,
			cfg.Password,
			cfg.Host,
			cfg.Port,
			cfg.Name,
			cfg.SSLMode,
		)
	} else {
		connStr = fmt.Sprintf(
			"postgres://%s@%s:%d/%s?sslmode=%s",
			cfg.User,
			cfg.Host,
			cfg.Port,
			cfg.Name,
			cfg.SSLMode,
		)
	}

	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database connection string: %w", err)
	}

	// Apply connection pool tuning
	if cfg.MaxConns > 0 {
		poolConfig.MaxConns = int32(cfg.MaxConns)
	} else {
		poolConfig.MaxConns = 25
	}

	if cfg.MinConns > 0 {
		poolConfig.MinConns = int32(cfg.MinConns)
	} else {
		poolConfig.MinConns = 5
	}

	poolConfig.MaxConnLifetime = 1 * time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute
	poolConfig.HealthCheckPeriod = 1 * time.Minute

	// Initialize pool with context timeout
	initCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(initCtx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection pool: %w", err)
	}

	// Verify connectivity via ping
	pingCtx, pingCancel := context.WithTimeout(ctx, 5*time.Second)
	defer pingCancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("database ping check failed: %w", err)
	}

	logger.Info().
		Str("host", cfg.Host).
		Int("port", cfg.Port).
		Str("database", cfg.Name).
		Int32("max_conns", poolConfig.MaxConns).
		Int32("min_conns", poolConfig.MinConns).
		Msg("Database connection pool initialized successfully")

	return pool, nil
}


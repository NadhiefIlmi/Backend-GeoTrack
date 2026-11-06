package http

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthHandler struct {
	startTime   time.Time
	environment string
	version     string
	dbPool      *pgxpool.Pool
}

func NewHealthHandler(env string, version string, dbPool *pgxpool.Pool) *HealthHandler {
	return &HealthHandler{
		startTime:   time.Now(),
		environment: env,
		version:     version,
		dbPool:      dbPool,
	}
}

type HealthData struct {
	Status        string `json:"status"`
	Database      string `json:"database,omitempty"`
	Version       string `json:"version"`
	Environment   string `json:"environment"`
	UptimeSeconds int64  `json:"uptime_seconds"`
	Timestamp     string `json:"timestamp"`
}

func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	uptime := int64(time.Since(h.startTime).Seconds())
	dbStatus := "N/A"

	if h.dbPool != nil {
		pingCtx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := h.dbPool.Ping(pingCtx); err != nil {
			dbStatus = "DOWN"
		} else {
			dbStatus = "UP"
		}
	}

	data := HealthData{
		Status:        "UP",
		Database:      dbStatus,
		Version:       h.version,
		Environment:   h.environment,
		UptimeSeconds: uptime,
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
	}

	SendSuccess(w, http.StatusOK, "Service is healthy and operational", data, nil)
}

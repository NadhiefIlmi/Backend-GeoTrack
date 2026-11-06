package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"geotrack/backend/internal/domain"
)

type dashboardRepository struct {
	pool *pgxpool.Pool
}

// NewDashboardRepository creates a PostgreSQL implementation of domain.DashboardRepository
func NewDashboardRepository(pool *pgxpool.Pool) domain.DashboardRepository {
	return &dashboardRepository{pool: pool}
}

func (r *dashboardRepository) GetAggregateCounts(ctx context.Context) (total int, status domain.StatusBreakdown, types domain.TypeBreakdown, err error) {
	query := `
		SELECT 
			COUNT(*) AS total_entities,
			COUNT(*) FILTER (WHERE status = 'active') AS active_count,
			COUNT(*) FILTER (WHERE status = 'inactive') AS inactive_count,
			COUNT(*) FILTER (WHERE status = 'maintenance') AS maintenance_count,
			COUNT(*) FILTER (WHERE type = 'vehicle') AS vehicle_count,
			COUNT(*) FILTER (WHERE type = 'iot_device') AS iot_count,
			COUNT(*) FILTER (WHERE type = 'facility') AS facility_count,
			COUNT(*) FILTER (WHERE type = 'infrastructure') AS infra_count
		FROM entities;
	`
	row := r.pool.QueryRow(ctx, query)
	err = row.Scan(
		&total,
		&status.Active,
		&status.Inactive,
		&status.Maintenance,
		&types.Vehicles,
		&types.IoTDevices,
		&types.Facilities,
		&types.Infrastructures,
	)
	if err != nil {
		return 0, status, types, fmt.Errorf("failed to fetch dashboard aggregates: %w", err)
	}

	return total, status, types, nil
}


package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"geotrack/backend/internal/domain"
)

type activityLogRepository struct {
	pool *pgxpool.Pool
}

// NewActivityLogRepository creates a PostgreSQL implementation of domain.ActivityLogRepository
func NewActivityLogRepository(pool *pgxpool.Pool) domain.ActivityLogRepository {
	return &activityLogRepository{pool: pool}
}

func (r *activityLogRepository) Create(ctx context.Context, log *domain.ActivityLog) error {
	query := `
		INSERT INTO activity_logs (id, entity_id, entity_code, entity_name, action, details, actor, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.pool.Exec(
		ctx,
		query,
		log.ID,
		log.EntityID,
		log.EntityCode,
		log.EntityName,
		log.Action,
		log.Details,
		log.Actor,
		log.CreatedAt,
	)
	return err
}

func (r *activityLogRepository) List(ctx context.Context, filter domain.ActivityFilter) ([]domain.ActivityLog, int, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	if filter.Search != "" {
		searchTerm := "%" + strings.TrimSpace(filter.Search) + "%"
		conditions = append(conditions, fmt.Sprintf("(entity_name ILIKE $%d OR entity_code ILIKE $%d OR details ILIKE $%d)", argIdx, argIdx, argIdx))
		args = append(args, searchTerm)
		argIdx++
	}

	if filter.Action != nil {
		conditions = append(conditions, fmt.Sprintf("action = $%d", argIdx))
		args = append(args, *filter.Action)
		argIdx++
	}

	if filter.EntityID != "" {
		conditions = append(conditions, fmt.Sprintf("entity_id = $%d", argIdx))
		args = append(args, filter.EntityID)
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// 1. Total count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM activity_logs %s", whereClause)
	var totalRecords int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&totalRecords); err != nil {
		return nil, 0, fmt.Errorf("failed to count activity logs: %w", err)
	}

	// 2. Pagination calculation
	limit := filter.Limit
	if limit <= 0 {
		limit = 10
	} else if limit > 100 {
		limit = 100
	}

	page := filter.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	// 3. Paginated data query
	dataQuery := fmt.Sprintf(`
		SELECT id, entity_id, entity_code, entity_name, action, details, actor, created_at
		FROM activity_logs
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)

	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query activity logs: %w", err)
	}
	defer rows.Close()

	logs := make([]domain.ActivityLog, 0, limit)
	for rows.Next() {
		var l domain.ActivityLog
		if err := rows.Scan(
			&l.ID,
			&l.EntityID,
			&l.EntityCode,
			&l.EntityName,
			&l.Action,
			&l.Details,
			&l.Actor,
			&l.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan activity log row: %w", err)
		}
		logs = append(logs, l)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("activity log rows iteration error: %w", err)
	}

	return logs, totalRecords, nil
}

func (r *activityLogRepository) GetByID(ctx context.Context, id string) (*domain.ActivityLog, error) {
	query := `
		SELECT id, entity_id, entity_code, entity_name, action, details, actor, created_at
		FROM activity_logs
		WHERE id = $1
	`
	var l domain.ActivityLog
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&l.ID,
		&l.EntityID,
		&l.EntityCode,
		&l.EntityName,
		&l.Action,
		&l.Details,
		&l.Actor,
		&l.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrActivityNotFound
		}
		return nil, err
	}
	return &l, nil
}

func (r *activityLogRepository) GetByEntityID(ctx context.Context, entityID string, limit int) ([]domain.ActivityLog, error) {
	if limit <= 0 {
		limit = 10
	}
	query := `
		SELECT id, entity_id, entity_code, entity_name, action, details, actor, created_at
		FROM activity_logs
		WHERE entity_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`
	rows, err := r.pool.Query(ctx, query, entityID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := make([]domain.ActivityLog, 0, limit)
	for rows.Next() {
		var l domain.ActivityLog
		if err := rows.Scan(
			&l.ID,
			&l.EntityID,
			&l.EntityCode,
			&l.EntityName,
			&l.Action,
			&l.Details,
			&l.Actor,
			&l.CreatedAt,
		); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}

	return logs, rows.Err()
}


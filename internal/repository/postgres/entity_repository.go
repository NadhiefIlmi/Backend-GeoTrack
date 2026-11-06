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

type entityRepository struct {
	pool *pgxpool.Pool
}

// NewEntityRepository creates a new PostgreSQL entity repository
func NewEntityRepository(pool *pgxpool.Pool) domain.EntityRepository {
	return &entityRepository{pool: pool}
}

func (r *entityRepository) Create(ctx context.Context, entity *domain.Entity) error {
	query := `
		INSERT INTO entities (id, code, name, type, status, latitude, longitude, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.pool.Exec(
		ctx,
		query,
		entity.ID,
		entity.Code,
		entity.Name,
		entity.Type,
		entity.Status,
		entity.Latitude,
		entity.Longitude,
		entity.Description,
		entity.CreatedAt,
		entity.UpdatedAt,
	)
	return err
}

func (r *entityRepository) GetByID(ctx context.Context, id string) (*domain.Entity, error) {
	query := `
		SELECT id, code, name, type, status, latitude, longitude, COALESCE(description, ''), created_at, updated_at
		FROM entities
		WHERE id = $1
	`
	var e domain.Entity
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&e.ID,
		&e.Code,
		&e.Name,
		&e.Type,
		&e.Status,
		&e.Latitude,
		&e.Longitude,
		&e.Description,
		&e.CreatedAt,
		&e.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrEntityNotFound
		}
		return nil, err
	}
	return &e, nil
}

func (r *entityRepository) GetByCode(ctx context.Context, code string) (*domain.Entity, error) {
	query := `
		SELECT id, code, name, type, status, latitude, longitude, COALESCE(description, ''), created_at, updated_at
		FROM entities
		WHERE code = $1
	`
	var e domain.Entity
	err := r.pool.QueryRow(ctx, query, code).Scan(
		&e.ID,
		&e.Code,
		&e.Name,
		&e.Type,
		&e.Status,
		&e.Latitude,
		&e.Longitude,
		&e.Description,
		&e.CreatedAt,
		&e.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrEntityNotFound
		}
		return nil, err
	}
	return &e, nil
}

func (r *entityRepository) List(ctx context.Context, filter domain.EntityFilter) ([]domain.Entity, int, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	if filter.Search != "" {
		searchTerm := "%" + strings.TrimSpace(filter.Search) + "%"
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR code ILIKE $%d)", argIdx, argIdx))
		args = append(args, searchTerm)
		argIdx++
	}

	if filter.Type != nil {
		conditions = append(conditions, fmt.Sprintf("type = $%d", argIdx))
		args = append(args, *filter.Type)
		argIdx++
	}

	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *filter.Status)
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// 1. Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM entities %s", whereClause)
	var totalRecords int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&totalRecords); err != nil {
		return nil, 0, fmt.Errorf("failed to count entities: %w", err)
	}

	// 2. Sorting whitelist
	sortBy := "created_at"
	switch filter.SortBy {
	case "name", "code", "status", "created_at":
		sortBy = filter.SortBy
	}

	order := "DESC"
	if strings.ToLower(filter.Order) == "asc" {
		order = "ASC"
	}

	// 3. Pagination calculation
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

	// 4. Data query
	dataQuery := fmt.Sprintf(`
		SELECT id, code, name, type, status, latitude, longitude, COALESCE(description, ''), created_at, updated_at
		FROM entities
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, sortBy, order, argIdx, argIdx+1)

	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query entities: %w", err)
	}
	defer rows.Close()

	entities := make([]domain.Entity, 0, limit)
	for rows.Next() {
		var e domain.Entity
		if err := rows.Scan(
			&e.ID,
			&e.Code,
			&e.Name,
			&e.Type,
			&e.Status,
			&e.Latitude,
			&e.Longitude,
			&e.Description,
			&e.CreatedAt,
			&e.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan entity row: %w", err)
		}
		entities = append(entities, e)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("entity rows error: %w", err)
	}

	return entities, totalRecords, nil
}

func (r *entityRepository) Update(ctx context.Context, entity *domain.Entity) error {
	query := `
		UPDATE entities
		SET code = $1, name = $2, type = $3, status = $4, latitude = $5, longitude = $6, description = $7, updated_at = $8
		WHERE id = $9
	`
	cmdTag, err := r.pool.Exec(
		ctx,
		query,
		entity.Code,
		entity.Name,
		entity.Type,
		entity.Status,
		entity.Latitude,
		entity.Longitude,
		entity.Description,
		entity.UpdatedAt,
		entity.ID,
	)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrEntityNotFound
	}
	return nil
}

func (r *entityRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM entities WHERE id = $1`
	cmdTag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrEntityNotFound
	}
	return nil
}


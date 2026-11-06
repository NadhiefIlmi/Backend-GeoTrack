package usecase

import (
	"context"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"geotrack/backend/internal/domain"
)

type activityLogUsecase struct {
	repo domain.ActivityLogRepository
}

// NewActivityLogUsecase creates a new ActivityLog business logic usecase instance
func NewActivityLogUsecase(repo domain.ActivityLogRepository) domain.ActivityLogUsecase {
	return &activityLogUsecase{repo: repo}
}

func (u *activityLogUsecase) RecordActivity(
	ctx context.Context,
	entityID string,
	entityCode string,
	entityName string,
	action domain.ActivityAction,
	details string,
	actor string,
) error {
	if strings.TrimSpace(actor) == "" {
		actor = "System"
	}

	logEntry := &domain.ActivityLog{
		ID:         uuid.New().String(),
		EntityID:   entityID,
		EntityCode: entityCode,
		EntityName: entityName,
		Action:     action,
		Details:    details,
		Actor:      actor,
		CreatedAt:  time.Now().UTC(),
	}

	return u.repo.Create(ctx, logEntry)
}

func (u *activityLogUsecase) List(ctx context.Context, filter domain.ActivityFilter) ([]domain.ActivityLog, domain.PaginationMeta, error) {
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

	filter.Limit = limit
	filter.Page = page

	logs, totalRecords, err := u.repo.List(ctx, filter)
	if err != nil {
		return nil, domain.PaginationMeta{}, err
	}

	totalPages := 0
	if totalRecords > 0 {
		totalPages = int(math.Ceil(float64(totalRecords) / float64(limit)))
	}

	meta := domain.PaginationMeta{
		Page:         page,
		Limit:        limit,
		TotalRecords: totalRecords,
		TotalPages:   totalPages,
	}

	return logs, meta, nil
}

func (u *activityLogUsecase) GetByID(ctx context.Context, id string) (*domain.ActivityLog, error) {
	return u.repo.GetByID(ctx, id)
}


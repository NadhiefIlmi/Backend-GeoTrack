package usecase

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"geotrack/backend/internal/domain"
	"geotrack/backend/pkg/logger"
)

type entityUsecase struct {
	repo         domain.EntityRepository
	activityRepo domain.ActivityLogRepository
}

// NewEntityUsecase creates a new entity business logic usecase instance
func NewEntityUsecase(repo domain.EntityRepository, activityRepo domain.ActivityLogRepository) domain.EntityUsecase {
	return &entityUsecase{
		repo:         repo,
		activityRepo: activityRepo,
	}
}

func (u *entityUsecase) Create(ctx context.Context, req domain.CreateEntityRequest) (*domain.Entity, error) {
	// 1. Validate coordinates range
	if req.Latitude < -90.0 || req.Latitude > 90.0 || req.Longitude < -180.0 || req.Longitude > 180.0 {
		return nil, domain.ErrInvalidCoordinates
	}

	// 2. Check for unique code conflict
	existing, err := u.repo.GetByCode(ctx, req.Code)
	if err == nil && existing != nil {
		return nil, domain.ErrDuplicateCode
	} else if err != nil && !errors.Is(err, domain.ErrEntityNotFound) {
		return nil, fmt.Errorf("failed to verify code uniqueness: %w", err)
	}

	// 3. Set default status if empty
	status := req.Status
	if status == "" {
		status = domain.StatusActive
	}

	now := time.Now().UTC()
	entity := &domain.Entity{
		ID:          uuid.New().String(),
		Code:        req.Code,
		Name:        req.Name,
		Type:        req.Type,
		Status:      status,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		Description: req.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// 4. Save entity into repository
	if err := u.repo.Create(ctx, entity); err != nil {
		return nil, fmt.Errorf("failed to create entity: %w", err)
	}

	// 5. Asynchronously / Best-effort record activity log
	if u.activityRepo != nil {
		logEntry := &domain.ActivityLog{
			ID:         uuid.New().String(),
			EntityID:   entity.ID,
			EntityCode: entity.Code,
			EntityName: entity.Name,
			Action:     domain.ActionCreated,
			Details:    fmt.Sprintf("Entity %s (%s) registered into system", entity.Name, entity.Code),
			Actor:      "System",
			CreatedAt:  now,
		}
		if err := u.activityRepo.Create(ctx, logEntry); err != nil {
			logger.Warn().Err(err).Str("entity_id", entity.ID).Msg("Failed to record activity log on create")
		}
	}

	return entity, nil
}

func (u *entityUsecase) GetByID(ctx context.Context, id string) (*domain.Entity, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *entityUsecase) List(ctx context.Context, filter domain.EntityFilter) ([]domain.Entity, domain.PaginationMeta, error) {
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

	entities, totalRecords, err := u.repo.List(ctx, filter)
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

	return entities, meta, nil
}

func (u *entityUsecase) Update(ctx context.Context, id string, req domain.UpdateEntityRequest) (*domain.Entity, error) {
	// 1. Validate coordinates range
	if req.Latitude < -90.0 || req.Latitude > 90.0 || req.Longitude < -180.0 || req.Longitude > 180.0 {
		return nil, domain.ErrInvalidCoordinates
	}

	// 2. Fetch existing entity
	existing, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 3. Verify code conflict if code is changed
	if existing.Code != req.Code {
		conflict, err := u.repo.GetByCode(ctx, req.Code)
		if err == nil && conflict != nil {
			return nil, domain.ErrDuplicateCode
		} else if err != nil && !errors.Is(err, domain.ErrEntityNotFound) {
			return nil, fmt.Errorf("failed to verify code uniqueness: %w", err)
		}
	}

	// 4. Track field changes before updating
	typeChanged := existing.Type != req.Type
	prevType := existing.Type
	statusChanged := existing.Status != req.Status
	prevStatus := existing.Status
	nameChanged := existing.Name != req.Name
	prevName := existing.Name
	coordsChanged := existing.Latitude != req.Latitude || existing.Longitude != req.Longitude

	now := time.Now().UTC()
	existing.Code = req.Code
	existing.Name = req.Name
	existing.Type = req.Type
	existing.Status = req.Status
	existing.Latitude = req.Latitude
	existing.Longitude = req.Longitude
	existing.Description = req.Description
	existing.UpdatedAt = now

	// 5. Update in repository
	if err := u.repo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("failed to update entity: %w", err)
	}

	// 6. Record activity audit log
	if u.activityRepo != nil {
		action := domain.ActionUpdated
		var detailParts []string

		if typeChanged {
			detailParts = append(detailParts, fmt.Sprintf("Type changed from %s to %s", prevType, existing.Type))
		}
		if statusChanged {
			action = domain.ActionStatusChanged
			detailParts = append(detailParts, fmt.Sprintf("Status changed from %s to %s", prevStatus, existing.Status))
		}
		if nameChanged {
			detailParts = append(detailParts, fmt.Sprintf("Name changed from \"%s\" to \"%s\"", prevName, existing.Name))
		}
		if coordsChanged {
			detailParts = append(detailParts, fmt.Sprintf("Coordinates updated to (%.4f, %.4f)", existing.Latitude, existing.Longitude))
		}

		var details string
		if len(detailParts) > 0 {
			details = strings.Join(detailParts, "; ")
		} else {
			details = fmt.Sprintf("Entity %s attributes updated", existing.Code)
		}

		logEntry := &domain.ActivityLog{
			ID:         uuid.New().String(),
			EntityID:   existing.ID,
			EntityCode: existing.Code,
			EntityName: existing.Name,
			Action:     action,
			Details:    details,
			Actor:      "System",
			CreatedAt:  now,
		}
		if err := u.activityRepo.Create(ctx, logEntry); err != nil {
			logger.Warn().Err(err).Str("entity_id", existing.ID).Msg("Failed to record activity log on update")
		}
	}

	return existing, nil
}

func (u *entityUsecase) Delete(ctx context.Context, id string) error {
	existing, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := u.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete entity: %w", err)
	}

	if u.activityRepo != nil {
		logEntry := &domain.ActivityLog{
			ID:         uuid.New().String(),
			EntityID:   existing.ID,
			EntityCode: existing.Code,
			EntityName: existing.Name,
			Action:     domain.ActionDeleted,
			Details:    fmt.Sprintf("Entity %s (%s) was deleted", existing.Name, existing.Code),
			Actor:      "System",
			CreatedAt:  time.Now().UTC(),
		}
		if err := u.activityRepo.Create(ctx, logEntry); err != nil {
			logger.Warn().Err(err).Str("entity_id", existing.ID).Msg("Failed to record activity log on delete")
		}
	}

	return nil
}


package domain

import (
	"context"
	"time"
)

type ActivityAction string

const (
	ActionCreated       ActivityAction = "created"
	ActionUpdated       ActivityAction = "updated"
	ActionStatusChanged ActivityAction = "status_changed"
	ActionDeleted       ActivityAction = "deleted"
)

// ActivityLog represents an audit log entry
type ActivityLog struct {
	ID         string         `json:"id"`
	EntityID   string         `json:"entity_id"`
	EntityCode string         `json:"entity_code"`
	EntityName string         `json:"entity_name"`
	Action     ActivityAction `json:"action"`
	Details    string         `json:"details"`
	Actor      string         `json:"actor"`
	CreatedAt  time.Time      `json:"created_at"`
}

// ActivityFilter defines filter criteria for activity queries
type ActivityFilter struct {
	Search   string
	Action   *ActivityAction
	EntityID string
	Page     int
	Limit    int
}

// ActivityLogRepository defines database contract for audit logs
type ActivityLogRepository interface {
	Create(ctx context.Context, log *ActivityLog) error
	List(ctx context.Context, filter ActivityFilter) ([]ActivityLog, int, error)
	GetByID(ctx context.Context, id string) (*ActivityLog, error)
	GetByEntityID(ctx context.Context, entityID string, limit int) ([]ActivityLog, error)
}

// ActivityLogUsecase defines business logic contract for audit logs
type ActivityLogUsecase interface {
	RecordActivity(ctx context.Context, entityID string, entityCode string, entityName string, action ActivityAction, details string, actor string) error
	List(ctx context.Context, filter ActivityFilter) ([]ActivityLog, PaginationMeta, error)
	GetByID(ctx context.Context, id string) (*ActivityLog, error)
}

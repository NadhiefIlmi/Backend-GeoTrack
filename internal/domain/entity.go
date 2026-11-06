package domain

import (
	"context"
	"time"
)

type EntityType string

const (
	TypeVehicle        EntityType = "vehicle"
	TypeIoTDevice      EntityType = "iot_device"
	TypeFacility       EntityType = "facility"
	TypeInfrastructure EntityType = "infrastructure"
)

type EntityStatus string

const (
	StatusActive      EntityStatus = "active"
	StatusInactive    EntityStatus = "inactive"
	StatusMaintenance EntityStatus = "maintenance"
)

// Entity represents a physical geospatial entity
type Entity struct {
	ID          string       `json:"id"`
	Code        string       `json:"code"`
	Name        string       `json:"name"`
	Type        EntityType   `json:"type"`
	Status      EntityStatus `json:"status"`
	Latitude    float64      `json:"latitude"`
	Longitude   float64      `json:"longitude"`
	Description string       `json:"description,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// CreateEntityRequest defines payload for creating a new entity
type CreateEntityRequest struct {
	Code        string       `json:"code" validate:"required,min=1,max=30"`
	Name        string       `json:"name" validate:"required,min=3,max=100"`
	Type        EntityType   `json:"type" validate:"required,oneof=vehicle iot_device facility infrastructure"`
	Status      EntityStatus `json:"status" validate:"omitempty,oneof=active inactive maintenance"`
	Latitude    float64      `json:"latitude" validate:"required,gte=-90,lte=90"`
	Longitude   float64      `json:"longitude" validate:"required,gte=-180,lte=180"`
	Description string       `json:"description" validate:"omitempty,max=1000"`
}

// UpdateEntityRequest defines payload for updating an existing entity
type UpdateEntityRequest struct {
	Code        string       `json:"code" validate:"required,min=1,max=30"`
	Name        string       `json:"name" validate:"required,min=3,max=100"`
	Type        EntityType   `json:"type" validate:"required,oneof=vehicle iot_device facility infrastructure"`
	Status      EntityStatus `json:"status" validate:"required,oneof=active inactive maintenance"`
	Latitude    float64      `json:"latitude" validate:"required,gte=-90,lte=90"`
	Longitude   float64      `json:"longitude" validate:"required,gte=-180,lte=180"`
	Description string       `json:"description" validate:"omitempty,max=1000"`
}

// EntityFilter defines filtering, sorting, and pagination parameters
type EntityFilter struct {
	Search string
	Type   *EntityType
	Status *EntityStatus
	SortBy string // "created_at", "name", "code", "status"
	Order  string // "asc", "desc"
	Page   int
	Limit  int
}

// PaginationMeta carries pagination calculation
type PaginationMeta struct {
	Page         int `json:"page"`
	Limit        int `json:"limit"`
	TotalRecords int `json:"total_records"`
	TotalPages   int `json:"total_pages"`
}

// EntityRepository defines the database access contract for entities
type EntityRepository interface {
	Create(ctx context.Context, entity *Entity) error
	GetByID(ctx context.Context, id string) (*Entity, error)
	GetByCode(ctx context.Context, code string) (*Entity, error)
	List(ctx context.Context, filter EntityFilter) ([]Entity, int, error)
	Update(ctx context.Context, entity *Entity) error
	Delete(ctx context.Context, id string) error
}

// EntityUsecase defines the business logic contract for entities
type EntityUsecase interface {
	Create(ctx context.Context, req CreateEntityRequest) (*Entity, error)
	GetByID(ctx context.Context, id string) (*Entity, error)
	List(ctx context.Context, filter EntityFilter) ([]Entity, PaginationMeta, error)
	Update(ctx context.Context, id string, req UpdateEntityRequest) (*Entity, error)
	Delete(ctx context.Context, id string) error
}


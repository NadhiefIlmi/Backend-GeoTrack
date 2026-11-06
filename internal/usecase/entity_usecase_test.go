package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"geotrack/backend/internal/domain"
	"geotrack/backend/internal/usecase"
	"geotrack/backend/tests/mocks"
)

func TestCreateEntity_Success(t *testing.T) {
	mockRepo := new(mocks.EntityRepositoryMock)
	mockActivity := new(mocks.ActivityLogRepositoryMock)
	uc := usecase.NewEntityUsecase(mockRepo, mockActivity)

	req := domain.CreateEntityRequest{
		Code:        "VH-001",
		Name:        "Fleet Alpha - Truck 01",
		Type:        domain.TypeVehicle,
		Status:      domain.StatusActive,
		Latitude:    -6.2088,
		Longitude:   106.8456,
		Description: "West Jakarta sector",
	}

	mockRepo.On("GetByCode", mock.Anything, "VH-001").Return(nil, domain.ErrEntityNotFound)
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(e *domain.Entity) bool {
		return e.Code == "VH-001" && e.Name == req.Name && e.Status == domain.StatusActive
	})).Return(nil)

	mockActivity.On("Create", mock.Anything, mock.MatchedBy(func(l *domain.ActivityLog) bool {
		return l.Action == domain.ActionCreated && l.EntityCode == "VH-001"
	})).Return(nil)

	res, err := uc.Create(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "VH-001", res.Code)
	assert.Equal(t, "Fleet Alpha - Truck 01", res.Name)
	mockRepo.AssertExpectations(t)
	mockActivity.AssertExpectations(t)
}

func TestCreateEntity_DuplicateCode(t *testing.T) {
	mockRepo := new(mocks.EntityRepositoryMock)
	mockActivity := new(mocks.ActivityLogRepositoryMock)
	uc := usecase.NewEntityUsecase(mockRepo, mockActivity)

	req := domain.CreateEntityRequest{
		Code:      "VH-001",
		Name:      "Fleet Alpha",
		Type:      domain.TypeVehicle,
		Latitude:  -6.2088,
		Longitude: 106.8456,
	}

	existing := &domain.Entity{ID: "existing-id", Code: "VH-001"}
	mockRepo.On("GetByCode", mock.Anything, "VH-001").Return(existing, nil)

	res, err := uc.Create(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, domain.ErrDuplicateCode, err)
}

func TestCreateEntity_InvalidCoordinates(t *testing.T) {
	mockRepo := new(mocks.EntityRepositoryMock)
	uc := usecase.NewEntityUsecase(mockRepo, nil)

	req := domain.CreateEntityRequest{
		Code:      "VH-001",
		Name:      "Fleet Alpha",
		Type:      domain.TypeVehicle,
		Latitude:  120.0, // Invalid latitude > 90
		Longitude: 106.8456,
	}

	res, err := uc.Create(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, domain.ErrInvalidCoordinates, err)
}

func TestGetByID_Success(t *testing.T) {
	mockRepo := new(mocks.EntityRepositoryMock)
	uc := usecase.NewEntityUsecase(mockRepo, nil)

	expected := &domain.Entity{
		ID:        "entity-123",
		Code:      "VH-001",
		Name:      "Fleet Alpha",
		Type:      domain.TypeVehicle,
		Status:    domain.StatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockRepo.On("GetByID", mock.Anything, "entity-123").Return(expected, nil)

	res, err := uc.GetByID(context.Background(), "entity-123")

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "VH-001", res.Code)
}

func TestGetByID_NotFound(t *testing.T) {
	mockRepo := new(mocks.EntityRepositoryMock)
	uc := usecase.NewEntityUsecase(mockRepo, nil)

	mockRepo.On("GetByID", mock.Anything, "non-existent").Return(nil, domain.ErrEntityNotFound)

	res, err := uc.GetByID(context.Background(), "non-existent")

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, domain.ErrEntityNotFound, err)
}

func TestList_Success(t *testing.T) {
	mockRepo := new(mocks.EntityRepositoryMock)
	uc := usecase.NewEntityUsecase(mockRepo, nil)

	filter := domain.EntityFilter{
		Page:  1,
		Limit: 10,
	}

	mockEntities := []domain.Entity{
		{ID: "1", Code: "VH-001", Name: "Truck 01"},
		{ID: "2", Code: "IOT-101", Name: "Sensor 01"},
	}

	mockRepo.On("List", mock.Anything, filter).Return(mockEntities, 2, nil)

	entities, meta, err := uc.List(context.Background(), filter)

	assert.NoError(t, err)
	assert.Len(t, entities, 2)
	assert.Equal(t, 2, meta.TotalRecords)
	assert.Equal(t, 1, meta.TotalPages)
}

func TestUpdateEntity_StatusChanged(t *testing.T) {
	mockRepo := new(mocks.EntityRepositoryMock)
	mockActivity := new(mocks.ActivityLogRepositoryMock)
	uc := usecase.NewEntityUsecase(mockRepo, mockActivity)

	existing := &domain.Entity{
		ID:        "entity-123",
		Code:      "VH-001",
		Name:      "Fleet Alpha",
		Type:      domain.TypeVehicle,
		Status:    domain.StatusActive,
		Latitude:  -6.2088,
		Longitude: 106.8456,
	}

	req := domain.UpdateEntityRequest{
		Code:      "VH-001",
		Name:      "Fleet Alpha",
		Type:      domain.TypeVehicle,
		Status:    domain.StatusMaintenance, // Status changed from active to maintenance
		Latitude:  -6.2088,
		Longitude: 106.8456,
	}

	mockRepo.On("GetByID", mock.Anything, "entity-123").Return(existing, nil)
	mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(e *domain.Entity) bool {
		return e.Status == domain.StatusMaintenance
	})).Return(nil)

	mockActivity.On("Create", mock.Anything, mock.MatchedBy(func(l *domain.ActivityLog) bool {
		return l.Action == domain.ActionStatusChanged
	})).Return(nil)

	res, err := uc.Update(context.Background(), "entity-123", req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, domain.StatusMaintenance, res.Status)
	mockRepo.AssertExpectations(t)
	mockActivity.AssertExpectations(t)
}

func TestDeleteEntity_Success(t *testing.T) {
	mockRepo := new(mocks.EntityRepositoryMock)
	mockActivity := new(mocks.ActivityLogRepositoryMock)
	uc := usecase.NewEntityUsecase(mockRepo, mockActivity)

	existing := &domain.Entity{
		ID:   "entity-123",
		Code: "VH-001",
		Name: "Fleet Alpha",
	}

	mockRepo.On("GetByID", mock.Anything, "entity-123").Return(existing, nil)
	mockRepo.On("Delete", mock.Anything, "entity-123").Return(nil)
	mockActivity.On("Create", mock.Anything, mock.MatchedBy(func(l *domain.ActivityLog) bool {
		return l.Action == domain.ActionDeleted
	})).Return(nil)

	err := uc.Delete(context.Background(), "entity-123")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockActivity.AssertExpectations(t)
}

func TestUpdateEntity_InvalidCoordinates(t *testing.T) {
	mockRepo := new(mocks.EntityRepositoryMock)
	uc := usecase.NewEntityUsecase(mockRepo, nil)

	req := domain.UpdateEntityRequest{
		Latitude:  120.0,
		Longitude: 100.0,
	}

	res, err := uc.Update(context.Background(), "id-1", req)
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, domain.ErrInvalidCoordinates, err)
}

func TestUpdateEntity_NotFound(t *testing.T) {
	mockRepo := new(mocks.EntityRepositoryMock)
	uc := usecase.NewEntityUsecase(mockRepo, nil)

	mockRepo.On("GetByID", mock.Anything, "non-existent").Return(nil, domain.ErrEntityNotFound)

	req := domain.UpdateEntityRequest{
		Code:      "VH-001",
		Latitude:  -6.2,
		Longitude: 106.8,
	}

	res, err := uc.Update(context.Background(), "non-existent", req)
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, domain.ErrEntityNotFound, err)
}

func TestUpdateEntity_DuplicateCode(t *testing.T) {
	mockRepo := new(mocks.EntityRepositoryMock)
	uc := usecase.NewEntityUsecase(mockRepo, nil)

	existing := &domain.Entity{ID: "e-1", Code: "VH-001"}
	conflicting := &domain.Entity{ID: "e-2", Code: "VH-002"}

	mockRepo.On("GetByID", mock.Anything, "e-1").Return(existing, nil)
	mockRepo.On("GetByCode", mock.Anything, "VH-002").Return(conflicting, nil)

	req := domain.UpdateEntityRequest{
		Code:      "VH-002", // Changed code conflicts with e-2
		Latitude:  -6.2,
		Longitude: 106.8,
	}

	res, err := uc.Update(context.Background(), "e-1", req)
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, domain.ErrDuplicateCode, err)
}

func TestDeleteEntity_NotFound(t *testing.T) {
	mockRepo := new(mocks.EntityRepositoryMock)
	uc := usecase.NewEntityUsecase(mockRepo, nil)

	mockRepo.On("GetByID", mock.Anything, "non-existent").Return(nil, domain.ErrEntityNotFound)

	err := uc.Delete(context.Background(), "non-existent")
	assert.Error(t, err)
	assert.Equal(t, domain.ErrEntityNotFound, err)
}


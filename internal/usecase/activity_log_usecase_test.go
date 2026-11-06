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

func TestRecordActivity_Success(t *testing.T) {
	mockRepo := new(mocks.ActivityLogRepositoryMock)
	uc := usecase.NewActivityLogUsecase(mockRepo)

	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(l *domain.ActivityLog) bool {
		return l.EntityID == "e-1" &&
			l.EntityCode == "VH-001" &&
			l.EntityName == "Truck 01" &&
			l.Action == domain.ActionCreated &&
			l.Actor == "System"
	})).Return(nil)

	err := uc.RecordActivity(context.Background(), "e-1", "VH-001", "Truck 01", domain.ActionCreated, "Entity created", "")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestListActivities_Success(t *testing.T) {
	mockRepo := new(mocks.ActivityLogRepositoryMock)
	uc := usecase.NewActivityLogUsecase(mockRepo)

	filter := domain.ActivityFilter{
		Page:  1,
		Limit: 10,
	}

	expectedLogs := []domain.ActivityLog{
		{
			ID:         "act-1",
			EntityID:   "e-1",
			EntityCode: "VH-001",
			EntityName: "Truck 01",
			Action:     domain.ActionStatusChanged,
			Details:    "active -> maintenance",
			Actor:      "Admin Dispatcher",
			CreatedAt:  time.Now(),
		},
	}

	mockRepo.On("List", mock.Anything, filter).Return(expectedLogs, 1, nil)

	logs, meta, err := uc.List(context.Background(), filter)

	assert.NoError(t, err)
	assert.Len(t, logs, 1)
	assert.Equal(t, 1, meta.TotalRecords)
	assert.Equal(t, 1, meta.TotalPages)
}

func TestGetActivityByID_Success(t *testing.T) {
	mockRepo := new(mocks.ActivityLogRepositoryMock)
	uc := usecase.NewActivityLogUsecase(mockRepo)

	expected := &domain.ActivityLog{
		ID:         "act-123",
		EntityID:   "e-1",
		EntityCode: "VH-001",
		Action:     domain.ActionCreated,
		CreatedAt:  time.Now(),
	}

	mockRepo.On("GetByID", mock.Anything, "act-123").Return(expected, nil)

	res, err := uc.GetByID(context.Background(), "act-123")

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "act-123", res.ID)
}

func TestGetActivityByID_NotFound(t *testing.T) {
	mockRepo := new(mocks.ActivityLogRepositoryMock)
	uc := usecase.NewActivityLogUsecase(mockRepo)

	mockRepo.On("GetByID", mock.Anything, "non-existent").Return(nil, domain.ErrActivityNotFound)

	res, err := uc.GetByID(context.Background(), "non-existent")

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, domain.ErrActivityNotFound, err)
}


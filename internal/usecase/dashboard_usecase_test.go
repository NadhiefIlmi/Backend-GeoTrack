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

func TestGetStats_Success(t *testing.T) {
	mockDashRepo := new(mocks.DashboardRepositoryMock)
	mockActRepo := new(mocks.ActivityLogRepositoryMock)
	uc := usecase.NewDashboardUsecase(mockDashRepo, mockActRepo)

	expectedStatus := domain.StatusBreakdown{Active: 5, Inactive: 1, Maintenance: 2}
	expectedTypes := domain.TypeBreakdown{Vehicles: 3, IoTDevices: 2, Facilities: 2, Infrastructures: 1}

	mockDashRepo.On("GetAggregateCounts", mock.Anything).Return(8, expectedStatus, expectedTypes, nil)

	mockLogs := []domain.ActivityLog{
		{
			ID:         "act-1",
			EntityCode: "VH-001",
			Action:     domain.ActionStatusChanged,
			CreatedAt:  time.Now(),
		},
	}
	mockActRepo.On("List", mock.Anything, domain.ActivityFilter{Page: 1, Limit: 5}).
		Return(mockLogs, 1, nil)

	stats, err := uc.GetStats(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, 8, stats.TotalEntities)
	assert.Equal(t, 5, stats.StatusSummary.Active)
	assert.Equal(t, 3, stats.TypeSummary.Vehicles)
	assert.Len(t, stats.RecentActivities, 1)

	mockDashRepo.AssertExpectations(t)
	mockActRepo.AssertExpectations(t)
}


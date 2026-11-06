package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"geotrack/backend/internal/domain"
	"geotrack/backend/internal/usecase"
	"geotrack/backend/pkg/validator"
	"geotrack/backend/tests/mocks"
)

func BenchmarkDashboardStats(b *testing.B) {
	mockDashRepo := new(mocks.DashboardRepositoryMock)
	mockActRepo := new(mocks.ActivityLogRepositoryMock)
	uc := usecase.NewDashboardUsecase(mockDashRepo, mockActRepo)

	expectedStatus := domain.StatusBreakdown{Active: 50, Inactive: 10, Maintenance: 20}
	expectedTypes := domain.TypeBreakdown{Vehicles: 30, IoTDevices: 20, Facilities: 20, Infrastructures: 10}

	mockDashRepo.On("GetAggregateCounts", mock.Anything).Return(80, expectedStatus, expectedTypes, nil)
	mockActRepo.On("List", mock.Anything, mock.Anything).Return([]domain.ActivityLog{}, 0, nil)

	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = uc.GetStats(ctx)
	}
}

func BenchmarkEntityValidation(b *testing.B) {
	req := domain.CreateEntityRequest{
		Code:        "VH-001",
		Name:        "Fleet Alpha - Heavy Truck 01",
		Type:        domain.TypeVehicle,
		Status:      domain.StatusActive,
		Latitude:    -6.2088,
		Longitude:   106.8456,
		Description: "Logistics asset in West Jakarta depot",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.ValidateStruct(req)
	}
}

func BenchmarkEntityFilterList(b *testing.B) {
	mockRepo := new(mocks.EntityRepositoryMock)
	uc := usecase.NewEntityUsecase(mockRepo, nil)

	filter := domain.EntityFilter{
		Page:  1,
		Limit: 10,
	}

	entities := make([]domain.Entity, 10)
	for i := 0; i < 10; i++ {
		entities[i] = domain.Entity{
			ID:        "uuid",
			Code:      "VH-001",
			Name:      "Truck",
			Type:      domain.TypeVehicle,
			Status:    domain.StatusActive,
			CreatedAt: time.Now(),
		}
	}

	mockRepo.On("List", mock.Anything, filter).Return(entities, 100, nil)

	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = uc.List(ctx, filter)
	}
}


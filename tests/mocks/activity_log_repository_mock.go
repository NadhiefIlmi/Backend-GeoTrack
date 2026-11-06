package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"geotrack/backend/internal/domain"
)

type ActivityLogRepositoryMock struct {
	mock.Mock
}

func (m *ActivityLogRepositoryMock) Create(ctx context.Context, log *domain.ActivityLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *ActivityLogRepositoryMock) List(ctx context.Context, filter domain.ActivityFilter) ([]domain.ActivityLog, int, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]domain.ActivityLog), args.Int(1), args.Error(2)
}

func (m *ActivityLogRepositoryMock) GetByID(ctx context.Context, id string) (*domain.ActivityLog, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ActivityLog), args.Error(1)
}

func (m *ActivityLogRepositoryMock) GetByEntityID(ctx context.Context, entityID string, limit int) ([]domain.ActivityLog, error) {
	args := m.Called(ctx, entityID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.ActivityLog), args.Error(1)
}

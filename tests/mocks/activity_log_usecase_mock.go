package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"geotrack/backend/internal/domain"
)

type ActivityLogUsecaseMock struct {
	mock.Mock
}

func (m *ActivityLogUsecaseMock) RecordActivity(
	ctx context.Context,
	entityID string,
	entityCode string,
	entityName string,
	action domain.ActivityAction,
	details string,
	actor string,
) error {
	args := m.Called(ctx, entityID, entityCode, entityName, action, details, actor)
	return args.Error(0)
}

func (m *ActivityLogUsecaseMock) List(ctx context.Context, filter domain.ActivityFilter) ([]domain.ActivityLog, domain.PaginationMeta, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Get(1).(domain.PaginationMeta), args.Error(2)
	}
	return args.Get(0).([]domain.ActivityLog), args.Get(1).(domain.PaginationMeta), args.Error(2)
}

func (m *ActivityLogUsecaseMock) GetByID(ctx context.Context, id string) (*domain.ActivityLog, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ActivityLog), args.Error(1)
}


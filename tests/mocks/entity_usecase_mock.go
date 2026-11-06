package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"geotrack/backend/internal/domain"
)

type EntityUsecaseMock struct {
	mock.Mock
}

func (m *EntityUsecaseMock) Create(ctx context.Context, req domain.CreateEntityRequest) (*domain.Entity, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Entity), args.Error(1)
}

func (m *EntityUsecaseMock) GetByID(ctx context.Context, id string) (*domain.Entity, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Entity), args.Error(1)
}

func (m *EntityUsecaseMock) List(ctx context.Context, filter domain.EntityFilter) ([]domain.Entity, domain.PaginationMeta, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Get(1).(domain.PaginationMeta), args.Error(2)
	}
	return args.Get(0).([]domain.Entity), args.Get(1).(domain.PaginationMeta), args.Error(2)
}

func (m *EntityUsecaseMock) Update(ctx context.Context, id string, req domain.UpdateEntityRequest) (*domain.Entity, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Entity), args.Error(1)
}

func (m *EntityUsecaseMock) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}


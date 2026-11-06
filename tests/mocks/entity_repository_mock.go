package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"geotrack/backend/internal/domain"
)

type EntityRepositoryMock struct {
	mock.Mock
}

func (m *EntityRepositoryMock) Create(ctx context.Context, entity *domain.Entity) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *EntityRepositoryMock) GetByID(ctx context.Context, id string) (*domain.Entity, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Entity), args.Error(1)
}

func (m *EntityRepositoryMock) GetByCode(ctx context.Context, code string) (*domain.Entity, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Entity), args.Error(1)
}

func (m *EntityRepositoryMock) List(ctx context.Context, filter domain.EntityFilter) ([]domain.Entity, int, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]domain.Entity), args.Int(1), args.Error(2)
}

func (m *EntityRepositoryMock) Update(ctx context.Context, entity *domain.Entity) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *EntityRepositoryMock) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}


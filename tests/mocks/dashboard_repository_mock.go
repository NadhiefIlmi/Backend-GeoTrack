package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"geotrack/backend/internal/domain"
)

type DashboardRepositoryMock struct {
	mock.Mock
}

func (m *DashboardRepositoryMock) GetAggregateCounts(ctx context.Context) (int, domain.StatusBreakdown, domain.TypeBreakdown, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Get(1).(domain.StatusBreakdown), args.Get(2).(domain.TypeBreakdown), args.Error(3)
}


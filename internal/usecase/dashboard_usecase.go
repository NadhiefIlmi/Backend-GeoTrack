package usecase

import (
	"context"

	"geotrack/backend/internal/domain"
)

type dashboardUsecase struct {
	dashboardRepo domain.DashboardRepository
	activityRepo  domain.ActivityLogRepository
}

// NewDashboardUsecase creates a new Dashboard business logic usecase instance
func NewDashboardUsecase(
	dashboardRepo domain.DashboardRepository,
	activityRepo domain.ActivityLogRepository,
) domain.DashboardUsecase {
	return &dashboardUsecase{
		dashboardRepo: dashboardRepo,
		activityRepo:  activityRepo,
	}
}

func (u *dashboardUsecase) GetStats(ctx context.Context) (*domain.DashboardStats, error) {
	total, statusSummary, typeSummary, err := u.dashboardRepo.GetAggregateCounts(ctx)
	if err != nil {
		return nil, err
	}

	recentActivities := make([]domain.ActivityLog, 0)
	if u.activityRepo != nil {
		filter := domain.ActivityFilter{
			Page:  1,
			Limit: 5,
		}
		logs, _, err := u.activityRepo.List(ctx, filter)
		if err == nil && logs != nil {
			recentActivities = logs
		}
	}

	stats := &domain.DashboardStats{
		TotalEntities:    total,
		StatusSummary:    statusSummary,
		TypeSummary:      typeSummary,
		RecentActivities: recentActivities,
	}

	return stats, nil
}


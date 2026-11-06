package domain

import "context"

type StatusBreakdown struct {
	Active      int `json:"active"`
	Inactive    int `json:"inactive"`
	Maintenance int `json:"maintenance"`
}

type TypeBreakdown struct {
	Vehicles        int `json:"vehicles"`
	IoTDevices      int `json:"iot_devices"`
	Facilities      int `json:"facilities"`
	Infrastructures int `json:"infrastructures"`
}

type DashboardStats struct {
	TotalEntities    int             `json:"total_entities"`
	StatusSummary    StatusBreakdown `json:"status_summary"`
	TypeSummary      TypeBreakdown   `json:"type_summary"`
	RecentActivities []ActivityLog   `json:"recent_activities"`
}

type DashboardRepository interface {
	GetAggregateCounts(ctx context.Context) (total int, status StatusBreakdown, types TypeBreakdown, err error)
}

type DashboardUsecase interface {
	GetStats(ctx context.Context) (*DashboardStats, error)
}


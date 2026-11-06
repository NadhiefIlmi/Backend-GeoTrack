package http

import (
	"net/http"

	"geotrack/backend/internal/domain"
)

type DashboardHandler struct {
	usecase domain.DashboardUsecase
}

// NewDashboardHandler creates a new HTTP delivery handler for dashboard metrics
func NewDashboardHandler(usecase domain.DashboardUsecase) *DashboardHandler {
	return &DashboardHandler{usecase: usecase}
}

func (h *DashboardHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.usecase.GetStats(r.Context())
	if err != nil {
		SendError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve dashboard statistics", nil)
		return
	}

	SendSuccess(w, http.StatusOK, "Dashboard statistics retrieved successfully", stats, nil)
}


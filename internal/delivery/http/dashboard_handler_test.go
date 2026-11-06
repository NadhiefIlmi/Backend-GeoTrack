package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"geotrack/backend/internal/config"
	deliveryHTTP "geotrack/backend/internal/delivery/http"
	"geotrack/backend/internal/domain"
	"geotrack/backend/tests/mocks"
)

func TestDashboardHandler_GetStats_Success(t *testing.T) {
	mockUC := new(mocks.DashboardUsecaseMock)
	cfg := &config.Config{App: config.AppConfig{Env: "test"}}
	handler := deliveryHTTP.NewDashboardHandler(mockUC)
	router := deliveryHTTP.NewRouter(cfg, nil, nil, nil, handler)

	expectedStats := &domain.DashboardStats{
		TotalEntities: 8,
		StatusSummary: domain.StatusBreakdown{Active: 5, Inactive: 1, Maintenance: 2},
		TypeSummary:   domain.TypeBreakdown{Vehicles: 3, IoTDevices: 2, Facilities: 2, Infrastructures: 1},
	}

	mockUC.On("GetStats", mock.Anything).Return(expectedStats, nil)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/dashboard/stats", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var res deliveryHTTP.SuccessResponse
	err := json.Unmarshal(rec.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.True(t, res.Success)
	assert.Equal(t, "Dashboard statistics retrieved successfully", res.Message)

	dataMap, ok := res.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(8), dataMap["total_entities"])
}


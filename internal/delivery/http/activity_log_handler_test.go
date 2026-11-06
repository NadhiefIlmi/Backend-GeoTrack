package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"geotrack/backend/internal/config"
	deliveryHTTP "geotrack/backend/internal/delivery/http"
	"geotrack/backend/internal/domain"
	"geotrack/backend/tests/mocks"
)

func TestActivityLogHandler_List(t *testing.T) {
	mockUC := new(mocks.ActivityLogUsecaseMock)
	cfg := &config.Config{App: config.AppConfig{Env: "test"}}
	handler := deliveryHTTP.NewActivityLogHandler(mockUC)
	router := deliveryHTTP.NewRouter(cfg, nil, nil, handler, nil)

	expectedLogs := []domain.ActivityLog{
		{
			ID:         "act-1",
			EntityID:   "e-1",
			EntityCode: "VH-001",
			EntityName: "Truck 01",
			Action:     domain.ActionCreated,
			Details:    "Entity created",
			Actor:      "System",
			CreatedAt:  time.Now(),
		},
	}
	expectedMeta := domain.PaginationMeta{
		Page:         1,
		Limit:        10,
		TotalRecords: 1,
		TotalPages:   1,
	}

	mockUC.On("List", mock.Anything, mock.AnythingOfType("domain.ActivityFilter")).
		Return(expectedLogs, expectedMeta, nil)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/activities?page=1&limit=10", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var res deliveryHTTP.SuccessResponse
	err := json.Unmarshal(rec.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.True(t, res.Success)
	assert.NotNil(t, res.Meta)
	assert.Equal(t, 1, res.Meta.TotalRecords)
}

func TestActivityLogHandler_GetByID_Success(t *testing.T) {
	mockUC := new(mocks.ActivityLogUsecaseMock)
	cfg := &config.Config{App: config.AppConfig{Env: "test"}}
	handler := deliveryHTTP.NewActivityLogHandler(mockUC)
	router := deliveryHTTP.NewRouter(cfg, nil, nil, handler, nil)

	expected := &domain.ActivityLog{
		ID:         "act-123",
		EntityID:   "e-1",
		EntityCode: "VH-001",
		EntityName: "Truck 01",
		Action:     domain.ActionStatusChanged,
		Details:    "Status changed to maintenance",
		Actor:      "Admin Dispatcher",
		CreatedAt:  time.Now(),
	}

	mockUC.On("GetByID", mock.Anything, "act-123").Return(expected, nil)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/activities/act-123", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var res deliveryHTTP.SuccessResponse
	err := json.Unmarshal(rec.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.True(t, res.Success)
}

func TestActivityLogHandler_GetByID_NotFound(t *testing.T) {
	mockUC := new(mocks.ActivityLogUsecaseMock)
	cfg := &config.Config{App: config.AppConfig{Env: "test"}}
	handler := deliveryHTTP.NewActivityLogHandler(mockUC)
	router := deliveryHTTP.NewRouter(cfg, nil, nil, handler, nil)

	mockUC.On("GetByID", mock.Anything, "non-existent").Return(nil, domain.ErrActivityNotFound)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/activities/non-existent", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var res deliveryHTTP.ErrorResponse
	err := json.Unmarshal(rec.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.False(t, res.Success)
	assert.Equal(t, "ACTIVITY_NOT_FOUND", res.Error.Code)
}


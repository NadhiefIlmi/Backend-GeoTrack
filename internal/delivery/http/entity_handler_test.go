package http_test

import (
	"bytes"
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

func TestEntityHandler_List(t *testing.T) {
	mockUC := new(mocks.EntityUsecaseMock)
	cfg := &config.Config{App: config.AppConfig{Env: "test"}}
	handler := deliveryHTTP.NewEntityHandler(mockUC)
	router := deliveryHTTP.NewRouter(cfg, nil, handler, nil, nil)

	expectedEntities := []domain.Entity{
		{ID: "e-1", Code: "VH-001", Name: "Truck 01", Type: domain.TypeVehicle, Status: domain.StatusActive},
	}
	expectedMeta := domain.PaginationMeta{
		Page:         1,
		Limit:        10,
		TotalRecords: 1,
		TotalPages:   1,
	}

	mockUC.On("List", mock.Anything, mock.AnythingOfType("domain.EntityFilter")).
		Return(expectedEntities, expectedMeta, nil)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/entities?page=1&limit=10", nil)
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

func TestEntityHandler_GetByID_Success(t *testing.T) {
	mockUC := new(mocks.EntityUsecaseMock)
	cfg := &config.Config{App: config.AppConfig{Env: "test"}}
	handler := deliveryHTTP.NewEntityHandler(mockUC)
	router := deliveryHTTP.NewRouter(cfg, nil, handler, nil, nil)

	expected := &domain.Entity{
		ID:        "e-123",
		Code:      "VH-001",
		Name:      "Truck 01",
		Type:      domain.TypeVehicle,
		Status:    domain.StatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockUC.On("GetByID", mock.Anything, "e-123").Return(expected, nil)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/entities/e-123", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var res deliveryHTTP.SuccessResponse
	err := json.Unmarshal(rec.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.True(t, res.Success)
}

func TestEntityHandler_GetByID_NotFound(t *testing.T) {
	mockUC := new(mocks.EntityUsecaseMock)
	cfg := &config.Config{App: config.AppConfig{Env: "test"}}
	handler := deliveryHTTP.NewEntityHandler(mockUC)
	router := deliveryHTTP.NewRouter(cfg, nil, handler, nil, nil)

	mockUC.On("GetByID", mock.Anything, "non-existent").Return(nil, domain.ErrEntityNotFound)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/entities/non-existent", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var res deliveryHTTP.ErrorResponse
	err := json.Unmarshal(rec.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.False(t, res.Success)
	assert.Equal(t, "ENTITY_NOT_FOUND", res.Error.Code)
}

func TestEntityHandler_Create_Success(t *testing.T) {
	mockUC := new(mocks.EntityUsecaseMock)
	cfg := &config.Config{App: config.AppConfig{Env: "test"}}
	handler := deliveryHTTP.NewEntityHandler(mockUC)
	router := deliveryHTTP.NewRouter(cfg, nil, handler, nil, nil)

	payload := domain.CreateEntityRequest{
		Code:      "VH-001",
		Name:      "Truck 01",
		Type:      domain.TypeVehicle,
		Status:      domain.StatusActive,
		Latitude:  -6.2088,
		Longitude: 106.8456,
	}
	body, _ := json.Marshal(payload)

	createdEntity := &domain.Entity{
		ID:        "new-uuid",
		Code:      "VH-001",
		Name:      "Truck 01",
		Type:      domain.TypeVehicle,
		Status:    domain.StatusActive,
		Latitude:  -6.2088,
		Longitude: 106.8456,
	}

	mockUC.On("Create", mock.Anything, payload).Return(createdEntity, nil)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/entities", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var res deliveryHTTP.SuccessResponse
	err := json.Unmarshal(rec.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.True(t, res.Success)
	assert.Equal(t, "Entity created successfully", res.Message)
}

func TestEntityHandler_Create_ValidationError(t *testing.T) {
	mockUC := new(mocks.EntityUsecaseMock)
	cfg := &config.Config{App: config.AppConfig{Env: "test"}}
	handler := deliveryHTTP.NewEntityHandler(mockUC)
	router := deliveryHTTP.NewRouter(cfg, nil, handler, nil, nil)

	// Missing code and invalid latitude > 90
	payload := domain.CreateEntityRequest{
		Code:      "",
		Name:      "No Code",
		Type:      domain.TypeVehicle,
		Latitude:  150.0,
		Longitude: 100.0,
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/entities", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)

	var res deliveryHTTP.ErrorResponse
	err := json.Unmarshal(rec.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.False(t, res.Success)
	assert.Equal(t, "VALIDATION_ERROR", res.Error.Code)
}

func TestEntityHandler_Create_DuplicateCode(t *testing.T) {
	mockUC := new(mocks.EntityUsecaseMock)
	cfg := &config.Config{App: config.AppConfig{Env: "test"}}
	handler := deliveryHTTP.NewEntityHandler(mockUC)
	router := deliveryHTTP.NewRouter(cfg, nil, handler, nil, nil)

	payload := domain.CreateEntityRequest{
		Code:      "VH-001",
		Name:      "Truck 01",
		Type:      domain.TypeVehicle,
		Status:    domain.StatusActive,
		Latitude:  -6.2088,
		Longitude: 106.8456,
	}
	body, _ := json.Marshal(payload)

	mockUC.On("Create", mock.Anything, payload).Return(nil, domain.ErrDuplicateCode)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/entities", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)

	var res deliveryHTTP.ErrorResponse
	err := json.Unmarshal(rec.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.False(t, res.Success)
	assert.Equal(t, "DUPLICATE_CODE", res.Error.Code)
}

func TestEntityHandler_Delete_Success(t *testing.T) {
	mockUC := new(mocks.EntityUsecaseMock)
	cfg := &config.Config{App: config.AppConfig{Env: "test"}}
	handler := deliveryHTTP.NewEntityHandler(mockUC)
	router := deliveryHTTP.NewRouter(cfg, nil, handler, nil, nil)

	mockUC.On("Delete", mock.Anything, "e-123").Return(nil)

	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/entities/e-123", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var res deliveryHTTP.SuccessResponse
	err := json.Unmarshal(rec.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.True(t, res.Success)
	assert.Equal(t, "Entity deleted successfully", res.Message)
}


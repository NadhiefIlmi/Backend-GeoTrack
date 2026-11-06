package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"geotrack/backend/internal/config"
	deliveryHTTP "geotrack/backend/internal/delivery/http"
)

func TestHealthCheckHandler(t *testing.T) {
	handler := deliveryHTTP.NewHealthHandler("test", "1.0.0", nil)

	req, err := http.NewRequest(http.MethodGet, "/api/v1/health", nil)
	assert.NoError(t, err)

	rec := httptest.NewRecorder()
	handler.Check(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var res deliveryHTTP.SuccessResponse
	err = json.Unmarshal(rec.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.True(t, res.Success)
	assert.Equal(t, "Service is healthy and operational", res.Message)

	dataMap, ok := res.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "UP", dataMap["status"])
	assert.Equal(t, "1.0.0", dataMap["version"])
	assert.Equal(t, "test", dataMap["environment"])
	assert.NotEmpty(t, dataMap["timestamp"])
}

func TestHealthCheckRouter(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{
			Env:            "test",
			AllowedOrigins: []string{"http://localhost:8081"},
		},
	}
	handler := deliveryHTTP.NewHealthHandler(cfg.App.Env, "1.0.0", nil)
	router := deliveryHTTP.NewRouter(cfg, handler, nil, nil, nil)

	req, err := http.NewRequest(http.MethodGet, "/api/v1/health", nil)
	assert.NoError(t, err)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
}


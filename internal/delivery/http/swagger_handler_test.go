package http_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	deliveryHTTP "geotrack/backend/internal/delivery/http"
)

func TestSwaggerRoutes(t *testing.T) {
	// Create temporary spec file
	tmpDir := t.TempDir()
	specFile := filepath.Join(tmpDir, "swagger.json")
	err := os.WriteFile(specFile, []byte(`{"openapi":"3.0.3","info":{"title":"Test API"}}`), 0644)
	assert.NoError(t, err)

	r := chi.NewRouter()
	deliveryHTTP.RegisterSwaggerRoutes(r, specFile)

	// Test 1: Fetch spec JSON
	reqJSON := httptest.NewRequest(http.MethodGet, "/swagger/doc.json", nil)
	recJSON := httptest.NewRecorder()
	r.ServeHTTP(recJSON, reqJSON)
	assert.Equal(t, http.StatusOK, recJSON.Code)
	assert.Equal(t, "application/json", recJSON.Header().Get("Content-Type"))
	assert.Contains(t, recJSON.Body.String(), "Test API")

	// Test 2: Fetch Swagger UI HTML
	reqHTML := httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil)
	recHTML := httptest.NewRecorder()
	r.ServeHTTP(recHTML, reqHTML)
	assert.Equal(t, http.StatusOK, recHTML.Code)
	assert.Contains(t, recHTML.Header().Get("Content-Type"), "text/html")
	assert.Contains(t, recHTML.Body.String(), "SwaggerUIBundle")

	// Test 3: Redirect from /swagger/
	reqRedirect := httptest.NewRequest(http.MethodGet, "/swagger/", nil)
	recRedirect := httptest.NewRecorder()
	r.ServeHTTP(recRedirect, reqRedirect)
	assert.Equal(t, http.StatusMovedPermanently, recRedirect.Code)
	assert.Equal(t, "/swagger/index.html", recRedirect.Header().Get("Location"))
}


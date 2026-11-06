package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"geotrack/backend/internal/config"
)

func TestLoadConfig_Defaults(t *testing.T) {
	cfg, err := config.LoadConfig("")
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	// Check default values
	assert.Equal(t, "GeoTrack Backend", cfg.App.Name)
	assert.Equal(t, 8080, cfg.App.Port)
	assert.Equal(t, "development", cfg.App.Env)
	assert.Contains(t, cfg.App.AllowedOrigins, "http://localhost:8081")
	assert.Contains(t, cfg.App.AllowedOrigins, "http://localhost:5173")

	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "geotrack_db", cfg.Database.Name)
	assert.Equal(t, 25, cfg.Database.MaxConns)
	assert.Equal(t, 5, cfg.Database.MinConns)
}

func TestLoadConfig_EnvOverride(t *testing.T) {
	os.Setenv("APP_PORT", "9090")
	os.Setenv("APP_ENV", "production")
	defer func() {
		os.Unsetenv("APP_PORT")
		os.Unsetenv("APP_ENV")
	}()

	cfg, err := config.LoadConfig("")
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, 9090, cfg.App.Port)
	assert.Equal(t, "production", cfg.App.Env)
}


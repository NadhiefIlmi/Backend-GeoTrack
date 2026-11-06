package logger_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"geotrack/backend/pkg/logger"
)

func TestLoggerInitialization(t *testing.T) {
	// Test dev init
	logger.Init("development")
	assert.NotNil(t, logger.Log)

	// Test logging calls do not panic
	assert.NotPanics(t, func() {
		logger.Info().Msg("Test info message")
		logger.Debug().Msg("Test debug message")
		logger.Warn().Msg("Test warn message")
		logger.Error().Msg("Test error message")
	})

	// Test prod init
	logger.Init("production")
	assert.NotNil(t, logger.Log)
}


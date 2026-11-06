package validator_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"geotrack/backend/internal/domain"
	"geotrack/backend/pkg/validator"
)

func TestValidateStruct_Valid(t *testing.T) {
	req := domain.CreateEntityRequest{
		Code:      "VH-001",
		Name:      "Valid Name",
		Type:      domain.TypeVehicle,
		Status:    domain.StatusActive,
		Latitude:  -6.2088,
		Longitude: 106.8456,
	}

	errs := validator.ValidateStruct(req)
	assert.Empty(t, errs)
}

func TestValidateStruct_Invalid(t *testing.T) {
	req := domain.CreateEntityRequest{
		Code:      "", // required
		Name:      "a", // min 3
		Type:      "invalid_type", // oneof
		Latitude:  100.0, // max 90
		Longitude: 200.0, // max 180
	}

	errs := validator.ValidateStruct(req)
	assert.NotEmpty(t, errs)
	assert.True(t, len(errs) >= 4)
}


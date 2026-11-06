package domain

import "errors"

var (
	ErrEntityNotFound      = errors.New("entity not found")
	ErrDuplicateCode       = errors.New("entity code already in use")
	ErrInvalidCoordinates  = errors.New("invalid coordinates: latitude must be [-90, 90] and longitude [-180, 180]")
	ErrInvalidEntityType   = errors.New("invalid entity type")
	ErrInvalidEntityStatus = errors.New("invalid entity status")
	ErrValidationFailed    = errors.New("input validation failed")
	ErrActivityNotFound    = errors.New("activity log not found")
)


package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// FieldError represents structured field validation failure
type FieldError struct {
	Field string `json:"field"`
	Issue string `json:"issue"`
}

// ValidateStruct validates a struct using tags and returns clean slice of FieldError
func ValidateStruct(s interface{}) []FieldError {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	var errors []FieldError
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			field := strings.ToLower(e.Field())
			var issue string

			switch e.Tag() {
			case "required":
				issue = fmt.Sprintf("%s is required", field)
			case "min":
				issue = fmt.Sprintf("%s must be at least %s characters", field, e.Param())
			case "max":
				issue = fmt.Sprintf("%s must be at most %s characters", field, e.Param())
			case "gte":
				issue = fmt.Sprintf("%s must be greater than or equal to %s", field, e.Param())
			case "lte":
				issue = fmt.Sprintf("%s must be less than or equal to %s", field, e.Param())
			case "oneof":
				issue = fmt.Sprintf("%s must be one of: %s", field, e.Param())
			default:
				issue = fmt.Sprintf("%s failed validation on '%s'", field, e.Tag())
			}

			errors = append(errors, FieldError{
				Field: field,
				Issue: issue,
			})
		}
	} else {
		errors = append(errors, FieldError{
			Field: "payload",
			Issue: err.Error(),
		})
	}

	return errors
}


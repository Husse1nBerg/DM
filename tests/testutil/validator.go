package testutil

import (
	"github.com/go-playground/validator/v10"
)

// TestValidator is a simple validator implementation for tests
type TestValidator struct{}

// Validate implements echo.Validator interface
func (v *TestValidator) Validate(i interface{}) error {
	if i == nil {
		return nil
	}
	validate := validator.New()
	return validate.Struct(i)
}

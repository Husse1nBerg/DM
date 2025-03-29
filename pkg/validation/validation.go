package validation

import (
	"bytes"
)

type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value,omitempty"`
	Message string `json:"message"`
}

// ValidationErrors is a collection of validation errors
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

func (ve ValidationErrors) Error() string {
	if len(ve.Errors) == 0 {
		return "validation failed"
	}

	var buf bytes.Buffer
	for _, err := range ve.Errors {
		buf.WriteString(err.Field)
		buf.WriteString(": ")
		buf.WriteString(err.Message)
		buf.WriteString("\n")
	}
	return buf.String()
}

// GetValidationMessage returns a user-friendly message for validation tags
func GetValidationMessage(tag string) string {
	switch tag {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		return "Value is less than minimum allowed"
	case "max":
		return "Value is more than maximum allowed"
	default:
		return "Validation failed for " + tag
	}
}

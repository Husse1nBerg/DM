package responses

import (
	"fmt"
	"net/http"

	"github.com/dockworks/dm-web-backend/pkg/errors"
	"github.com/dockworks/dm-web-backend/pkg/validation"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// BaseResponse is the foundation for all API responses
// @Description Standard response structure for all API endpoints
type BaseResponse struct {
	Code        int         `json:"-"`
	Pretty      bool        `json:"-"`
	Data        interface{} `json:"data,omitempty"`
	Message     interface{} `json:"message,omitempty"`
	Error       interface{} `json:"error,omitempty"`
	Details     interface{} `json:"details,omitempty"`
	Total       int64       `json:"total,omitempty"`
	PerPage     int32       `json:"perPage,omitempty"`
	CurrentPage int32       `json:"currentPage,omitempty"`
	LastPage    int32       `json:"lastPage,omitempty"`
}

// ValidationError represents a field-specific validation error
// @Description Specific validation error for a single field
type ValidationError struct {
	Field   string `json:"field" example:"username"`
	Tag     string `json:"tag" example:"required"`
	Value   string `json:"value" example:"invalid_value"`
	Message string `json:"message" example:"Username is required"`
}

// Error represents an error response with optional validation details
// @Description Error response structure with optional validation details
type Error struct {
	Message string            `json:"message" example:"Validation failed"`
	Details []ValidationError `json:"details,omitempty"`
}

func (r BaseResponse) JSON(ctx echo.Context) error {
	if r.Message == "" && r.Message == nil && r.Error == nil {
		r.Message = http.StatusText(r.Code)
	}

	if err, ok := r.Error.(error); ok {
		if errors.Is(err, errors.DatabaseInternalError) {
			r.Code = http.StatusInternalServerError
		}

		if errors.Is(err, errors.DatabaseRecordNotFound) {
			r.Code = http.StatusNotFound
		}

		// Handle structured validation errors
		if validationErrs, ok := err.(validation.ValidationErrors); ok {
			r.Code = http.StatusBadRequest
			r.Error = "Validation failed"
			r.Details = validationErrs.Errors
		} else {
			r.Error = err.Error()
		}
	}

	if r.Pretty {
		return ctx.JSONPretty(r.Code, r, "\t")
	}

	return ctx.JSON(r.Code, r)
}

func NewSuccessResponse(data interface{}) BaseResponse {
	return BaseResponse{
		Code: http.StatusOK,
		Data: data,
	}
}

func NewErrorResponse(code int, error interface{}) BaseResponse {
	if validationErrs, ok := error.(validation.ValidationErrors); ok {
		return BaseResponse{
			Code:    http.StatusBadRequest,
			Error:   "Validation failed",
			Details: validationErrs.Errors,
		}
	}

	if valErrs, ok := error.(validator.ValidationErrors); ok {
		validationErrors := make([]validation.ValidationError, 0, len(valErrs))
		for _, err := range valErrs {
			validationErrors = append(validationErrors, validation.ValidationError{
				Field:   err.Field(),
				Tag:     err.Tag(),
				Value:   fmt.Sprintf("%v", err.Value()),
				Message: validation.GetValidationMessage(err.Tag()),
			})
		}

		return BaseResponse{
			Code:    http.StatusBadRequest,
			Error:   "Validation failed",
			Details: validationErrors,
		}
	}

	return BaseResponse{
		Code:  code,
		Error: error,
	}
}

func NewErrorResponseWithDetails(code int, error interface{}, details interface{}) BaseResponse {
	return BaseResponse{
		Code:    code,
		Error:   error,
		Details: details,
	}
}

func NewMessageResponse(code int, message string) BaseResponse {
	return BaseResponse{
		Code:    code,
		Message: message,
	}
}

func NewPaginatedResponse(data interface{}, total int64, perPage, currentPage int32) BaseResponse {
	var lastPage int32 = 1
	if total > 0 && perPage > 0 {
		lastPage = int32((total + int64(perPage) - 1) / int64(perPage))
	}

	return BaseResponse{
		Code:        http.StatusOK,
		Data:        data,
		Total:       total,
		PerPage:     perPage,
		CurrentPage: currentPage,
		LastPage:    lastPage,
	}
}

package responses

import (
	"github.com/labstack/echo/v4"
)

// SadieResponse is the response format for SADIE endpoints
// @Description SADIE API response structure
type SadieResponse struct {
	Success     bool                   `json:"success"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Description string                 `json:"description,omitempty"`
	Steps       []string               `json:"steps,omitempty"`
}

func (r SadieResponse) JSON(ctx echo.Context) error {
	return ctx.JSON(200, r)
}

// NewSadieSuccessResponse creates a successful SADIE response
func NewSadieSuccessResponse(data map[string]interface{}, description string, steps []string) SadieResponse {
	return SadieResponse{
		Success:     true,
		Data:        data,
		Description: description,
		Steps:       steps,
	}
}

// NewSadieErrorResponse creates an error SADIE response
func NewSadieErrorResponse(error string, description string, steps []string, data map[string]interface{}) SadieResponse {
	if data == nil {
		data = make(map[string]interface{})
	}
	data["error"] = error
	return SadieResponse{
		Success:     false,
		Data:        data,
		Description: description,
		Steps:       steps,
	}
}


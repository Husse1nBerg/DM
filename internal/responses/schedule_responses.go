package responses

import (
	"github.com/dockworks/dm-web-backend/pkg/dme"
)

// ScheduleLabelResponse represents a schedule label
// @Description Schedule label with color and description
type ScheduleLabelResponse struct {
	ID          int    `json:"id" example:"0"`
	Color       string `json:"color" example:"#FF0000"`
	Description string `json:"description" example:"Urgent"`
}

// ScheduleLabelsResponse represents a list of schedule labels
// @Description List of schedule labels
type ScheduleLabelsResponse struct {
	Labels []ScheduleLabelResponse `json:"labels"`
}

// ScheduleResponse represents the raw schedule payload returned by DME
// @Description Raw DME schedule payload; structure varies per endpoint
type ScheduleResponse map[string]interface{}

// NewScheduleLabelResponse creates a new ScheduleLabelResponse from a DME ScheduleLabel
func NewScheduleLabelResponse(label dme.ScheduleLabel) ScheduleLabelResponse {
	return ScheduleLabelResponse{
		ID:          label.ID,
		Color:       label.Color,
		Description: label.Description,
	}
}

// NewScheduleLabelsResponse creates a new ScheduleLabelsResponse from a slice of DME ScheduleLabels
func NewScheduleLabelsResponse(labels []dme.ScheduleLabel) ScheduleLabelsResponse {
	labelResponses := make([]ScheduleLabelResponse, len(labels))
	for i, label := range labels {
		labelResponses[i] = NewScheduleLabelResponse(label)
	}
	return ScheduleLabelsResponse{
		Labels: labelResponses,
	}
}

// ScheduleUpdateResponse represents the response from a schedule update operation
// @Description Schedule update response
type ScheduleUpdateResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// ConvertScheduleResponse converts a DME API response to a standardized schedule response
func ConvertScheduleResponse(dmeResponse interface{}) interface{} {
	// Pass through the DME API response as-is
	// The frontend expects the raw structure from the DME API
	return dmeResponse
}

// ConvertScheduleUpdateResponse converts a DME schedule update response
func ConvertScheduleUpdateResponse(dmeResponse interface{}) ScheduleUpdateResponse {
	// Convert the DME response to a standardized update response
	return ScheduleUpdateResponse{
		Success: true,
		Message: "Schedule updated successfully",
		Data:    dmeResponse,
	}
}

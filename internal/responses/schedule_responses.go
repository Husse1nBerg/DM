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

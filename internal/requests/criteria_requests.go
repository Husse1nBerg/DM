package requests

import (
	"encoding/json"
	"fmt"

	"github.com/dockworks/dm-web-backend/pkg/models"
	"github.com/go-playground/validator/v10"
)

// CreateCriteriaRequest represents the request body for creating criteria
type CreateCriteriaRequest struct {
	Name        string      `json:"name" validate:"required,max=255"`
	Description string      `json:"description"`
	Criteria    interface{} `json:"criteria" validate:"required"`
}

// UpdateCriteriaRequest represents the request body for updating criteria
type UpdateCriteriaRequest struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Criteria    interface{} `json:"criteria"`
}

// Validate validates the create criteria request
func (r *CreateCriteriaRequest) Validate() error {
	validate := validator.New()
	if err := validate.Struct(r); err != nil {
		return err
	}

	// Validate the criteria structure
	if err := r.validateCriteria(); err != nil {
		return fmt.Errorf("invalid criteria structure: %w", err)
	}

	return nil
}

// Validate validates the update criteria request
func (r *UpdateCriteriaRequest) Validate() error {
	validate := validator.New()
	if err := validate.Struct(r); err != nil {
		return err
	}

	// Validate the criteria structure
	if err := r.validateCriteria(); err != nil {
		return fmt.Errorf("invalid criteria structure: %w", err)
	}

	return nil
}

// validateCriteria validates the criteria JSON structure
func (r *CreateCriteriaRequest) validateCriteria() error {
	// Convert interface{} to JSON bytes
	criteriaBytes, err := json.Marshal(r.Criteria)
	if err != nil {
		return fmt.Errorf("failed to marshal criteria: %w", err)
	}

	var qb models.QueryBuilder
	if err := json.Unmarshal(criteriaBytes, &qb); err != nil {
		return fmt.Errorf("invalid JSON structure: %w", err)
	}

	return qb.Validate()
}

// validateCriteria validates the criteria JSON structure
func (r *UpdateCriteriaRequest) validateCriteria() error {
	// Convert interface{} to JSON bytes
	criteriaBytes, err := json.Marshal(r.Criteria)
	if err != nil {
		return fmt.Errorf("failed to marshal criteria: %w", err)
	}

	var qb models.QueryBuilder
	if err := json.Unmarshal(criteriaBytes, &qb); err != nil {
		return fmt.Errorf("invalid JSON structure: %w", err)
	}

	return qb.Validate()
}

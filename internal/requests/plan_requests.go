package requests

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// CreateNotesMessagesPlanRequest represents the request to create a new notes and messages plan
type CreateNotesMessagesPlanRequest struct {
	Name          string  `json:"name" validate:"required" example:"Basic Plan"`
	MonthlyPrice  float64 `json:"monthlyPrice" validate:"required,min=0" example:"29.99"`
	TextLimit     *int32  `json:"textLimit,omitempty" example:"1000"`
	EmailLimit    *string `json:"emailLimit,omitempty" example:"Unlimited Emails"`
	UserLimit     *string `json:"userLimit,omitempty" example:"Unlimited Users"`
	IsMostPopular *bool   `json:"isMostPopular,omitempty" example:"false"`
}

// UpdateNotesMessagesPlanRequest represents the request to update an existing notes and messages plan
type UpdateNotesMessagesPlanRequest struct {
	Name          *string  `json:"name,omitempty" example:"Basic Plan"`
	MonthlyPrice  *float64 `json:"monthlyPrice,omitempty" validate:"omitempty,min=0" example:"29.99"`
	TextLimit     *int32   `json:"textLimit,omitempty" example:"1000"`
	EmailLimit    *string  `json:"emailLimit,omitempty" example:"Unlimited Emails"`
	UserLimit     *string  `json:"userLimit,omitempty" example:"Unlimited Users"`
	IsMostPopular *bool    `json:"isMostPopular,omitempty" example:"false"`
}

// CreateStoragePlanRequest represents the request to create a new storage plan
type CreateStoragePlanRequest struct {
	Name           string  `json:"name" validate:"required" example:"Basic Storage"`
	MonthlyPrice   float64 `json:"monthlyPrice" validate:"required,min=0" example:"19.99"`
	StorageLimitGB *int32  `json:"storageLimitGB,omitempty" example:"10"`
	UserLimit      *string `json:"userLimit,omitempty" example:"Unlimited Users"`
	IsMostPopular  *bool   `json:"isMostPopular,omitempty" example:"false"`
}

// UpdateStoragePlanRequest represents the request to update an existing storage plan
type UpdateStoragePlanRequest struct {
	Name           *string  `json:"name,omitempty" example:"Basic Storage"`
	MonthlyPrice   *float64 `json:"monthlyPrice,omitempty" validate:"omitempty,min=0" example:"19.99"`
	StorageLimitGB *int32   `json:"storageLimitGB,omitempty" example:"10"`
	UserLimit      *string  `json:"userLimit,omitempty" example:"Unlimited Users"`
	IsMostPopular  *bool    `json:"isMostPopular,omitempty" example:"false"`
}

// PlanIDParam represents the URL parameter for plan ID
type PlanIDParam struct {
	PlanID uuid.UUID `param:"planId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// Validate performs custom validation on the create notes and messages plan request
func (r *CreateNotesMessagesPlanRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// Validate performs custom validation on the update notes and messages plan request
func (r *UpdateNotesMessagesPlanRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// Validate performs custom validation on the create storage plan request
func (r *CreateStoragePlanRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// Validate performs custom validation on the update storage plan request
func (r *UpdateStoragePlanRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

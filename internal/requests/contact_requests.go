package requests

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// ContactType represents the type of contact
type ContactType string

const (
	ContactTypePhone ContactType = "phone"
	ContactTypeEmail ContactType = "email"
)

// CreateContactRequest represents the request body for creating a contact
type CreateContactRequest struct {
	Type        ContactType `json:"type" validate:"required,oneof=phone email"`
	Name        string      `json:"name" validate:"required"`
	Description string      `json:"description"`
	Email       string      `json:"email" validate:"omitempty,email"`
	Phone       string      `json:"phone"`
	IsCPContact *bool       `json:"is_cp_contact"`
}

// UpdateContactRequest represents the request body for updating a contact
type UpdateContactRequest struct {
	Type        ContactType `json:"type" validate:"required,oneof=phone email"`
	Name        string      `json:"name" validate:"required"`
	Description string      `json:"description"`
	Email       string      `json:"email" validate:"omitempty,email"`
	Phone       string      `json:"phone"`
	IsCPContact *bool       `json:"is_cp_contact"`
}

// ListContactsRequest represents the parameters to list contacts for a marina with filtering, sorting, and pagination
// swagger:parameters ListContacts
type ListContactsRequest struct {
	MarinaID string `json:"marinaId" validate:"required"`
	FilterSortParams
}

// Validate validates the request
func (r *CreateContactRequest) Validate() error {
	validate := validator.New()
	if err := validate.Struct(r); err != nil {
		return err
	}

	// Additional validation based on type
	if r.Type == ContactTypeEmail && r.Email == "" {
		return fmt.Errorf("email is required when type is email")
	}
	if r.Type == ContactTypePhone && r.Phone == "" {
		return fmt.Errorf("phone is required when type is phone")
	}

	return nil
}

// Validate validates the request
func (r *UpdateContactRequest) Validate() error {
	validate := validator.New()
	if err := validate.Struct(r); err != nil {
		return err
	}

	// Additional validation based on type
	if r.Type == ContactTypeEmail && r.Email == "" {
		return fmt.Errorf("email is required when type is email")
	}
	if r.Type == ContactTypePhone && r.Phone == "" {
		return fmt.Errorf("phone is required when type is phone")
	}

	return nil
}

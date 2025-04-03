package requests

// CreateOrganizationRequest defines the parameters for creating a new organization
type CreateOrganizationRequest struct {
	Email    string                `json:"email" validate:"required,email" example:"org@example.com"`
	Name     string                `json:"name" validate:"required" example:"Sample Organization"`
	Image    *string               `json:"image,omitempty" example:"https://example.com/logo.png"`
	Website  *string               `json:"website,omitempty" example:"https://example.com"`
	Country  *string               `json:"country,omitempty" example:"USA"`
	Phone    *string               `json:"phone,omitempty" example:"+1 555-123-4567"`
	IsActive bool                  `json:"is_active" example:"true"`
	IsTest   bool                  `json:"is_test" example:"false"`
	Address  *CreateAddressRequest `json:"address,omitempty"`
}

// UpdateOrganizationRequest defines the parameters for updating an organization
type UpdateOrganizationRequest struct {
	Email    *string `json:"email,omitempty" validate:"omitempty,email" example:"updated@example.com"`
	Name     *string `json:"name,omitempty" example:"Updated Organization"`
	Image    *string `json:"image,omitempty" example:"https://example.com/new-logo.png"`
	Website  *string `json:"website,omitempty" example:"https://updated-example.com"`
	Country  *string `json:"country,omitempty" example:"Canada"`
	Phone    *string `json:"phone,omitempty" example:"+1 555-987-6543"`
	IsActive *bool   `json:"is_active,omitempty" example:"true"`
	IsTest   *bool   `json:"is_test,omitempty" example:"false"`
}

// GetOrganizationsPaginatedRequest defines the pagination parameters for organization list
type GetOrganizationsPaginatedRequest struct {
	Limit  int32 `query:"limit" validate:"omitempty,min=1,max=100" example:"10"`
	Offset int32 `query:"offset" validate:"omitempty,min=0" example:"0"`
}

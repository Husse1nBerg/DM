package requests

// CreateAddressRequest defines the parameters for creating a new address
type CreateAddressRequest struct {
	Street     *string  `json:"street,omitempty" example:"123 Main St"`
	City       *string  `json:"city,omitempty" example:"San Francisco"`
	State      *string  `json:"state,omitempty" example:"CA"`
	PostalCode *string  `json:"postal_code,omitempty" example:"94105"`
	Country    *string  `json:"country,omitempty" example:"USA"`
	Latitude   *float64 `json:"latitude,omitempty" example:"37.7749"`
	Longitude  *float64 `json:"longitude,omitempty" example:"-122.4194"`
}

// UpdateAddressRequest defines the parameters for updating an address
type UpdateAddressRequest struct {
	Street     *string  `json:"street,omitempty" example:"123 Main St"`
	City       *string  `json:"city,omitempty" example:"San Francisco"`
	State      *string  `json:"state,omitempty" example:"CA"`
	PostalCode *string  `json:"postal_code,omitempty" example:"94105"`
	Country    *string  `json:"country,omitempty" example:"USA"`
	Latitude   *float64 `json:"latitude,omitempty" example:"37.7749"`
	Longitude  *float64 `json:"longitude,omitempty" example:"-122.4194"`
}

// UpdateOrgAddressRequest defines a request to update an organization and its address
type UpdateOrgAddressRequest struct {
	Email    *string               `json:"email,omitempty" validate:"omitempty,email" example:"updated@example.com"`
	Name     *string               `json:"name,omitempty" example:"Updated Organization"`
	Image    *string               `json:"image,omitempty" example:"https://example.com/new-logo.png"`
	Website  *string               `json:"website,omitempty" example:"https://updated-example.com"`
	Country  *string               `json:"country,omitempty" example:"Canada"`
	Phone    *string               `json:"phone,omitempty" example:"+1 555-987-6543"`
	IsActive *bool                 `json:"is_active,omitempty" example:"true"`
	IsTest   *bool                 `json:"is_test,omitempty" example:"false"`
	Address  *UpdateAddressRequest `json:"address,omitempty"`
}

package responses

import (
	"github.com/dockworks/dm-web-backend/pkg/dme"
)

// ClerkResponse represents a system clerk (user) in the DME system
// @Description System clerk/user information from DME API
type ClerkResponse struct {
	ID          string `json:"id" example:"CLERK001"`
	Name        string `json:"name" example:"John Doe"`
	FirstName   string `json:"firstName" example:"John"`
	LastName    string `json:"lastName" example:"Doe"`
	Email       string `json:"email" example:"john.doe@marina.com"`
	Phone       string `json:"phone" example:"+15551234567"`
	IsActive    bool   `json:"isActive" example:"true"`
	Department  string `json:"department" example:"Service"`
	Role        string `json:"role" example:"Technician"`
	LastLogin   string `json:"lastLogin" example:"2024-01-15T10:30:00Z"`
	CreatedDate string `json:"createdDate" example:"2023-06-01T08:00:00Z"`
}

// ClerkListResponse represents a list of system clerks
// @Description List of system clerks/users from DME API
type ClerkListResponse struct {
	Clerks []ClerkResponse `json:"clerks"`
}

// LocationResponse represents a business location from DME API
// @Description Business location information from DME API
type LocationResponse struct {
	ID             string `json:"id" example:"LOC001"`
	Name           string `json:"name" example:"Main Marina"`
	LocationNumber string `json:"locationNumber" example:"001"`
	Address1       string `json:"address1" example:"123 Marina Drive"`
	Address2       string `json:"address2" example:"Suite 100"`
	Address3       string `json:"address3" example:""`
	City           string `json:"city" example:"Harbor City"`
	State          string `json:"state" example:"CA"`
	ZipCode        string `json:"zipCode" example:"90210"`
	Country        string `json:"country" example:"USA"`
	Phone          string `json:"phone" example:"+15551234567"`
	Fax            string `json:"fax" example:"+15551234568"`
	BillToAddress1 string `json:"billToAddress1" example:"123 Marina Drive"`
	BillToAddress2 string `json:"billToAddress2" example:"Suite 100"`
	BillToAddress3 string `json:"billToAddress3" example:""`
	BillToCity     string `json:"billToCity" example:"Harbor City"`
	BillToState    string `json:"billToState" example:"CA"`
	BillToZip      string `json:"billToZip" example:"90210"`
	BillToCountry  string `json:"billToCountry" example:"USA"`
	BillToPhone    string `json:"billToPhone" example:"+15551234567"`
	BillToFax      string `json:"billToFax" example:"+15551234568"`
	DMPayClientID  string `json:"dmPayClientId" example:"DMPAY123"`
}

// LocationListResponse represents a list of business locations
// @Description List of business locations from DME API
type LocationListResponse struct {
	Locations []LocationResponse `json:"locations"`
}

// NewClerkResponse creates a new ClerkResponse from a DME Clerk
func NewClerkResponse(clerk dme.Clerk) ClerkResponse {
	return ClerkResponse{
		ID:          clerk.ID,
		Name:        clerk.Name,
		FirstName:   clerk.FirstName,
		LastName:    clerk.LastName,
		Email:       clerk.Email,
		Phone:       clerk.Phone,
		IsActive:    clerk.IsActive,
		Department:  clerk.Department,
		Role:        clerk.Role,
		LastLogin:   clerk.LastLogin,
		CreatedDate: clerk.CreatedDate,
	}
}

// NewClerkListResponse creates a new ClerkListResponse from a slice of DME Clerks
func NewClerkListResponse(clerks []dme.Clerk) ClerkListResponse {
	clerkResponses := make([]ClerkResponse, len(clerks))
	for i, clerk := range clerks {
		clerkResponses[i] = NewClerkResponse(clerk)
	}
	return ClerkListResponse{
		Clerks: clerkResponses,
	}
}

// NewLocationResponse creates a new LocationResponse from a DME Location
func NewLocationResponse(location dme.Location) LocationResponse {
	return LocationResponse{
		ID:             location.ID,
		Name:           location.Name,
		LocationNumber: location.LocationNumber,
		Address1:       location.Address1,
		Address2:       location.Address2,
		Address3:       location.Address3,
		City:           location.City,
		State:          location.State,
		ZipCode:        location.ZipCode,
		Country:        location.Country,
		Phone:          location.Phone,
		Fax:            location.Fax,
		BillToAddress1: location.BillToAddress1,
		BillToAddress2: location.BillToAddress2,
		BillToAddress3: location.BillToAddress3,
		BillToCity:     location.BillToCity,
		BillToState:    location.BillToState,
		BillToZip:      location.BillToZip,
		BillToCountry:  location.BillToCountry,
		BillToPhone:    location.BillToPhone,
		BillToFax:      location.BillToFax,
		DMPayClientID:  location.DMPayClientID,
	}
}

// NewLocationListResponse creates a new LocationListResponse from a slice of DME Locations
func NewLocationListResponse(locations []dme.Location) LocationListResponse {
	locationResponses := make([]LocationResponse, len(locations))
	for i, location := range locations {
		locationResponses[i] = NewLocationResponse(location)
	}
	return LocationListResponse{
		Locations: locationResponses,
	}
}

// DepartmentResponse represents a department from DME API
// @Description Department information from DME API
type DepartmentResponse struct {
	ID          string `json:"id" example:"DEPT001"`
	Description string `json:"description" example:"Service Department"`
	InternalCOS string `json:"internalCOS" example:"INT001"`
	RetailCOS   string `json:"retailCOS" example:"RET001"`
}

// DepartmentListResponse represents a list of departments
// @Description List of departments from DME API
type DepartmentListResponse struct {
	Departments []DepartmentResponse `json:"departments"`
}

// NewDepartmentResponse creates a new DepartmentResponse from a DME Department
func NewDepartmentResponse(department dme.Department) DepartmentResponse {
	return DepartmentResponse{
		ID:          department.ID,
		Description: department.Description,
		InternalCOS: department.InternalCOS,
		RetailCOS:   department.RetailCOS,
	}
}

// NewDepartmentListResponse creates a new DepartmentListResponse from a slice of DME Departments
func NewDepartmentListResponse(departments []dme.Department) DepartmentListResponse {
	departmentResponses := make([]DepartmentResponse, len(departments))
	for i, department := range departments {
		departmentResponses[i] = NewDepartmentResponse(department)
	}
	return DepartmentListResponse{
		Departments: departmentResponses,
	}
}
package requests

// ClerkRequest represents the required parameters to retrieve a specific clerk
type ClerkRequest struct {
	ClerkID  string `json:"clerkId" query:"clerkId" validate:"required" example:"CLERK001"`
	SystemID string `json:"systemId" query:"systemId" validate:"required" example:"SYS001"`
}

// ClerkListRequest represents the required parameters to list all clerks
type ClerkListRequest struct {
	SystemID        string `json:"systemId" query:"systemId" validate:"required" example:"SYS001"`
	IncludeInactive *bool  `json:"includeInactive" query:"includeInactive" example:"false"`
}

// LocationListRequest represents the required parameters to list all locations
type LocationListRequest struct {
	SystemID string `json:"systemId" query:"systemId" validate:"required" example:"SYS001"`
}

package requests

// VendorListRequest represents the request to list all vendors
type VendorListRequest struct {
	SystemID string `json:"systemId" query:"systemId" validate:"required"`
}

// VendorSearchRequest represents the request to search vendors
type VendorSearchRequest struct {
	SystemID     string `json:"systemId" query:"systemId" validate:"required"`
	SearchString string `json:"searchString" query:"searchString" validate:"required,min=1"`
	DirectHit    *bool  `json:"directHit" query:"directHit"`
}

// VendorRetrieveRequest represents the request to retrieve a single vendor
type VendorRetrieveRequest struct {
	SystemID string `json:"systemId" query:"systemId" validate:"required"`
	VendorID string `json:"vendorId" query:"vendorId" validate:"required"`
}


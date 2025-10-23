package requests

import "github.com/dockworks/dm-web-backend/pkg/dme"

// EstimateListRequest represents a request to list estimates by page
type EstimateListRequest struct {
	Page     int `query:"page" validate:"required,min=1"`
	PageSize int `query:"pageSize" validate:"required,min=1,max=100"`
}

// EstimateRetrieveRequest represents a request to retrieve an estimate by ID
type EstimateRetrieveRequest struct {
	Id     string `query:"Id" validate:"required"`
	Detail bool   `query:"Detail" default:"true"`
}

// EstimateSearchRequest represents a request to search for estimates
type EstimateSearchRequest struct {
	SearchString string `query:"SearchString" validate:"required"`
	DirectHit    string `query:"DirectHit" validate:"oneof=true false"`
}

// EstimateCreateRequest represents a request to create a new estimate
type EstimateCreateRequest struct {
	ClerkId         string           `json:"clerkId"`
	CustId          string           `json:"custId" validate:"required"`
	BoatId          string           `json:"boatId"`
	BoatName        string           `json:"boatName"`
	CustomerPhone   string           `json:"customerPhone"`
	CustomerEmail   string           `json:"customerEmail"`
	Comments        string           `json:"comments"`
	LocationCode    string           `json:"locationCode"`
	EstCompDate     string           `json:"estCompDate"`
	EstStartDate    string           `json:"estStartDate"`
	CustPromiseDate string           `json:"custPromiseDate"`
	CategoryCode    string           `json:"categoryCode"`
	Title           string           `json:"title" validate:"required"`
	OperationCodes  []OperationCode  `json:"operationCodes"`
	Attachments     []dme.Attachment `json:"attachments"`
}

// EstimateUpdateRequest represents a request to update an existing estimate
type EstimateUpdateRequest struct {
	EstId           string                  `json:"estId"`
	ClerkId         string                  `json:"clerkId"`
	CustId          string                  `json:"custId"`
	BoatId          string                  `json:"boatId"`
	BoatName        string                  `json:"boatName"`
	CustomerPhone   string                  `json:"customerPhone"`
	CustomerEmail   string                  `json:"customerEmail"`
	Comments        string                  `json:"comments"`
	LocationCode    string                  `json:"locationCode"`
	EstCompDate     string                  `json:"estCompDate"`
	EstStartDate    string                  `json:"estStartDate"`
	CustPromiseDate string                  `json:"custPromiseDate"`
	CategoryCode    string                  `json:"categoryCode"`
	Title           string                  `json:"title"`
	OperationCodes  []OperationCode         `json:"operationCodes"`
	Attachments     []AttachmentWithPublic  `json:"attachments"`
}

// EstimatesForCustomerRequest represents a request to list estimates for a specific customer
type EstimatesForCustomerRequest struct {
	CustId           string `query:"custId" validate:"required"`
	Status           string `query:"status"`
	LocationCodeList string `query:"locationCodeList"`
}

// EstimateDeleteOperationRequest represents a request to delete an operation from an estimate
type EstimateDeleteOperationRequest struct {
	WorkOrder string `query:"WorkOrder" validate:"required"`
	Operation string `query:"Operation" validate:"required"`
}

// EstimateRetrieveListRequest represents a request to retrieve a list of estimates
type EstimateRetrieveListRequest struct {
	LastUpdateDate string   `json:"lastUpdateDate,omitempty"`
	LastUpdateTime string   `json:"lastUpdateTime,omitempty"`
	Status         string   `json:"status,omitempty"`
	WoIds          []string `json:"woIds,omitempty"`
	Detail         bool     `json:"detail"`
	Page           int      `json:"page" validate:"required,min=0"`
	PageSize       int      `json:"pageSize" validate:"required,min=1,max=100"`
}

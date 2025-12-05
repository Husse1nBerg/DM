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
	Status          string                  `json:"status"`
	Type            string                  `json:"type"`
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
	WorkOrder string `json:"workOrder" query:"WorkOrder" validate:"required"`
	Operation string `json:"operation" query:"Operation" validate:"required"`
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

// SubmitEstimateSubletEntryRequest represents a request to submit a sublet entry for an estimate
type SubmitEstimateSubletEntryRequest struct {
	WorkOrderId         string  `json:"workOrderId" validate:"required"`
	OpCode              string  `json:"opCode" validate:"required"`
	VendorId            string  `json:"vendorId" validate:"required"`
	PurchaseDate        string  `json:"purchaseDate" validate:"required"`
	PartsPrice          float64 `json:"partsPrice"`
	PartsCost           float64 `json:"partsCost"`
	LaborPrice          float64 `json:"laborPrice"`
	LaborCost           float64 `json:"laborCost"`
	Description         string  `json:"description,omitempty"`
	SubletDiscount      float64 `json:"subletDiscount"`
	SubletLaborDiscount float64 `json:"subletLaborDiscount"`
	LocationCode        string  `json:"locationCode,omitempty"`
	Department          string  `json:"department,omitempty"`
}

// SubmitEstimatePartEntryRequest represents a request to submit a part entry for an estimate
type SubmitEstimatePartEntryRequest struct {
	EstimateId   string  `json:"estimateId" validate:"required"`
	OpCode       string  `json:"opCode" validate:"required"`
	PartNumber   string  `json:"partNumber" validate:"required"`
	Quantity     float64 `json:"quantity" validate:"required,min=0.01"`
	UnitPrice    float64 `json:"unitPrice"`
	Description  string  `json:"description,omitempty"`
	LocationCode string  `json:"locationCode,omitempty"`
}

// SubmitEstimateLaborEntryRequest represents a request to submit a labor entry for an estimate
type SubmitEstimateLaborEntryRequest struct {
	EstimateId        string  `json:"estimateId" validate:"required"`
	OpCode            string  `json:"opCode" validate:"required"`
	TechId            string  `json:"techId" validate:"required"`
	Date              string  `json:"date" validate:"required"`
	StartTime         string  `json:"startTime,omitempty"`
	StopTime          string  `json:"stopTime,omitempty"`
	Hours             float64 `json:"hours,omitempty"`
	Comments          string  `json:"comments,omitempty"`
	Department        string  `json:"department,omitempty"`
	IsApproved        *bool   `json:"isApproved,omitempty"`
	FlagLaborFinished *bool   `json:"flagLaborFinished,omitempty"`
}

// RetrieveEstimatePartsRequest represents a request to retrieve parts for an estimate
type RetrieveEstimatePartsRequest struct {
	EstimatesId string `query:"estimatesId" validate:"required"`
	Opcode      string `query:"opcode,omitempty"`
}

// RetrieveEstimateLaborRequest represents a request to retrieve labor entries for an estimate
type RetrieveEstimateLaborRequest struct {
	EstimatesId string `query:"estimatesId" validate:"required"`
	Opcode      string `query:"opcode,omitempty"`
}
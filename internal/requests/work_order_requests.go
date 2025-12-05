package requests

import "github.com/dockworks/dm-web-backend/pkg/dme"

// WorkOrderListRequest represents a request to list work orders by page
type WorkOrderListRequest struct {
	Page     int `query:"page" validate:"required,min=1"`
	PageSize int `query:"pageSize" validate:"required,min=1,max=100"`
}

// WorkOrderRetrieveRequest represents a request to retrieve a work order by ID
type WorkOrderRetrieveRequest struct {
	Id     string `query:"Id" validate:"required"`
	Detail bool   `query:"Detail" default:"true"`
}

// WorkOrderSearchRequest represents a request to search for work orders
type WorkOrderSearchRequest struct {
	SearchString string `query:"SearchString" validate:"required"`
	DirectHit    string `query:"DirectHit" validate:"oneof=true false"`
}

// OperationCode represents an operation code within a work order
type OperationCode struct {
	Opcode                 string  `json:"opcode"`
	Desc                   string  `json:"desc"`
	LongDesc               string  `json:"longDesc"`
	TechDesc               string  `json:"techDesc"`
	CategoryCode           string  `json:"categoryCode"`
	EstimatedParts         float64 `json:"estimatedParts"`
	EstimatedLabor         float64 `json:"estimatedLabor"`
	EstimatedLaborHours    float64 `json:"estimatedLaborHours"`
	EstimatedFreight       float64 `json:"estimatedFreight"`
	EstimatedEquipment     float64 `json:"estimatedEquipment"`
	EstimatedSublet        float64 `json:"estimatedSublet"`
	EstimatedMileage       float64 `json:"estimatedMileage"`
	EstimatedMiscSupply    float64 `json:"estimatedMiscSupply"`
	EstimatedBillCodes     float64 `json:"estimatedBillCodes"`
	Approved               bool    `json:"approved"`
	FlatRateAmount         float64 `json:"flatRateAmount"`
	FlatRatePerFootRate    float64 `json:"flatRatePerFootRate"`
	FlatRatePerFootMethod  string  `json:"flatRatePerFootMethod"`
	LaborFinished          bool    `json:"laborFinished"`
	StandardHours          float64 `json:"standardHours"`
	EstCompDate            string  `json:"estCompDate"`
	EstStartDate           string  `json:"estStartDate"`
	CustPromiseDate        string  `json:"custPromiseDate"`
	ReqCompDate            string  `json:"reqCompDate"`
	ForecastedPartsCharges float64 `json:"forecastedPartsCharges"`
	ForecastedLaborCharges float64 `json:"forecastedLaborCharges"`
	ForecastedLaborHours   float64 `json:"forecastedLaborHours"`
}

// WorkOrderCreateRequest represents a request to create a new work order
type WorkOrderCreateRequest struct {
	WoId            string           `json:"woId"`
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

// WorkOrderUpdateRequest represents a request to update an existing work order
type WorkOrderUpdateRequest struct {
	WoId            string                  `json:"woId" validate:"required"`
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

// WorkOrdersForCustomerRequest represents a request to list work orders for a specific customer
type WorkOrdersForCustomerRequest struct {
	CustId           string `query:"custId" validate:"required"`
	Status           string `query:"status"`
	LocationCodeList string `query:"locationCodeList"`
}

// WorkOrderCompletedRequest represents a request to retrieve completed work orders
type WorkOrderCompletedRequest struct {
	CompleteDate string `query:"CompleteDate" validate:"required"`
}

// WorkOrderCreateFromEstimateRequest represents a request to create a work order from an existing estimate
type WorkOrderCreateFromEstimateRequest struct {
	EstimateId        string `json:"EstimateId" validate:"required"`
	WithDetail        bool   `json:"WithDetail" default:"false"`
	WithUnapprovedOps bool   `json:"WithUnapprovedOps" default:"false"`
}

// WorkOrderDeleteOperationRequest represents a request to delete an operation from a work order
type WorkOrderDeleteOperationRequest struct {
	WorkOrder string `json:"workOrder" query:"WorkOrder" validate:"required"`
	Operation string `json:"operation" query:"Operation" validate:"required"`
}

// WorkOrderListNewOrChangedRequest represents a request to list new or changed work orders
type WorkOrderListNewOrChangedRequest struct {
	AsOfDate string `query:"AsOfDate" validate:"required"`
	Page     int    `query:"page" validate:"required,min=1"`
	PageSize int    `query:"pageSize" validate:"required,min=1,max=100"`
}

// WorkOrderRetrieveListRequest represents a request to retrieve a list of work orders
type WorkOrderRetrieveListRequest struct {
	// Add fields as needed based on API requirements
	WithDetail bool `json:"withDetail"`
	Status     string `json:"status,omitempty"`
}

// WorkOrderPartEntryRequest represents a request to submit a part entry
type WorkOrderPartEntryRequest struct {
	WorkOrderID string  `json:"workOrderId" validate:"required"`
	PartNumber  string  `json:"partNumber" validate:"required"`
	Quantity    int     `json:"quantity" validate:"required,min=1"`
	UnitPrice   float64 `json:"unitPrice" validate:"required,min=0"`
	Description string  `json:"description"`
	OperationID string  `json:"operationId"`
}

// WorkOrderTimeEntryRequest represents a request to submit a time entry
type WorkOrderTimeEntryRequest struct {
	TechnicianID       string `json:"techId" validate:"required"`
	WorkOrderID        string `json:"workOrderId" validate:"required"`
	OperationID        string `json:"opCode" validate:"required"`
	Date               string `json:"date" validate:"required"`
	StartTime          string `json:"startTime" validate:"required"`
	StopTime           string `json:"stopTime" validate:"required"`
	IsApproved         *bool  `json:"isApproved"`
	Comments           string `json:"comments"`
	TimeEntryUID       string `json:"timeEntryUId,omitempty"`
	FlagLaborFinished  *bool  `json:"flagLaborFinished"`
}

// WorkOrderListTimeEntryRequest represents a request to list time entries
type WorkOrderListTimeEntryRequest struct {
	AsOfDate string `query:"AsOfDate" validate:"required"`
	Page     int    `query:"page" validate:"required,min=1"`
	PageSize int    `query:"pageSize" validate:"required,min=1,max=100"`
}

// RetrieveAllOperationsRequest represents a request to retrieve all operation codes with pagination and filters
type RetrieveAllOperationsRequest struct {
	Page         int    `json:"page" validate:"required,min=0"`
	PageSize     int    `json:"pageSize" validate:"required,min=1,max=1000"`
	OpCode       string `json:"opCode,omitempty"`
	CategoryCode string `json:"categoryCode,omitempty"`
	Desc         string `json:"desc,omitempty"`
}

// SearchAllOperationsRequest represents a request to search for operation codes
type SearchAllOperationsRequest struct {
	SearchString string `json:"searchString" validate:"required"`
	DirectHit    bool   `json:"directHit"`
}

// SubmitWorkOrderSubletEntryRequest represents a request to submit a sublet entry for a work order
type SubmitWorkOrderSubletEntryRequest struct {
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

// RetrieveWorkOrderLaborDetailRequest represents a request to retrieve labor detail for a work order
type RetrieveWorkOrderLaborDetailRequest struct {
	WorkOrderId string `query:"workOrderId" validate:"required"`
	Opcode      string `query:"opcode,omitempty"`
}

// RetrieveWorkOrderPartsRequest represents a request to retrieve parts for a work order
type RetrieveWorkOrderPartsRequest struct {
	WorkOrderId string `query:"workOrderId" validate:"required"`
	Opcode      string `query:"opcode,omitempty"`
}
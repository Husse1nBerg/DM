package requests

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
	ForecastedPartsCharges float64 `json:"forecastedPartsCharges"`
	ForecastedLaborCharges float64 `json:"forecastedLaborCharges"`
	ForecastedLaborHours   float64 `json:"forecastedLaborHours"`
}

// WorkOrderCreateRequest represents a request to create a new work order
type WorkOrderCreateRequest struct {
	WoId            string          `json:"woId"`
	ClerkId         string          `json:"clerkId"`
	CustId          string          `json:"custId" validate:"required"`
	BoatId          string          `json:"boatId"`
	BoatName        string          `json:"boatName"`
	CustomerPhone   string          `json:"customerPhone"`
	CustomerEmail   string          `json:"customerEmail"`
	Comments        string          `json:"comments"`
	LocationCode    string          `json:"locationCode"`
	EstCompDate     string          `json:"estCompDate"`
	EstStartDate    string          `json:"estStartDate"`
	CustPromiseDate string          `json:"custPromiseDate"`
	CategoryCode    string          `json:"categoryCode"`
	Title           string          `json:"title" validate:"required"`
	OperationCodes  []OperationCode `json:"operationCodes"`
}

// WorkOrderUpdateRequest represents a request to update an existing work order
type WorkOrderUpdateRequest struct {
	WoId            string          `json:"woId" validate:"required"`
	ClerkId         string          `json:"clerkId"`
	CustId          string          `json:"custId"`
	BoatId          string          `json:"boatId"`
	BoatName        string          `json:"boatName"`
	CustomerPhone   string          `json:"customerPhone"`
	CustomerEmail   string          `json:"customerEmail"`
	Comments        string          `json:"comments"`
	LocationCode    string          `json:"locationCode"`
	EstCompDate     string          `json:"estCompDate"`
	EstStartDate    string          `json:"estStartDate"`
	CustPromiseDate string          `json:"custPromiseDate"`
	CategoryCode    string          `json:"categoryCode"`
	Title           string          `json:"title"`
	OperationCodes  []OperationCode `json:"operationCodes"`
}

// WorkOrdersForCustomerRequest represents a request to list work orders for a specific customer
type WorkOrdersForCustomerRequest struct {
	CustId           string `query:"custId" validate:"required"`
	Status           string `query:"status"`
	LocationCodeList string `query:"locationCodeList"`
}

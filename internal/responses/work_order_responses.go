package responses

import (
	"github.com/dockworks/dm-web-backend/pkg/dme"
)

// WorkOrderListShortResponse represents a paginated list of work orders
type WorkOrderListShortResponse struct {
	Data        []dme.WorkOrderShort `json:"data"`
	Total       int64                `json:"total" example:"100"`
	PerPage     int32                `json:"perPage" example:"10"`
	CurrentPage int32                `json:"currentPage" example:"1"`
	LastPage    int32                `json:"lastPage" example:"10"`
}

// WorkOrderListResponse represents a paginated list of work orders
type WorkOrderListResponse struct {
	Data        []dme.WorkOrder `json:"data"`
	Total       int64           `json:"total" example:"100"`
	PerPage     int32           `json:"perPage" example:"10"`
	CurrentPage int32           `json:"currentPage" example:"1"`
	LastPage    int32           `json:"lastPage" example:"10"`
}

// WorkOrderSearchResponse represents the response for a work order search
type WorkOrderSearchResponse struct {
	Data []dme.WorkOrderSearch `json:"data"`
}

// WorkOrderResponse represents a single work order response
type WorkOrderResponse struct {
	Data dme.WorkOrder `json:"data"`
}

// WorkOrderShortListResponse represents a list of work order short information
type WorkOrderShortListResponse struct {
	Data []dme.WorkOrderShort `json:"data"`
}

// WorkOrderOperationsResponse represents a list of available work order operations
type WorkOrderOperationsResponse struct {
	Data []dme.WorkOrderOperation `json:"data"`
}

// WorkOrderAllOperationsResponse represents a paginated list of all operation codes
type WorkOrderAllOperationsResponse struct {
	CurrentPage int                      `json:"currentPage"`
	MaxPages    int                      `json:"maxPages"`
	PageSize    int                      `json:"pageSize"`
	ListName    string                   `json:"listName"`
	Content     []dme.WorkOrderOperation `json:"content"`
}

// WorkOrderCompletedResponse represents a list of completed work orders
type WorkOrderCompletedResponse struct {
	Data []dme.WorkOrder `json:"data"`
}

// WorkOrderCreateResponse represents the response when creating a work order
type WorkOrderCreateResponse struct {
	WoId       string   `json:"woId"`
	Operations []string `json:"operations"`
	Result     string   `json:"result"`
}

// WorkOrderDeleteOperationResponse represents the response from deleting a work order operation
type WorkOrderDeleteOperationResponse struct {
	Result  string `json:"result"`
	Message string `json:"message"`
}

// swag:response WorkOrderRetrieveResponse
type WorkOrderRetrieveResponse = WorkOrderResponse

// swag:response WorkOrderCreateFromEstimateResponse
type WorkOrderCreateFromEstimateResponse = WorkOrderCreateResponse

// swag:response WorkOrderDeleteOperationResponse
type WorkOrderDeleteOperationResponseSwagger = WorkOrderDeleteOperationResponse

// ConvertWorkOrderListShort converts DME WorkOrderListShort to WorkOrderListShortResponse
func ConvertWorkOrderListShort(dmeResponse *dme.WorkOrderListShort) *WorkOrderListShortResponse {
	return &WorkOrderListShortResponse{
		Data:        dmeResponse.Content,
		Total:       int64(dmeResponse.MaxPages * dmeResponse.PageSize),
		PerPage:     int32(dmeResponse.PageSize),
		CurrentPage: int32(dmeResponse.CurrentPage),
		LastPage:    int32(dmeResponse.MaxPages),
	}
}

// ConvertWorkOrderList converts DME WorkOrderList to WorkOrderListResponse
func ConvertWorkOrderList(dmeResponse *dme.WorkOrderList) *WorkOrderListResponse {
	return &WorkOrderListResponse{
		Data:        dmeResponse.Content,
		Total:       int64(dmeResponse.MaxPages * dmeResponse.PageSize),
		PerPage:     int32(dmeResponse.PageSize),
		CurrentPage: int32(dmeResponse.CurrentPage),
		LastPage:    int32(dmeResponse.MaxPages),
	}
}

// ConvertWorkOrderSearch converts DME WorkOrderSearch to WorkOrderSearchResponse
func ConvertWorkOrderSearch(dmeResponse *[]dme.WorkOrderSearch) *WorkOrderSearchResponse {
	return &WorkOrderSearchResponse{
		Data: *dmeResponse,
	}
}

// ConvertWorkOrder converts DME WorkOrder to WorkOrderResponse
func ConvertWorkOrder(dmeResponse *dme.WorkOrder) *WorkOrderResponse {
	return &WorkOrderResponse{
		Data: *dmeResponse,
	}
}

// ConvertWorkOrderShortList converts a slice of WorkOrderShort to WorkOrderShortListResponse
func ConvertWorkOrderShortList(dmeResponse []dme.WorkOrderShort) *WorkOrderShortListResponse {
	return &WorkOrderShortListResponse{
		Data: dmeResponse,
	}
}

// ConvertWorkOrderOperations converts a slice of WorkOrderOperation to WorkOrderOperationsResponse
func ConvertWorkOrderOperations(dmeResponse []dme.WorkOrderOperation) *WorkOrderOperationsResponse {
	return &WorkOrderOperationsResponse{
		Data: dmeResponse,
	}
}

// ConvertCompletedWorkOrders converts a slice of WorkOrder to WorkOrderCompletedResponse
func ConvertCompletedWorkOrders(dmeResponse []dme.WorkOrder) *WorkOrderCompletedResponse {
	return &WorkOrderCompletedResponse{
		Data: dmeResponse,
	}
}

// WorkOrderSubletsResponse represents a list of work order sublets
type WorkOrderSubletsResponse struct {
	Data []interface{} `json:"data"`
}

// WorkOrderGroupDescriptionsResponse represents a list of work order group descriptions
type WorkOrderGroupDescriptionsResponse struct {
	Data []interface{} `json:"data"`
}

// WorkOrderPartEntryResponse represents the response from submitting a part entry
type WorkOrderPartEntryResponse struct {
	Data interface{} `json:"data"`
	Result string `json:"result"`
	Message string `json:"message"`
}

// WorkOrderTimeEntryResponse represents the response from submitting a time entry
type WorkOrderTimeEntryResponse struct {
	Data interface{} `json:"data"`
	Result string `json:"result"`
	Message string `json:"message"`
}

// WorkOrderTimeEntriesResponse represents a list of time entries
type WorkOrderTimeEntriesResponse struct {
	Data interface{} `json:"data"`
}

// WorkOrderListNewOrChangedResponse represents a paginated list of new or changed work orders
type WorkOrderListNewOrChangedResponse struct {
	Data        []dme.WorkOrder `json:"data"`
	Total       int64           `json:"total" example:"100"`
	PerPage     int32           `json:"perPage" example:"10"`
	CurrentPage int32           `json:"currentPage" example:"1"`
	LastPage    int32           `json:"lastPage" example:"10"`
}

// ConvertWorkOrderSublets converts a slice of interface{} to WorkOrderSubletsResponse
func ConvertWorkOrderSublets(dmeResponse []interface{}) *WorkOrderSubletsResponse {
	return &WorkOrderSubletsResponse{
		Data: dmeResponse,
	}
}

// ConvertWorkOrderGroupDescriptions converts a slice of interface{} to WorkOrderGroupDescriptionsResponse
func ConvertWorkOrderGroupDescriptions(dmeResponse []interface{}) *WorkOrderGroupDescriptionsResponse {
	return &WorkOrderGroupDescriptionsResponse{
		Data: dmeResponse,
	}
}

// ConvertWorkOrderPartEntry converts interface{} to WorkOrderPartEntryResponse
func ConvertWorkOrderPartEntry(dmeResponse *interface{}) *WorkOrderPartEntryResponse {
	return &WorkOrderPartEntryResponse{
		Data: *dmeResponse,
		Result: "success",
		Message: "Part entry submitted successfully",
	}
}

// ConvertWorkOrderTimeEntry converts interface{} to WorkOrderTimeEntryResponse
func ConvertWorkOrderTimeEntry(dmeResponse *interface{}) *WorkOrderTimeEntryResponse {
	return &WorkOrderTimeEntryResponse{
		Data: *dmeResponse,
		Result: "success",
		Message: "Time entry submitted successfully",
	}
}

// ConvertWorkOrderTimeEntries converts interface{} to WorkOrderTimeEntriesResponse
func ConvertWorkOrderTimeEntries(dmeResponse *interface{}) *WorkOrderTimeEntriesResponse {
	return &WorkOrderTimeEntriesResponse{
		Data: *dmeResponse,
	}
}

// ConvertWorkOrderListNewOrChanged converts DME WorkOrderList to WorkOrderListNewOrChangedResponse
func ConvertWorkOrderListNewOrChanged(dmeResponse *dme.WorkOrderList) *WorkOrderListNewOrChangedResponse {
	return &WorkOrderListNewOrChangedResponse{
		Data:        dmeResponse.Content,
		Total:       int64(dmeResponse.MaxPages * dmeResponse.PageSize),
		PerPage:     int32(dmeResponse.PageSize),
		CurrentPage: int32(dmeResponse.CurrentPage),
		LastPage:    int32(dmeResponse.MaxPages),
	}
}

// ConvertWorkOrderAllOperations converts DME OperationsListResponse to WorkOrderAllOperationsResponse
func ConvertWorkOrderAllOperations(dmeResponse *dme.OperationsListResponse) *WorkOrderAllOperationsResponse {
	return &WorkOrderAllOperationsResponse{
		CurrentPage: dmeResponse.CurrentPage,
		MaxPages:    dmeResponse.MaxPages,
		PageSize:    dmeResponse.PageSize,
		ListName:    dmeResponse.ListName,
		Content:     dmeResponse.Content,
	}
}

// SearchOperationResult represents a single operation search result
type SearchOperationResult struct {
	Opcode       string `json:"opcode"`
	Desc         string `json:"desc"`
	CategoryCode string `json:"categoryCode"`
}

// SearchAllOperationsResponse represents the response for searching operations
type SearchAllOperationsResponse struct {
	Data []SearchOperationResult `json:"data"`
}

// ConvertSearchOperations converts search results to SearchAllOperationsResponse
func ConvertSearchOperations(dmeResponse []map[string]interface{}) *SearchAllOperationsResponse {
	results := make([]SearchOperationResult, 0, len(dmeResponse))
	for _, item := range dmeResponse {
		result := SearchOperationResult{}
		if opcode, ok := item["opcode"].(string); ok {
			result.Opcode = opcode
		}
		if desc, ok := item["desc"].(string); ok {
			result.Desc = desc
		}
		if categoryCode, ok := item["categoryCode"].(string); ok {
			result.CategoryCode = categoryCode
		}
		results = append(results, result)
	}
	return &SearchAllOperationsResponse{
		Data: results,
	}
}

// SubmitSubletEntryResponse represents the response from submitting a sublet entry
type SubmitSubletEntryResponse struct {
	Success             bool     `json:"success"`
	Errors              []string `json:"errors,omitempty"`
	SubletEntryId       string   `json:"subletEntryId,omitempty"`
	SPID                string   `json:"spid,omitempty"`
	WorkOrderId         string   `json:"workOrderId,omitempty"`
	OpCode              string   `json:"opCode,omitempty"`
	VendorId            string   `json:"vendorId,omitempty"`
	PurchaseDate        string   `json:"purchaseDate,omitempty"`
	PartsPrice          float64  `json:"partsPrice"`
	PartsCost           float64  `json:"partsCost"`
	LaborPrice          float64  `json:"laborPrice"`
	LaborCost           float64  `json:"laborCost"`
	Description         string   `json:"description,omitempty"`
	TotalPrice          float64  `json:"totalPrice"`
	TotalCost           float64  `json:"totalCost"`
	BillToName          string   `json:"billToName,omitempty"`
	BillToAddress1      string   `json:"billToAddress1,omitempty"`
	BillToAddress2      string   `json:"billToAddress2,omitempty"`
	BillToAddress3      string   `json:"billToAddress3,omitempty"`
	BillToAddress4      string   `json:"billToAddress4,omitempty"`
	BillToAddress5      string   `json:"billToAddress5,omitempty"`
	BillToAddress6      string   `json:"billToAddress6,omitempty"`
	ShipToName          string   `json:"shipToName,omitempty"`
	ShipToAddress1      string   `json:"shipToAddress1,omitempty"`
	ShipToAddress2      string   `json:"shipToAddress2,omitempty"`
	ShipToAddress3      string   `json:"shipToAddress3,omitempty"`
	ShipToAddress4      string   `json:"shipToAddress4,omitempty"`
	ShipToAddress5      string   `json:"shipToAddress5,omitempty"`
	ShipToAddress6      string   `json:"shipToAddress6,omitempty"`
}

// ConvertSubmitSubletResult converts DME response to SubmitSubletEntryResponse
func ConvertSubmitSubletResult(dmeResponse *interface{}) *SubmitSubletEntryResponse {
	// Convert the interface to a map for easier access
	responseMap, ok := (*dmeResponse).(map[string]interface{})
	if !ok {
		return &SubmitSubletEntryResponse{
			Success: false,
			Errors:  []string{"Invalid response format"},
		}
	}

	response := &SubmitSubletEntryResponse{}
	
	// Extract fields from the response map
	if success, ok := responseMap["success"].(bool); ok {
		response.Success = success
	}
	
	if errors, ok := responseMap["errors"].([]interface{}); ok {
		for _, err := range errors {
			if errStr, ok := err.(string); ok {
				response.Errors = append(response.Errors, errStr)
			}
		}
	}
	
	// Extract all other fields
	if subletEntryId, ok := responseMap["subletEntryId"].(string); ok {
		response.SubletEntryId = subletEntryId
	}
	if spid, ok := responseMap["spid"].(string); ok {
		response.SPID = spid
	}
	if workOrderId, ok := responseMap["workOrderId"].(string); ok {
		response.WorkOrderId = workOrderId
	}
	if opCode, ok := responseMap["opCode"].(string); ok {
		response.OpCode = opCode
	}
	if vendorId, ok := responseMap["vendorId"].(string); ok {
		response.VendorId = vendorId
	}
	if purchaseDate, ok := responseMap["purchaseDate"].(string); ok {
		response.PurchaseDate = purchaseDate
	}
	if partsPrice, ok := responseMap["partsPrice"].(float64); ok {
		response.PartsPrice = partsPrice
	}
	if partsCost, ok := responseMap["partsCost"].(float64); ok {
		response.PartsCost = partsCost
	}
	if laborPrice, ok := responseMap["laborPrice"].(float64); ok {
		response.LaborPrice = laborPrice
	}
	if laborCost, ok := responseMap["laborCost"].(float64); ok {
		response.LaborCost = laborCost
	}
	if description, ok := responseMap["description"].(string); ok {
		response.Description = description
	}
	if totalPrice, ok := responseMap["totalPrice"].(float64); ok {
		response.TotalPrice = totalPrice
	}
	if totalCost, ok := responseMap["totalCost"].(float64); ok {
		response.TotalCost = totalCost
	}
	
	// Billing address fields
	if billToName, ok := responseMap["billToName"].(string); ok {
		response.BillToName = billToName
	}
	if billToAddress1, ok := responseMap["billToAddress1"].(string); ok {
		response.BillToAddress1 = billToAddress1
	}
	if billToAddress2, ok := responseMap["billToAddress2"].(string); ok {
		response.BillToAddress2 = billToAddress2
	}
	if billToAddress3, ok := responseMap["billToAddress3"].(string); ok {
		response.BillToAddress3 = billToAddress3
	}
	if billToAddress4, ok := responseMap["billToAddress4"].(string); ok {
		response.BillToAddress4 = billToAddress4
	}
	if billToAddress5, ok := responseMap["billToAddress5"].(string); ok {
		response.BillToAddress5 = billToAddress5
	}
	if billToAddress6, ok := responseMap["billToAddress6"].(string); ok {
		response.BillToAddress6 = billToAddress6
	}
	
	// Shipping address fields
	if shipToName, ok := responseMap["shipToName"].(string); ok {
		response.ShipToName = shipToName
	}
	if shipToAddress1, ok := responseMap["shipToAddress1"].(string); ok {
		response.ShipToAddress1 = shipToAddress1
	}
	if shipToAddress2, ok := responseMap["shipToAddress2"].(string); ok {
		response.ShipToAddress2 = shipToAddress2
	}
	if shipToAddress3, ok := responseMap["shipToAddress3"].(string); ok {
		response.ShipToAddress3 = shipToAddress3
	}
	if shipToAddress4, ok := responseMap["shipToAddress4"].(string); ok {
		response.ShipToAddress4 = shipToAddress4
	}
	if shipToAddress5, ok := responseMap["shipToAddress5"].(string); ok {
		response.ShipToAddress5 = shipToAddress5
	}
	if shipToAddress6, ok := responseMap["shipToAddress6"].(string); ok {
		response.ShipToAddress6 = shipToAddress6
	}
	
	return response
}

// WorkOrderLaborDetailResponse represents a list of labor detail entries for a work order
type WorkOrderLaborDetailResponse struct {
	Data []dme.LaborEntry `json:"data"`
}

// ConvertWorkOrderLaborDetail converts DME response to WorkOrderLaborDetailResponse
func ConvertWorkOrderLaborDetail(dmeResponse []dme.LaborEntry) *WorkOrderLaborDetailResponse {
	return &WorkOrderLaborDetailResponse{
		Data: dmeResponse,
	}
}

// WorkOrderPartsResponse represents a list of parts for a work order
type WorkOrderPartsResponse struct {
	Data []dme.WorkOrderDetailPartEntry `json:"data"`
}

// ConvertWorkOrderParts converts DME response to WorkOrderPartsResponse
func ConvertWorkOrderParts(dmeResponse []dme.WorkOrderDetailPartEntry) *WorkOrderPartsResponse {
	return &WorkOrderPartsResponse{
		Data: dmeResponse,
	}
}

// WorkOrderPartDetailResponse represents detailed part information for a work order operation
type WorkOrderPartDetailResponse struct {
	Data []dme.WorkOrderPartDetail `json:"data"`
}

// ConvertWorkOrderPartDetail converts DME response to WorkOrderPartDetailResponse
func ConvertWorkOrderPartDetail(dmeResponse []dme.WorkOrderPartDetail) *WorkOrderPartDetailResponse {
	return &WorkOrderPartDetailResponse{
		Data: dmeResponse,
	}
}

// WorkOrderLaborDetailRecordsResponse represents comprehensive labor detail records for a work order operation
type WorkOrderLaborDetailRecordsResponse struct {
	Data []dme.WorkOrderLaborDetailRecord `json:"data"`
}

// ConvertWorkOrderLaborDetailRecords converts DME response to WorkOrderLaborDetailRecordsResponse
func ConvertWorkOrderLaborDetailRecords(dmeResponse []dme.WorkOrderLaborDetailRecord) *WorkOrderLaborDetailRecordsResponse {
	return &WorkOrderLaborDetailRecordsResponse{
		Data: dmeResponse,
	}
}
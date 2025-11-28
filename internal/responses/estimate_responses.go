package responses

import (
	"github.com/dockworks/dm-web-backend/pkg/dme"
)

// EstimateListShortResponse represents a paginated list of estimates
type EstimateListShortResponse struct {
	Data        []dme.WorkOrderShort `json:"data"`
	Total       int64                `json:"total" example:"100"`
	PerPage     int32                `json:"perPage" example:"10"`
	CurrentPage int32                `json:"currentPage" example:"1"`
	LastPage    int32                `json:"lastPage" example:"10"`
}

// EstimateListResponse represents a paginated list of estimates
type EstimateListResponse struct {
	Data        []dme.WorkOrder `json:"data"`
	Total       int64           `json:"total" example:"100"`
	PerPage     int32           `json:"perPage" example:"10"`
	CurrentPage int32           `json:"currentPage" example:"1"`
	LastPage    int32           `json:"lastPage" example:"10"`
}

// EstimateSearchResponse represents the response for an estimate search
type EstimateSearchResponse struct {
	Data []dme.WorkOrderSearch `json:"data"`
}

// EstimateResponse represents a single estimate response
type EstimateResponse struct {
	Data dme.WorkOrder `json:"data"`
}

// EstimateShortListResponse represents a list of estimate short information
type EstimateShortListResponse struct {
	Data []dme.WorkOrderShort `json:"data"`
}

// EstimateSubletsResponse represents a list of estimate sublets
type EstimateSubletsResponse struct {
	Data []interface{} `json:"data"`
}

// EstimateCreateResponse represents the response when creating an estimate
type EstimateCreateResponse struct {
	EstId      string   `json:"estId"`
	Operations []string `json:"operations"`
	Result     string   `json:"result"`
}

// EstimateDeleteOperationResponse represents the response from deleting an estimate operation
type EstimateDeleteOperationResponse struct {
	Result  string `json:"result"`
	Message string `json:"message"`
}

// EstimateUpdateResponse represents the response when updating an estimate
type EstimateUpdateResponse struct {
	Data dme.WorkOrder `json:"data"`
}

// swag:response EstimateRetrieveResponse
type EstimateRetrieveResponse = EstimateResponse

// swag:response EstimateDeleteOperationResponse
type EstimateDeleteOperationResponseSwagger = EstimateDeleteOperationResponse

// ConvertEstimateListShort converts DME WorkOrderListShort to EstimateListShortResponse
func ConvertEstimateListShort(dmeResponse *dme.WorkOrderListShort) *EstimateListShortResponse {
	return &EstimateListShortResponse{
		Data:        dmeResponse.Content,
		Total:       int64(dmeResponse.MaxPages * dmeResponse.PageSize),
		PerPage:     int32(dmeResponse.PageSize),
		CurrentPage: int32(dmeResponse.CurrentPage),
		LastPage:    int32(dmeResponse.MaxPages),
	}
}

// ConvertEstimateList converts DME WorkOrderList to EstimateListResponse
func ConvertEstimateList(dmeResponse *dme.WorkOrderList) *EstimateListResponse {
	return &EstimateListResponse{
		Data:        dmeResponse.Content,
		Total:       int64(dmeResponse.MaxPages * dmeResponse.PageSize),
		PerPage:     int32(dmeResponse.PageSize),
		CurrentPage: int32(dmeResponse.CurrentPage),
		LastPage:    int32(dmeResponse.MaxPages),
	}
}

// ConvertEstimateSearch converts DME WorkOrderSearch to EstimateSearchResponse
func ConvertEstimateSearch(dmeResponse *[]dme.WorkOrderSearch) *EstimateSearchResponse {
	return &EstimateSearchResponse{
		Data: *dmeResponse,
	}
}

// ConvertEstimate converts DME WorkOrder to EstimateResponse
func ConvertEstimate(dmeResponse *dme.WorkOrder) *EstimateResponse {
	return &EstimateResponse{
		Data: *dmeResponse,
	}
}

// ConvertEstimateShortList converts a slice of WorkOrderShort to EstimateShortListResponse
func ConvertEstimateShortList(dmeResponse []dme.WorkOrderShort) *EstimateShortListResponse {
	return &EstimateShortListResponse{
		Data: dmeResponse,
	}
}

// ConvertEstimateSublets converts a slice of interface{} to EstimateSubletsResponse
func ConvertEstimateSublets(dmeResponse []interface{}) *EstimateSubletsResponse {
	return &EstimateSubletsResponse{
		Data: dmeResponse,
	}
}

// ConvertEstimateUpdate converts DME WorkOrder to EstimateUpdateResponse
func ConvertEstimateUpdate(dmeResponse *dme.WorkOrder) *EstimateUpdateResponse {
	return &EstimateUpdateResponse{
		Data: *dmeResponse,
	}
}

// SubmitEstimateSubletEntryResponse represents the response from submitting a sublet entry for an estimate
type SubmitEstimateSubletEntryResponse struct {
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

// ConvertEstimateSubmitSubletResult converts DME response to SubmitEstimateSubletEntryResponse
func ConvertEstimateSubmitSubletResult(dmeResponse *interface{}) *SubmitEstimateSubletEntryResponse {
	// Convert the interface to a map for easier access
	responseMap, ok := (*dmeResponse).(map[string]interface{})
	if !ok {
		return &SubmitEstimateSubletEntryResponse{
			Success: false,
			Errors:  []string{"Invalid response format"},
		}
	}

	response := &SubmitEstimateSubletEntryResponse{}
	
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
	
	// Extract all other fields (same as work order version)
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
	
	// Billing and shipping address fields (same as work order version)
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

// SubmitEstimatePartEntryResponse represents the response from submitting a part entry for an estimate
type SubmitEstimatePartEntryResponse struct {
	Data    interface{} `json:"data"`
	Result  string      `json:"result"`
	Message string      `json:"message"`
}

// ConvertEstimateSubmitPartResult converts DME response to SubmitEstimatePartEntryResponse
func ConvertEstimateSubmitPartResult(dmeResponse *interface{}) *SubmitEstimatePartEntryResponse {
	return &SubmitEstimatePartEntryResponse{
		Data:    *dmeResponse,
		Result:  "success",
		Message: "Part entry submitted successfully",
	}
}

// EstimatePartsResponse represents a list of parts for an estimate
type EstimatePartsResponse struct {
	Data []dme.WorkOrderDetailPartEntry `json:"data"`
}

// ConvertEstimateParts converts DME response to EstimatePartsResponse
func ConvertEstimateParts(dmeResponse []dme.WorkOrderDetailPartEntry) *EstimatePartsResponse {
	return &EstimatePartsResponse{
		Data: dmeResponse,
	}
}

// EstimateLaborResponse represents a list of labor entries for an estimate
type EstimateLaborResponse struct {
	Data []dme.LaborEntry `json:"data"`
}

// ConvertEstimateLabor converts DME response to EstimateLaborResponse
func ConvertEstimateLabor(dmeResponse []dme.LaborEntry) *EstimateLaborResponse {
	return &EstimateLaborResponse{
		Data: dmeResponse,
	}
}

// SubmitEstimateLaborEntryResponse represents the response from submitting a labor entry for an estimate
type SubmitEstimateLaborEntryResponse struct {
	Data    interface{} `json:"data"`
	Result  string      `json:"result"`
	Message string      `json:"message"`
}

// ConvertEstimateSubmitLaborResult converts DME response to SubmitEstimateLaborEntryResponse
func ConvertEstimateSubmitLaborResult(dmeResponse *interface{}) *SubmitEstimateLaborEntryResponse {
	return &SubmitEstimateLaborEntryResponse{
		Data:    *dmeResponse,
		Result:  "success",
		Message: "Labor entry submitted successfully",
	}
}
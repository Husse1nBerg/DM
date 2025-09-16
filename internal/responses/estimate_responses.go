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

// ConvertEstimateList converts DME WorkOrder array to EstimateListResponse
func ConvertEstimateList(dmeResponse []dme.WorkOrder) *EstimateListResponse {
	return &EstimateListResponse{
		Data: dmeResponse,
		// DME Estimates API returns a plain array without pagination metadata
		// Setting basic values for API consistency
		Total:       int64(len(dmeResponse)),
		PerPage:     int32(len(dmeResponse)),
		CurrentPage: 1,
		LastPage:    1,
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

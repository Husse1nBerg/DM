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

// swag:response WorkOrderRetrieveResponse
type WorkOrderRetrieveResponse = WorkOrderResponse

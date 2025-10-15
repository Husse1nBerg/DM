package responses

import "github.com/dockworks/dm-web-backend/pkg/dme"

// ServiceOpCodesResponse represents a list of operation codes
type ServiceOpCodesResponse struct {
	CurrentPage int                        `json:"currentPage"`
	MaxPages    int                        `json:"maxPages"`
	PageSize    int                        `json:"pageSize"`
	ListName    string                     `json:"listName"`
	OpCodes     []dme.WorkOrderOperation   `json:"opCodes"`
}

// ServiceWOCategoryCodesResponse represents a list of work order category codes
type ServiceWOCategoryCodesResponse struct {
	CurrentPage int                        `json:"currentPage"`
	MaxPages    int                        `json:"maxPages"`
	PageSize    int                        `json:"pageSize"`
	ListName    string                     `json:"listName"`
	OpCodes     []dme.WorkOrderOperation   `json:"opCodes"`
}

// ServiceOPCategoryCodesResponse represents a list of operation category codes
type ServiceOPCategoryCodesResponse struct {
	CurrentPage int                        `json:"currentPage"`
	MaxPages    int                        `json:"maxPages"`
	PageSize    int                        `json:"pageSize"`
	ListName    string                     `json:"listName"`
	OpCodes     []dme.WorkOrderOperation   `json:"opCodes"`
}

// ServiceOperationDescriptionsResponse represents operation descriptions
type ServiceOperationDescriptionsResponse struct {
	Data interface{} `json:"data"`
}

// ServiceTechniciansResponse represents a list of technicians
type ServiceTechniciansResponse struct {
	Data interface{} `json:"data"`
}

// ConvertServiceOpCodes converts OpCodeListResponse to ServiceOpCodesResponse
func ConvertServiceOpCodes(dmeResponse *dme.OpCodeListResponse) *ServiceOpCodesResponse {
	return &ServiceOpCodesResponse{
		CurrentPage: dmeResponse.CurrentPage,
		MaxPages:    dmeResponse.MaxPages,
		PageSize:    dmeResponse.PageSize,
		ListName:    dmeResponse.ListName,
		OpCodes:     dmeResponse.OpCodes,
	}
}

// ConvertServiceWOCategoryCodes converts OpCodeListResponse to ServiceWOCategoryCodesResponse
func ConvertServiceWOCategoryCodes(dmeResponse *dme.OpCodeListResponse) *ServiceWOCategoryCodesResponse {
	return &ServiceWOCategoryCodesResponse{
		CurrentPage: dmeResponse.CurrentPage,
		MaxPages:    dmeResponse.MaxPages,
		PageSize:    dmeResponse.PageSize,
		ListName:    dmeResponse.ListName,
		OpCodes:     dmeResponse.OpCodes,
	}
}

// ConvertServiceOPCategoryCodes converts OpCodeListResponse to ServiceOPCategoryCodesResponse
func ConvertServiceOPCategoryCodes(dmeResponse *dme.OpCodeListResponse) *ServiceOPCategoryCodesResponse {
	return &ServiceOPCategoryCodesResponse{
		CurrentPage: dmeResponse.CurrentPage,
		MaxPages:    dmeResponse.MaxPages,
		PageSize:    dmeResponse.PageSize,
		ListName:    dmeResponse.ListName,
		OpCodes:     dmeResponse.OpCodes,
	}
}

// ConvertServiceOperationDescriptions converts interface{} to ServiceOperationDescriptionsResponse
func ConvertServiceOperationDescriptions(dmeResponse *interface{}) *ServiceOperationDescriptionsResponse {
	return &ServiceOperationDescriptionsResponse{
		Data: *dmeResponse,
	}
}

// ConvertServiceTechnicians converts interface{} to ServiceTechniciansResponse
func ConvertServiceTechnicians(dmeResponse *interface{}) *ServiceTechniciansResponse {
	return &ServiceTechniciansResponse{
		Data: *dmeResponse,
	}
}

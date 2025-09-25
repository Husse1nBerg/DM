package responses

// ServiceOpCodesResponse represents a list of operation codes
type ServiceOpCodesResponse struct {
	Data interface{} `json:"data"`
}

// ServiceWOCategoryCodesResponse represents a list of work order category codes
type ServiceWOCategoryCodesResponse struct {
	Data interface{} `json:"data"`
}

// ServiceOPCategoryCodesResponse represents a list of operation category codes
type ServiceOPCategoryCodesResponse struct {
	Data interface{} `json:"data"`
}

// ServiceOperationDescriptionsResponse represents operation descriptions
type ServiceOperationDescriptionsResponse struct {
	Data interface{} `json:"data"`
}

// ServiceTechniciansResponse represents a list of technicians
type ServiceTechniciansResponse struct {
	Data interface{} `json:"data"`
}

// ConvertServiceOpCodes converts interface{} to ServiceOpCodesResponse
func ConvertServiceOpCodes(dmeResponse *interface{}) *ServiceOpCodesResponse {
	return &ServiceOpCodesResponse{
		Data: *dmeResponse,
	}
}

// ConvertServiceWOCategoryCodes converts interface{} to ServiceWOCategoryCodesResponse
func ConvertServiceWOCategoryCodes(dmeResponse *interface{}) *ServiceWOCategoryCodesResponse {
	return &ServiceWOCategoryCodesResponse{
		Data: *dmeResponse,
	}
}

// ConvertServiceOPCategoryCodes converts interface{} to ServiceOPCategoryCodesResponse
func ConvertServiceOPCategoryCodes(dmeResponse *interface{}) *ServiceOPCategoryCodesResponse {
	return &ServiceOPCategoryCodesResponse{
		Data: *dmeResponse,
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

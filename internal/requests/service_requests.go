package requests

// ServiceListNewOrChangedOpCodesRequest represents a request to list new or changed operation codes
type ServiceListNewOrChangedOpCodesRequest struct {
	AsOfDate string `query:"AsOfDate" validate:"required"`
	Page     int    `query:"page" validate:"required,min=1"`
	PageSize int    `query:"pageSize" validate:"required,min=1,max=100"`
}

// ServiceListWOCategoryCodesRequest represents a request to list work order category codes
type ServiceListWOCategoryCodesRequest struct {
	Page     int `query:"page" validate:"required,min=1"`
	PageSize int `query:"pageSize" validate:"required,min=1,max=100"`
}

// ServiceListOPCategoryCodesRequest represents a request to list operation category codes
type ServiceListOPCategoryCodesRequest struct {
	Page     int `query:"page" validate:"required,min=1"`
	PageSize int `query:"pageSize" validate:"required,min=1,max=100"`
}

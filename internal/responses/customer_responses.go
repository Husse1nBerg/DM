package responses

import (
	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/pkg/dme"
	"github.com/google/uuid"
)

// CustomerListShortResponse represents a paginated list of customers
type CustomerListShortResponse struct {
	Data        []dme.CustomerShort `json:"data"`
	Total       int64               `json:"total" example:"100"`
	PerPage     int32               `json:"perPage" example:"10"`
	CurrentPage int32               `json:"currentPage" example:"1"`
	LastPage    int32               `json:"lastPage" example:"10"`
}

// CustomerListResponse represents a paginated list of customers
type CustomerListResponse struct {
	Data        []dme.Customer `json:"data"`
	Total       int64          `json:"total" example:"100"`
	PerPage     int32          `json:"perPage" example:"10"`
	CurrentPage int32          `json:"currentPage" example:"1"`
	LastPage    int32          `json:"lastPage" example:"10"`
}

// CustomerSearchResponse represents the response for a customer search
type CustomerSearchResponse struct {
	Data []dme.CustomerSearch `json:"data"`
}

// CustomerResponse represents a single customer response
type CustomerResponse struct {
	Data dme.Customer `json:"data"`
}

// CustomerSettingsResponse represents customer settings response
type CustomerSettingsResponse struct {
	MarinaID       uuid.UUID `json:"marinaId"`
	CustomerID     string    `json:"customerId"`
	EnablePortal   bool      `json:"enablePortal"`
	CustomerUserID *string   `json:"customerUserId"`
}

// ConvertCustomerSettings converts DB CustomerSetting to CustomerSettingsResponse
func ConvertCustomerSettings(settings db.CustomerSetting) *CustomerSettingsResponse {
	enablePortal := false
	if settings.EnablePortal != nil {
		enablePortal = *settings.EnablePortal
	}
	customerUserID := ""
	if settings.CustomerUserID != uuid.Nil {
		customerUserID = settings.CustomerUserID.String()
	}

	return &CustomerSettingsResponse{
		MarinaID:       settings.MarinaID,
		CustomerID:     settings.CustomerID,
		EnablePortal:   enablePortal,
		CustomerUserID: &customerUserID,
	}
}

// ConvertCustomerListShort converts DME CustomerListShort to CustomerListShortResponse
func ConvertCustomerListShort(dmeResponse *dme.CustomerListShort) *CustomerListShortResponse {
	return &CustomerListShortResponse{
		Data:        dmeResponse.Content,
		Total:       int64(dmeResponse.MaxPages * dmeResponse.PageSize),
		PerPage:     int32(dmeResponse.PageSize),
		CurrentPage: int32(dmeResponse.CurrentPage),
		LastPage:    int32(dmeResponse.MaxPages),
	}
}

// ConvertCustomerList converts DME CustomerList to CustomerListResponse
func ConvertCustomerList(dmeResponse *dme.CustomerList) *CustomerListResponse {
	return &CustomerListResponse{
		Data:        dmeResponse.Content,
		Total:       int64(dmeResponse.MaxPages * dmeResponse.PageSize),
		PerPage:     int32(dmeResponse.PageSize),
		CurrentPage: int32(dmeResponse.CurrentPage),
		LastPage:    int32(dmeResponse.MaxPages),
	}
}

// ConvertCustomerSearch converts DME CustomerSearch to CustomerSearchResponse
func ConvertCustomerSearch(dmeResponse *[]dme.CustomerSearch) *CustomerSearchResponse {
	return &CustomerSearchResponse{
		Data: *dmeResponse,
	}
}

// ConvertCustomer converts DME Customer to CustomerResponse
func ConvertCustomer(dmeResponse *dme.Customer) *CustomerResponse {
	return &CustomerResponse{
		Data: *dmeResponse,
	}
}

// swag:response CustomerRetrieveResponse
type CustomerRetrieveResponse = CustomerResponse

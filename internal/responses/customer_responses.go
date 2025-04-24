package responses

import (
	"github.com/dockworks/dm-web-backend/pkg/dme"
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
	Customer dme.Customer `json:"customer"`
}

// ConvertCustomerListShort converts DME CustomerListShort to CustomerListShortResponse
func ConvertCustomerListShort(dmeResponse *dme.CustomerListShort) *CustomerListShortResponse {
	return &CustomerListShortResponse{
		Data:        dmeResponse.Content,
		Total:       int64(len(dmeResponse.Content)),
		PerPage:     int32(dmeResponse.PageSize),
		CurrentPage: int32(dmeResponse.CurrentPage),
		LastPage:    int32(dmeResponse.MaxPages),
	}
}

// ConvertCustomerList converts DME CustomerList to CustomerListResponse
func ConvertCustomerList(dmeResponse *dme.CustomerList) *CustomerListResponse {
	return &CustomerListResponse{
		Data:        dmeResponse.Content,
		Total:       int64(len(dmeResponse.Content)),
		PerPage:     int32(dmeResponse.PageSize),
		CurrentPage: int32(dmeResponse.CurrentPage),
		LastPage:    int32(dmeResponse.MaxPages),
	}
}

// ConvertCustomerSearch converts DME CustomerSearch to CustomerSearchResponse
func ConvertCustomerSearch(dmeResponse *dme.CustomerSearch) *CustomerSearchResponse {
	return &CustomerSearchResponse{
		Data: []dme.CustomerSearch{*dmeResponse},
	}
}

// ConvertCustomer converts DME Customer to CustomerResponse
func ConvertCustomer(dmeResponse *dme.Customer) *CustomerResponse {
	return &CustomerResponse{
		Customer: *dmeResponse,
	}
}

// swag:response CustomerRetrieveResponse
type CustomerRetrieveResponse = CustomerResponse

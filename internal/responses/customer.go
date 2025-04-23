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

// swag:response CustomerSearchResponse
type CustomerSearchResponse interface{} // This is a generic response type since the DME API returns different structures

// swag:response CustomerRetrieveResponse
type CustomerRetrieveResponse interface{} // This is a generic response type since the DME API returns different structures

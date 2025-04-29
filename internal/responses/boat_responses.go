package responses

import (
	"github.com/dockworks/dm-web-backend/pkg/dme"
)

// BoatListResponse represents a paginated list of boats
type BoatListResponse struct {
	Data        []dme.Boat `json:"data"`
	Total       int64      `json:"total" example:"100"`
	PerPage     int32      `json:"perPage" example:"10"`
	CurrentPage int32      `json:"currentPage" example:"1"`
	LastPage    int32      `json:"lastPage" example:"10"`
}

// BoatResponse represents a single boat response
type BoatResponse struct {
	Data dme.Boat `json:"data"`
}

// BoatCreateUpdateResponse represents the response after creating or updating a boat
type BoatCreateUpdateResponse struct {
	BoatID string `json:"boatID"`
}

// BoatSearchResponse represents the response for boat search results
type BoatSearchResponse struct {
	Data []dme.BoatSearch `json:"data"`
}

// ConvertBoatList converts a generic boat list to BoatListResponse
func ConvertBoatList(content []dme.Boat, currentPage, maxPages, pageSize int) *BoatListResponse {
	return &BoatListResponse{
		Data:        content,
		Total:       int64(maxPages * pageSize),
		PerPage:     int32(pageSize),
		CurrentPage: int32(currentPage),
		LastPage:    int32(maxPages),
	}
}

// ConvertBoat converts DME Boat to BoatResponse
func ConvertBoat(dmeResponse *dme.Boat) *BoatResponse {
	return &BoatResponse{
		Data: *dmeResponse,
	}
}

// ConvertBoatSearch converts a slice of DME BoatSearch to BoatSearchResponse
func ConvertBoatSearch(dmeResponse *[]dme.BoatSearch) *BoatSearchResponse {
	return &BoatSearchResponse{
		Data: *dmeResponse,
	}
}

// ConvertBoatCreateUpdate converts a boat ID to BoatCreateUpdateResponse
func ConvertBoatCreateUpdate(boatID string) *BoatCreateUpdateResponse {
	return &BoatCreateUpdateResponse{
		BoatID: boatID,
	}
}

// swag:response BoatRetrieveResponse
type BoatRetrieveResponse = BoatResponse

// swag:response BoatCreateResponse
type BoatCreateResponse = BoatCreateUpdateResponse

// swag:response BoatUpdateResponse
type BoatUpdateResponse = BoatCreateUpdateResponse

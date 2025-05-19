package requests

import "encoding/json"

// DMEAPIRequest represents a request to the DME API
type DMEAPIRequest struct {
	OrganizationID string          `json:"organizationId" validate:"required,uuid"`
	SystemID       string          `json:"systemId" validate:"required"`
	Method         string          `json:"method" validate:"required,oneof=GET POST PUT PATCH DELETE"`
	Endpoint       string          `json:"endpoint" validate:"required"`
	Body           json.RawMessage `json:"body"`
}

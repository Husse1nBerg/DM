package responses

import "github.com/dockworks/dm-web-backend/pkg/dme"

// CustomerContractsResponse represents the contracts returned for a customer
type CustomerContractsResponse struct {
	Contracts []dme.CustomerContract `json:"contracts"`
}

func NewCustomerContractsResponse(contracts []dme.CustomerContract) CustomerContractsResponse {
	return CustomerContractsResponse{Contracts: contracts}
}

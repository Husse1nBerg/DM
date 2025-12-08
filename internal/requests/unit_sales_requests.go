package requests

// CustomerContractsRequest represents a request to retrieve customer contracts
type CustomerContractsRequest struct {
	CustomerID string `query:"CustomerId" validate:"required"`
}

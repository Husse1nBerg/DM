package requests

// CustomerListRequest represents a request to list customers by page
type CustomerListRequest struct {
	Page     int `query:"page" validate:"required,min=1"`
	PageSize int `query:"pageSize" validate:"required,min=1,max=100"`
}

// CustomerRetrieveRequest represents a request to retrieve a customer by ID
type CustomerRetrieveRequest struct {
	CustomerID string `query:"CustomerId" validate:"required"`
}

// CustomerSearchRequest represents a request to search for customers
type CustomerSearchRequest struct {
	SearchString string `query:"SearchString" validate:"required"`
	DirectHit    bool   `query:"DirectHit"`
}

package requests

// BoatsListRequest represents a request to list boats by page
type BoatsListRequest struct {
	Page     int `query:"page" validate:"required,min=1"`
	PageSize int `query:"pageSize" validate:"required,min=1,max=100"`
}

// CustomerRetrieveRequest represents a request to retrieve a customer by ID
type BoatRetrieveRequest struct {
	BoatID string `query:"BoatId" validate:"required"`
}

// BoatSearchRequest represents a request to search for boats
type BoatSearchRequest struct {
	SearchString string `query:"SearchString" validate:"required"`
	DirectHit    bool   `query:"DirectHit"`
}

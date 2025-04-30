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

type BoatCreateRequest struct {
	OwnerID      string `json:"ownerId"`
	Name         string `json:"name"`
	Registration string `json:"registration"`
	Year         string `json:"year"`
	Make         string `json:"make"`
	Model        string `json:"model"`
	Hin          string `json:"hin"`
	Height       string `json:"height"`
}

package requests

import "github.com/go-playground/validator/v10"

type PaginationQuery struct {
	Page     int32 `query:"page" validate:"gte=1" default:"1"`
	PageSize int32 `query:"pageSize" validate:"gte=1,lte=100" default:"10"`
}

func (p *PaginationQuery) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

// FilterSortParams provides a standard structure for filtering and sorting across endpoints
// Filters can be used for arbitrary key-value pairs for flexible filtering
type FilterSortParams struct {
	Page      int32             `query:"page" json:"page" validate:"gte=1" default:"1"`
	PageSize  int32             `query:"pageSize" json:"pageSize" validate:"gte=1,lte=100" default:"10"`
	SortBy    string            `query:"sortBy" json:"sortBy"`
	SortOrder string            `query:"sortOrder" json:"sortOrder" validate:"omitempty,oneof=asc desc"`
	Filters   map[string]string `query:"filters" json:"filters" validate:"omitempty"`
}

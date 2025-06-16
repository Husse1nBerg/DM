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

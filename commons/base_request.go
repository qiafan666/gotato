package commons

import "context"

type BaseRequest struct {
	Ctx      context.Context `json:"ctx"`
	Language string          `json:"language"`
}

type BasePaginationNoRequire struct {
	CurrentPage int `json:"current_page"`
	PageCount   int `json:"page_count"`
}

type BasePagination struct {
	CurrentPage int `json:"current_page" validate:"required,min=1"`
	PageCount   int `json:"page_count" validate:"required,max=50"`
}

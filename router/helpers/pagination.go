package helpers

import (
	"net/http"
)

type PaginationParams struct {
	Page     uint
	PageSize uint
}

func GetPaginationParams(r *http.Request, defaultPageSize uint) (PaginationParams, error) {
	params := PaginationParams{
		Page:     1,
		PageSize: defaultPageSize,
	}

	var optionalPageSize Optional[uint64]
	optionalPageSize = GetUintQueryParam("page", r)
	if optionalPageSize.IsPresent() {
		params.Page = uint(optionalPageSize.Get())
	}

	optionalPageSize = GetUintQueryParam("pageSize", r)
	if optionalPageSize.IsPresent() {
		params.PageSize = uint(optionalPageSize.Get())
	}

	return params, nil
}

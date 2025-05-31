package helpers

import (
	"database/sql"
	"net/http"
	"strconv"
)

type PaginationParams struct {
	Page     uint
	PageSize uint
}

func GetPaginationParams(r *http.Request, defaultPageSize uint) PaginationParams {
	params := PaginationParams{
		Page:     1,
		PageSize: defaultPageSize,
	}

	page := getNullUintQueryParam("page", r)
	if page.Valid {
		params.Page = uint(page.Int64)
	}

	pageSize := getNullUintQueryParam("pageSize", r)
	if pageSize.Valid {
		params.PageSize = uint(pageSize.Int64)
	}

	return params
}

func getNullUintQueryParam(key string, r *http.Request) sql.NullInt64 {
	values := r.URL.Query()[key]
	if len(values) == 0 {
		return sql.NullInt64{Valid: false}
	}
	val, err := strconv.ParseUint(values[0], 10, 64)
	if err != nil {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{Int64: int64(val), Valid: true}
}

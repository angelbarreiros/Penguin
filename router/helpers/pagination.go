package helpers

import (
	"database/sql"
	"net/http"
	"strconv"
)

// DefaultMaxPageSize bounds pageSize in GetPaginationParams.
const DefaultMaxPageSize uint = 1000

type PaginationParams struct {
	Page     uint
	PageSize uint
}

// Offset returns (Page-1)*PageSize.
func (p PaginationParams) Offset() uint {
	return (p.Page - 1) * p.PageSize
}

// GetPaginationParams reads page (>= 1) and pageSize (1..DefaultMaxPageSize,
// or defaultPageSize when that is larger).
func GetPaginationParams(r *http.Request, defaultPageSize uint) PaginationParams {
	return GetPaginationParamsWithMax(r, defaultPageSize, max(DefaultMaxPageSize, defaultPageSize))
}

// GetPaginationParamsWithMax is GetPaginationParams with an explicit page size cap.
func GetPaginationParamsWithMax(r *http.Request, defaultPageSize uint, maxPageSize uint) PaginationParams {
	if defaultPageSize == 0 {
		defaultPageSize = 1
	}
	if maxPageSize == 0 {
		maxPageSize = defaultPageSize
	}
	params := PaginationParams{
		Page:     1,
		PageSize: min(defaultPageSize, maxPageSize),
	}

	page := getNullUintQueryParam("page", r)
	if page.Valid && page.Int64 >= 1 {
		params.Page = uint(page.Int64)
	}

	pageSize := getNullUintQueryParam("pageSize", r)
	if pageSize.Valid && pageSize.Int64 >= 1 {
		params.PageSize = min(uint(pageSize.Int64), maxPageSize)
	}

	return params
}

func getNullUintQueryParam(key string, r *http.Request) sql.NullInt64 {
	values := r.URL.Query()[key]
	if len(values) == 0 {
		return sql.NullInt64{Valid: false}
	}
	// Cap at 2^31-1 so (page-1)*pageSize cannot overflow a uint.
	val, err := strconv.ParseUint(values[0], 10, 31)
	if err != nil {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{Int64: int64(val), Valid: true}
}

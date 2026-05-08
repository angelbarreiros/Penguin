package helpers

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	routerErrors "github.com/angelbarreiros/Penguin/router/errors"
	"github.com/google/uuid"
)

func GetNullUint64ArrayQueryParam(queryParameter string, r *http.Request) ([]sql.NullInt64, error) {
	values := r.URL.Query()[queryParameter]
	result := make([]sql.NullInt64, 0, len(values))
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			result = append(result, sql.NullInt64{Valid: false})
			continue
		}
		if !numericRegex.MatchString(v) {
			return nil, routerErrors.ErrQueryParameterWrongType(queryParameter, "uint")
		}
		parsed, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return nil, routerErrors.ErrQueryParameterWrongType(queryParameter, "uint")
		}
		result = append(result, sql.NullInt64{Int64: int64(parsed), Valid: true})
	}
	return result, nil
}

func GetNullStringArrayQueryParam(queryParameter string, r *http.Request) ([]sql.NullString, error) {
	values := r.URL.Query()[queryParameter]
	result := make([]sql.NullString, 0, len(values))
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			result = append(result, sql.NullString{Valid: false})
			continue
		}
		result = append(result, sql.NullString{String: v, Valid: true})
	}
	return result, nil
}

func GetNullInt32ArrayQueryParam(queryParameter string, r *http.Request) ([]sql.NullInt32, error) {
	values := r.URL.Query()[queryParameter]
	result := make([]sql.NullInt32, 0, len(values))
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			result = append(result, sql.NullInt32{Valid: false})
			continue
		}
		if !numericRegex.MatchString(v) {
			return nil, routerErrors.ErrQueryParameterWrongType(queryParameter, "int32")
		}
		parsed, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return nil, routerErrors.ErrQueryParameterWrongType(queryParameter, "int32")
		}
		result = append(result, sql.NullInt32{Int32: int32(parsed), Valid: true})
	}
	return result, nil
}

func GetNullBoolArrayQueryParam(queryParameter string, r *http.Request) ([]sql.NullBool, error) {
	values := r.URL.Query()[queryParameter]
	result := make([]sql.NullBool, 0, len(values))
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			result = append(result, sql.NullBool{Valid: false})
			continue
		}
		if !boolRegex.MatchString(v) {
			return nil, routerErrors.ErrQueryParameterWrongType(queryParameter, "bool")
		}
		parsed, err := strconv.ParseBool(v)
		if err != nil {
			return nil, routerErrors.ErrQueryParameterWrongType(queryParameter, "bool")
		}
		result = append(result, sql.NullBool{Bool: parsed, Valid: true})
	}
	return result, nil
}

func GetNullUUIDArrayQueryParam(queryParameter string, r *http.Request) ([]uuid.NullUUID, error) {
	values := r.URL.Query()[queryParameter]
	result := make([]uuid.NullUUID, 0, len(values))
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			result = append(result, uuid.NullUUID{Valid: false})
			continue
		}
		parsed, err := uuid.Parse(v)
		if err != nil {
			return nil, routerErrors.ErrQueryParameterWrongType(queryParameter, "UUID")
		}
		result = append(result, uuid.NullUUID{UUID: parsed, Valid: true})
	}
	return result, nil
}

func GetNullTimeArrayQueryParam(queryParameter string, r *http.Request) ([]sql.NullTime, error) {
	values := r.URL.Query()[queryParameter]
	result := make([]sql.NullTime, 0, len(values))
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			result = append(result, sql.NullTime{Valid: false})
			continue
		}
		parsed, err := time.Parse(iSO8601UTC, v)
		if err == nil {
			result = append(result, sql.NullTime{Time: parsed, Valid: true})
			continue
		}
		parsed, err = time.Parse(iSO8601WithMS, v)
		if err != nil {
			return nil, routerErrors.ErrQueryParameterWrongType(queryParameter, "time")
		}
		result = append(result, sql.NullTime{Time: parsed, Valid: true})
	}
	return result, nil
}

func GetNullByteArrayQueryParam(queryParameter string, r *http.Request) ([]sql.NullByte, error) {
	values := r.URL.Query()[queryParameter]
	result := make([]sql.NullByte, 0, len(values))
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			result = append(result, sql.NullByte{Valid: false})
			continue
		}
		if !numericRegex.MatchString(v) {
			return nil, routerErrors.ErrQueryParameterWrongType(queryParameter, "byte")
		}
		parsed, err := strconv.ParseUint(v, 10, 8)
		if err != nil {
			return nil, routerErrors.ErrQueryParameterWrongType(queryParameter, "byte")
		}
		result = append(result, sql.NullByte{Byte: byte(parsed), Valid: true})
	}
	return result, nil
}

func GetNullFloat64ArrayQueryParam(queryParameter string, r *http.Request) ([]sql.NullFloat64, error) {
	values := r.URL.Query()[queryParameter]
	result := make([]sql.NullFloat64, 0, len(values))
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			result = append(result, sql.NullFloat64{Valid: false})
			continue
		}
		if !numericRegex.MatchString(v) && !floatRegex.MatchString(v) {
			return nil, routerErrors.ErrQueryParameterWrongType(queryParameter, "float64")
		}
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, routerErrors.ErrQueryParameterWrongType(queryParameter, "float64")
		}
		result = append(result, sql.NullFloat64{Float64: parsed, Valid: true})
	}
	return result, nil
}

// Non-nullable array query param functions

func GetUint64ArrayQueryParam(queryParameter string, r *http.Request) ([]uint64, error) {
	values := r.URL.Query()[queryParameter]
	result := make([]uint64, 0, len(values))
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if !numericRegex.MatchString(v) {
			return nil, routerErrors.ErrQueryParameterWrongType(queryParameter, "uint")
		}
		parsed, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return nil, routerErrors.ErrQueryParameterWrongType(queryParameter, "uint")
		}
		result = append(result, parsed)
	}
	return result, nil
}

func GetStringArrayQueryParam(queryParameter string, r *http.Request) ([]string, error) {
	values := r.URL.Query()[queryParameter]
	result := make([]string, 0, len(values))
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v != "" {
			result = append(result, v)
		}
	}
	return result, nil
}

func GetInt32ArrayQueryParam(queryParameter string, r *http.Request) ([]int32, error) {
	values := r.URL.Query()[queryParameter]
	result := make([]int32, 0, len(values))
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if !numericRegex.MatchString(v) {
			return nil, routerErrors.ErrQueryParameterWrongType(queryParameter, "int32")
		}
		parsed, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return nil, routerErrors.ErrQueryParameterWrongType(queryParameter, "int32")
		}
		result = append(result, int32(parsed))
	}
	return result, nil
}

func GetBoolArrayQueryParam(queryParameter string, r *http.Request) ([]bool, error) {
	values := r.URL.Query()[queryParameter]
	result := make([]bool, 0, len(values))
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if !boolRegex.MatchString(v) {
			return nil, routerErrors.ErrQueryParameterWrongType(queryParameter, "bool")
		}
		parsed, err := strconv.ParseBool(v)
		if err != nil {
			return nil, routerErrors.ErrQueryParameterWrongType(queryParameter, "bool")
		}
		result = append(result, parsed)
	}
	return result, nil
}

func GetUUIDArrayQueryParam(queryParameter string, r *http.Request) ([]uuid.UUID, error) {
	values := r.URL.Query()[queryParameter]
	result := make([]uuid.UUID, 0, len(values))
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		parsed, err := uuid.Parse(v)
		if err != nil {
			return nil, routerErrors.ErrQueryParameterWrongType(queryParameter, "UUID")
		}
		result = append(result, parsed)
	}
	return result, nil
}

func GetTimeArrayQueryParam(queryParameter string, r *http.Request) ([]time.Time, error) {
	values := r.URL.Query()[queryParameter]
	result := make([]time.Time, 0, len(values))
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		parsed, err := time.Parse(iSO8601UTC, v)
		if err == nil {
			result = append(result, parsed)
			continue
		}
		parsed, err = time.Parse(iSO8601WithMS, v)
		if err != nil {
			return nil, routerErrors.ErrQueryParameterWrongType(queryParameter, "time")
		}
		result = append(result, parsed)
	}
	return result, nil
}

func GetByteArrayQueryParam(queryParameter string, r *http.Request) ([]byte, error) {
	values := r.URL.Query()[queryParameter]
	result := make([]byte, 0, len(values))
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if !numericRegex.MatchString(v) {
			return nil, routerErrors.ErrQueryParameterWrongType(queryParameter, "byte")
		}
		parsed, err := strconv.ParseUint(v, 10, 8)
		if err != nil {
			return nil, routerErrors.ErrQueryParameterWrongType(queryParameter, "byte")
		}
		result = append(result, byte(parsed))
	}
	return result, nil
}

func GetFloat64ArrayQueryParam(queryParameter string, r *http.Request) ([]float64, error) {
	values := r.URL.Query()[queryParameter]
	result := make([]float64, 0, len(values))
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if !numericRegex.MatchString(v) && !floatRegex.MatchString(v) {
			return nil, routerErrors.ErrQueryParameterWrongType(queryParameter, "float64")
		}
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, routerErrors.ErrQueryParameterWrongType(queryParameter, "float64")
		}
		result = append(result, parsed)
	}
	return result, nil
}

func GetInt64ArrayQueryParam(queryParameter string, r *http.Request) ([]int64, error) {
	values := r.URL.Query()[queryParameter]
	result := make([]int64, 0, len(values))
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if !numericRegex.MatchString(v) {
			return nil, routerErrors.ErrQueryParameterWrongType(queryParameter, "int64")
		}
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, routerErrors.ErrQueryParameterWrongType(queryParameter, "int64")
		}
		result = append(result, parsed)
	}
	return result, nil
}

func GetInt16ArrayQueryParam(queryParameter string, r *http.Request) ([]int16, error) {
	values := r.URL.Query()[queryParameter]
	result := make([]int16, 0, len(values))
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if !numericRegex.MatchString(v) {
			return nil, routerErrors.ErrQueryParameterWrongType(queryParameter, "int16")
		}
		parsed, err := strconv.ParseInt(v, 10, 16)
		if err != nil {
			return nil, routerErrors.ErrQueryParameterWrongType(queryParameter, "int16")
		}
		result = append(result, int16(parsed))
	}
	return result, nil
}

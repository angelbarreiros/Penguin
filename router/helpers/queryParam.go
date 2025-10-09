package helpers

import (
	"database/sql"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	routerErrors "github.com/angelbarreiros/Penguin/router/errors"
	"github.com/google/uuid"
)

func GetNullUint64QueryParam(queryParameter string, r *http.Request) (sql.NullInt64, error) {
	id := strings.TrimSpace(r.URL.Query().Get(queryParameter))
	if id == "" {
		return sql.NullInt64{Valid: false}, nil
	}
	if !numericRegex.MatchString(id) {
		return sql.NullInt64{Valid: false}, routerErrors.ErrQueryParameterWrongType(queryParameter, "uint")
	}
	idUint, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return sql.NullInt64{Valid: false}, routerErrors.ErrQueryParameterWrongType(queryParameter, "uint")
	}
	return sql.NullInt64{Int64: int64(idUint), Valid: true}, nil
}

func GetNullInt64QueryParam(queryParameter string, r *http.Request) (sql.NullInt64, error) {
	id := strings.TrimSpace(r.URL.Query().Get(queryParameter))
	if id == "" {
		return sql.NullInt64{Valid: false}, nil
	}
	if !numericRegex.MatchString(id) {
		return sql.NullInt64{Valid: false}, routerErrors.ErrQueryParameterWrongType(queryParameter, "int")
	}
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return sql.NullInt64{Valid: false}, routerErrors.ErrQueryParameterWrongType(queryParameter, "int")
	}
	return sql.NullInt64{Int64: idInt, Valid: true}, nil
}

func GetNullBoolQueryParam(queryParameter string, r *http.Request) (sql.NullBool, error) {
	id := strings.TrimSpace(r.URL.Query().Get(queryParameter))
	if id == "" {
		return sql.NullBool{Valid: false}, nil
	}
	if !boolRegex.MatchString(id) {
		return sql.NullBool{Valid: false}, routerErrors.ErrQueryParameterWrongType(queryParameter, "bool")
	}
	idBool, err := strconv.ParseBool(id)
	if err != nil {
		return sql.NullBool{Valid: false}, routerErrors.ErrQueryParameterWrongType(queryParameter, "bool")
	}
	return sql.NullBool{Bool: idBool, Valid: true}, nil
}

func GetNullUUIDQueryParam(queryParameter string, r *http.Request) (uuid.NullUUID, error) {
	id := strings.TrimSpace(r.URL.Query().Get(queryParameter))
	if id == "" {
		return uuid.NullUUID{Valid: false}, nil
	}
	idUUID, err := uuid.Parse(id)
	if err != nil {
		return uuid.NullUUID{Valid: false}, routerErrors.ErrQueryParameterWrongType(queryParameter, "UUID")
	}
	return uuid.NullUUID{UUID: idUUID, Valid: true}, nil
}

func GetNullTimeQueryParam(queryParameter string, r *http.Request) (sql.NullTime, error) {
	id := strings.TrimSpace(r.URL.Query().Get(queryParameter))
	if id == "" {
		return sql.NullTime{Valid: false}, nil
	}

	parsedTime, err := time.Parse(iSO8601UTC, id)
	if err == nil {
		return sql.NullTime{Time: parsedTime, Valid: true}, nil
	}

	parsedTime, err = time.Parse(iSO8601WithMS, id)
	if err != nil {
		return sql.NullTime{Valid: false}, routerErrors.ErrQueryParameterWrongType(queryParameter, "time")
	}

	return sql.NullTime{Time: parsedTime, Valid: true}, nil
}

func GetNullStringQueryParam(queryParameter string, r *http.Request) (sql.NullString, error) {
	id := strings.TrimSpace(r.URL.Query().Get(queryParameter))
	if id == "" {
		return sql.NullString{Valid: false}, nil
	}
	return sql.NullString{String: id, Valid: true}, nil
}

func GetNullByteQueryParam(queryParameter string, r *http.Request) (sql.NullByte, error) {
	id := strings.TrimSpace(r.URL.Query().Get(queryParameter))
	if id == "" {
		return sql.NullByte{Valid: false}, nil
	}
	if !numericRegex.MatchString(id) {
		return sql.NullByte{Valid: false}, routerErrors.ErrQueryParameterWrongType(queryParameter, "byte")
	}
	idInt, err := strconv.ParseUint(id, 10, 8)
	if err != nil {
		return sql.NullByte{Valid: false}, routerErrors.ErrQueryParameterWrongType(queryParameter, "byte")
	}
	return sql.NullByte{Byte: byte(idInt), Valid: true}, nil
}

func GetNullFloat64QueryParam(queryParameter string, r *http.Request) (sql.NullFloat64, error) {
	id := strings.TrimSpace(r.URL.Query().Get(queryParameter))
	if id == "" {
		return sql.NullFloat64{Valid: false}, nil
	}
	if !numericRegex.MatchString(id) && !regexp.MustCompile(`^[0-9]*\.?[0-9]+$`).MatchString(id) {
		return sql.NullFloat64{Valid: false}, routerErrors.ErrQueryParameterWrongType(queryParameter, "float64")
	}
	idFloat, err := strconv.ParseFloat(id, 64)
	if err != nil {
		return sql.NullFloat64{Valid: false}, routerErrors.ErrQueryParameterWrongType(queryParameter, "float64")
	}
	return sql.NullFloat64{Float64: idFloat, Valid: true}, nil
}

func GetNullInt16QueryParam(queryParameter string, r *http.Request) (sql.NullInt16, error) {
	id := strings.TrimSpace(r.URL.Query().Get(queryParameter))
	if id == "" {
		return sql.NullInt16{Valid: false}, nil
	}
	if !numericRegex.MatchString(id) {
		return sql.NullInt16{Valid: false}, routerErrors.ErrQueryParameterWrongType(queryParameter, "int16")
	}
	idInt, err := strconv.ParseInt(id, 10, 16)
	if err != nil {
		return sql.NullInt16{Valid: false}, routerErrors.ErrQueryParameterWrongType(queryParameter, "int16")
	}
	return sql.NullInt16{Int16: int16(idInt), Valid: true}, nil
}

func GetNullInt32QueryParam(queryParameter string, r *http.Request) (sql.NullInt32, error) {
	id := strings.TrimSpace(r.URL.Query().Get(queryParameter))
	if id == "" {
		return sql.NullInt32{Valid: false}, nil
	}
	if !numericRegex.MatchString(id) {
		return sql.NullInt32{Valid: false}, routerErrors.ErrQueryParameterWrongType(queryParameter, "int32")
	}
	idInt, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return sql.NullInt32{Valid: false}, routerErrors.ErrQueryParameterWrongType(queryParameter, "int32")
	}
	return sql.NullInt32{Int32: int32(idInt), Valid: true}, nil
}

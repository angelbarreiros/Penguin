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

func GetUintPathValue(identifier string, r *http.Request) (uint64, error) {
	const maxLength = 20

	var id string = r.PathValue(identifier)
	id = strings.TrimSpace(id)
	if id == "" {
		return 0, routerErrors.ErrPathVariableMissing(identifier)
	}

	if len(id) > maxLength {
		return 0, routerErrors.ErrPathVariableTooLong(identifier, maxLength)
	}

	if !numericRegex.MatchString(id) {
		return 0, routerErrors.ErrPathVariableWrongType(identifier, "uint")
	}

	idUint, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return 0, routerErrors.ErrPathVariableWrongType(identifier, "uint")
	}
	return idUint, nil
}
func GetUUidPathValue(identifier string, r *http.Request) (uuid.UUID, error) {
	var id string = r.PathValue(identifier)
	var err error
	var idUUID uuid.UUID
	id = strings.TrimSpace(id)
	if id == "" {
		return uuid.Nil, routerErrors.ErrPathVariableMissing(identifier)
	}
	idUUID, err = uuid.Parse(id)
	if err != nil {
		return uuid.Nil, routerErrors.ErrPathVariableWrongType(identifier, "UUID")
	}
	return idUUID, nil
}

type MaxStringLengthOption func() int

func WithCustomLengthOption(length int) MaxStringLengthOption {
	if length <= 0 || length > 2147483647 {
		length = defaultMaxLength
	}
	return func() int {
		return length
	}
}

func GetStringPathValue(identifier string, r *http.Request, options ...MaxStringLengthOption) (string, error) {
	var maxLength int = defaultMaxLength
	if len(options) > 1 {
		panic("GetStringQueryValue: more than one option provided")
	}
	if len(options) > 0 {
		maxLength = options[0]()
	}

	id := strings.TrimSpace(r.PathValue(identifier))
	if id == "" {
		return "", routerErrors.ErrPathVariableMissing(identifier)
	}
	id = sanitizeRegex.ReplaceAllString(id, "")
	if id == "" {
		return "", routerErrors.ErrPathVariableMissing(identifier)
	}
	if len(id) > maxLength {
		return "", routerErrors.ErrPathVariableTooLong(identifier, maxLength)
	}

	return id, nil
}

func GetNullUint64PathValue(identifier string, r *http.Request) (sql.NullInt64, error) {
	id := strings.TrimSpace(r.PathValue(identifier))
	if id == "" {
		return sql.NullInt64{Valid: false}, nil
	}
	if !numericRegex.MatchString(id) {
		return sql.NullInt64{Valid: false}, routerErrors.ErrPathVariableWrongType(identifier, "uint64")
	}
	idUint, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return sql.NullInt64{Valid: false}, routerErrors.ErrPathVariableWrongType(identifier, "uint64")
	}
	return sql.NullInt64{Int64: int64(idUint), Valid: true}, nil
}

func GetNullInt64PathValue(identifier string, r *http.Request) (sql.NullInt64, error) {
	id := strings.TrimSpace(r.PathValue(identifier))
	if id == "" {
		return sql.NullInt64{Valid: false}, nil
	}
	if !numericRegex.MatchString(id) {
		return sql.NullInt64{Valid: false}, routerErrors.ErrPathVariableWrongType(identifier, "int64")
	}
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return sql.NullInt64{Valid: false}, routerErrors.ErrPathVariableWrongType(identifier, "int64")
	}
	return sql.NullInt64{Int64: idInt, Valid: true}, nil
}

func GetNullBoolPathValue(identifier string, r *http.Request) (sql.NullBool, error) {
	id := strings.TrimSpace(r.PathValue(identifier))
	if id == "" {
		return sql.NullBool{Valid: false}, nil
	}
	idBool, err := strconv.ParseBool(id)
	if err != nil {
		return sql.NullBool{Valid: false}, routerErrors.ErrPathVariableWrongType(identifier, "bool")
	}
	return sql.NullBool{Bool: idBool, Valid: true}, nil
}

func GetNullUUIDPathValue(identifier string, r *http.Request) (uuid.NullUUID, error) {
	id := strings.TrimSpace(r.PathValue(identifier))
	if id == "" {
		return uuid.NullUUID{Valid: false}, nil
	}
	idUUID, err := uuid.Parse(id)
	if err != nil {
		return uuid.NullUUID{Valid: false}, routerErrors.ErrPathVariableWrongType(identifier, "UUID")
	}
	return uuid.NullUUID{UUID: idUUID, Valid: true}, nil
}

func GetNullTimePathValue(identifier string, r *http.Request) (sql.NullTime, error) {
	id := strings.TrimSpace(r.PathValue(identifier))
	if id == "" {
		return sql.NullTime{Valid: false}, nil
	}
	// Try RFC3339 format first
	idTime, err := time.Parse(time.RFC3339, id)
	if err != nil {
		// Try RFC3339Nano format
		idTime, err = time.Parse(time.RFC3339Nano, id)
		if err != nil {
			return sql.NullTime{Valid: false}, routerErrors.ErrPathVariableWrongType(identifier, "time")
		}
	}
	return sql.NullTime{Time: idTime, Valid: true}, nil
}

func GetNullStringPathValue(identifier string, r *http.Request) (sql.NullString, error) {
	id := strings.TrimSpace(r.PathValue(identifier))
	if id == "" {
		return sql.NullString{Valid: false}, nil
	}
	return sql.NullString{String: id, Valid: true}, nil
}

func GetNullBytePathValue(identifier string, r *http.Request) (sql.NullByte, error) {
	id := strings.TrimSpace(r.PathValue(identifier))
	if id == "" {
		return sql.NullByte{Valid: false}, nil
	}
	if len(id) != 1 {
		return sql.NullByte{Valid: false}, routerErrors.ErrPathVariableWrongType(identifier, "byte")
	}
	return sql.NullByte{Byte: id[0], Valid: true}, nil
}

func GetNullFloat64PathValue(identifier string, r *http.Request) (sql.NullFloat64, error) {
	id := strings.TrimSpace(r.PathValue(identifier))
	if id == "" {
		return sql.NullFloat64{Valid: false}, nil
	}
	if !numericRegex.MatchString(id) {
		return sql.NullFloat64{Valid: false}, routerErrors.ErrPathVariableWrongType(identifier, "float64")
	}
	idFloat, err := strconv.ParseFloat(id, 64)
	if err != nil {
		return sql.NullFloat64{Valid: false}, routerErrors.ErrPathVariableWrongType(identifier, "float64")
	}
	return sql.NullFloat64{Float64: idFloat, Valid: true}, nil
}

func GetNullInt16PathValue(identifier string, r *http.Request) (sql.NullInt16, error) {
	id := strings.TrimSpace(r.PathValue(identifier))
	if id == "" {
		return sql.NullInt16{Valid: false}, nil
	}
	if !numericRegex.MatchString(id) {
		return sql.NullInt16{Valid: false}, routerErrors.ErrPathVariableWrongType(identifier, "int16")
	}
	idInt, err := strconv.ParseInt(id, 10, 16)
	if err != nil {
		return sql.NullInt16{Valid: false}, routerErrors.ErrPathVariableWrongType(identifier, "int16")
	}
	return sql.NullInt16{Int16: int16(idInt), Valid: true}, nil
}

func GetNullInt32PathValue(identifier string, r *http.Request) (sql.NullInt32, error) {
	id := strings.TrimSpace(r.PathValue(identifier))
	if id == "" {
		return sql.NullInt32{Valid: false}, nil
	}
	if !numericRegex.MatchString(id) {
		return sql.NullInt32{Valid: false}, routerErrors.ErrPathVariableWrongType(identifier, "int32")
	}
	idInt, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return sql.NullInt32{Valid: false}, routerErrors.ErrPathVariableWrongType(identifier, "int32")
	}
	return sql.NullInt32{Int32: int32(idInt), Valid: true}, nil
}

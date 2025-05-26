package helpers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	routerErrors "github.com/angelbarreiros/Penguin/router/errors"
	"github.com/google/uuid"
)

func GetUintQueryParam(queryParameter string, r *http.Request) Optional[uint64] {
	var id string = r.URL.Query().Get(queryParameter)
	id = strings.TrimSpace(id)
	if id == "" {
		return Optional[uint64]{present: false, hasErrors: nil}
	}
	if !numericRegex.MatchString(id) {
		return Optional[uint64]{present: false, hasErrors: routerErrors.ErrQueryParameterWrongType(queryParameter, "uint")}
	}
	idUint, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return Optional[uint64]{present: false, hasErrors: routerErrors.ErrQueryParameterWrongType(queryParameter, "uint")}
	}
	return Optional[uint64]{value: idUint, present: true, hasErrors: nil}
}

func GetIntQueryParam(queryParameter string, r *http.Request) Optional[int64] {
	var id string = r.URL.Query().Get(queryParameter)
	id = strings.TrimSpace(id)
	if id == "" {
		return Optional[int64]{present: false, hasErrors: nil}
	}
	if !numericRegex.MatchString(id) {
		return Optional[int64]{present: false, hasErrors: routerErrors.ErrQueryParameterWrongType(queryParameter, "int")}
	}
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return Optional[int64]{present: false, hasErrors: routerErrors.ErrQueryParameterWrongType(queryParameter, "int")}
	}
	return Optional[int64]{value: idInt, present: true, hasErrors: nil}
}

func GetBoolQueryParam(queryParameter string, r *http.Request) Optional[bool] {
	var id string = r.URL.Query().Get(queryParameter)
	id = strings.TrimSpace(id)
	if id == "" {
		return Optional[bool]{present: false, hasErrors: nil}
	}
	if !boolRegex.MatchString(id) {
		return Optional[bool]{present: false, hasErrors: routerErrors.ErrQueryParameterWrongType(queryParameter, "bool")}
	}
	idBool, err := strconv.ParseBool(id)
	if err != nil {
		return Optional[bool]{present: false, hasErrors: routerErrors.ErrQueryParameterWrongType(queryParameter, "bool")}
	}
	return Optional[bool]{value: idBool, present: true, hasErrors: nil}
}

func GetUUIDQueryParam(queryParameter string, r *http.Request) Optional[uuid.UUID] {
	var id string = r.URL.Query().Get(queryParameter)
	id = strings.TrimSpace(id)
	if id == "" {
		return Optional[uuid.UUID]{present: false, hasErrors: nil}
	}
	idUUID, err := uuid.Parse(id)
	if err != nil {
		return Optional[uuid.UUID]{present: false, hasErrors: routerErrors.ErrQueryParameterWrongType(queryParameter, "UUID")}
	}
	return Optional[uuid.UUID]{value: idUUID, present: true, hasErrors: nil}
}
func GetTimeQueryParam(queryParameter string, r *http.Request) Optional[time.Time] {
	var id string = strings.TrimSpace(r.URL.Query().Get(queryParameter))
	if id == "" {
		return Optional[time.Time]{present: false, hasErrors: nil}
	}

	parsedTime, err := time.Parse(iSO8601UTC, id)
	if err == nil {
		return Optional[time.Time]{value: parsedTime, present: true, hasErrors: nil}
	}

	parsedTime, err = time.Parse(iSO8601WithMS, id)
	if err != nil {
		return Optional[time.Time]{present: false, hasErrors: routerErrors.ErrQueryParameterWrongType(queryParameter, "time")}
	}

	return Optional[time.Time]{value: parsedTime, present: true, hasErrors: nil}
}

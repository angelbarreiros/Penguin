package helpers

import (
	"net/http"
	"strconv"
	"strings"

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

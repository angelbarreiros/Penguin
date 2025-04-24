package routerErrors

import (
	"errors"
	"fmt"
)

var (
	ErrPathVariableMissing = func(name string) error {
		return fmt.Errorf("path variable '%s' is missing", name)
	}
	ErrPathVariableWrongType = func(name, expectedType string) error {
		return fmt.Errorf("path variable '%s' is of the wrong type, it must be '%s'", name, expectedType)
	}
	ErrPathVariableTooLong = func(name string, maxLength int) error {
		return fmt.Errorf("path variable '%s' exceeds maximum length of %d", name, maxLength)
	}
)

var (
	ErrQueryParameterMissing = func(name string) error {
		return fmt.Errorf("query parameter '%s' is missing", name)
	}
	ErrQueryParameterWrongType = func(name, expectedType string) error {
		return fmt.Errorf("query parameter '%s' is of the wrong type, it must be '%s'", name, expectedType)
	}
	ErrQueryParameterEmpty = func(name string) error {
		return fmt.Errorf("query parameter '%s' is empty", name)
	}
	ErrQueryParameterTooLong = func(name string, maxLength int) error {
		return fmt.Errorf("query parameter '%s' exceeds maximum length of %d", name, maxLength)
	}
)

var (
	ErrRequestBodyMissing = func() error {
		return errors.New("request body is missing")
	}
	ErrRequestBodyInvalid = func(reason string) error {
		return fmt.Errorf("request body is invalid: %s", reason)
	}
	ErrRequestBodyTooLarge = func(maxSize int) error {
		return fmt.Errorf("request body exceeds maximum size of %d bytes", maxSize)
	}
)

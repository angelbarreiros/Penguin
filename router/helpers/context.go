package helpers

import (
	"fmt"
	"net/http"
)

// GetContextValue reads a typed value from the request context. key may be
// any comparable value, including the custom key types accepted by
// auth.WithCustomContextKey.
func GetContextValue[T any](r *http.Request, key any) (T, error) {
	var value T
	v := r.Context().Value(key)
	if v == nil {
		return value, fmt.Errorf("key %v not found in context", key)
	}
	val, ok := v.(T)
	if !ok {
		return value, fmt.Errorf("value for key %v is not of expected type %T", key, value)
	}
	return val, nil
}

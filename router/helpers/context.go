package helpers

import (
	"fmt"
	"net/http"
)

func GetContextValue[T any](r *http.Request, key string) (T, error) {
	var value T
	v := r.Context().Value(key)
	if v == nil {
		return value, fmt.Errorf("key %q not found in context", key)
	}
	val, ok := v.(T)
	if !ok {
		return value, fmt.Errorf("value for key %q is not of expected type %T", key, value)
	}
	return val, nil
}

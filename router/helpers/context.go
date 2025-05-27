package helpers

import "net/http"

func GetValue[T any](r *http.Request, key string) T {
	var value T
	if v, ok := r.Context().Value(key).(T); ok {
		value = v
	}
	return value
}

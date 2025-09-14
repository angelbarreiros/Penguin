package middlewares

import "net/http"

func WithQueryParametersObligation(queryParameters []string, hf http.HandlerFunc) http.HandlerFunc {
	return queryParametersObligation(queryParameters)(hf)
}
func queryParametersObligation(queryParameters []string) middlewareFunc {
	return func(hf http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			for _, queryParameter := range queryParameters {
				if r.URL.Query().Get(queryParameter) == "" {
					http.Error(w, "Query parameter "+queryParameter+" is required", http.StatusBadRequest)
					return
				}
			}
			hf(w, r)
		}
	}
}

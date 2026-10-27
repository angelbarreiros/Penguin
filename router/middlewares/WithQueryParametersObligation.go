package middlewares

import "net/http"

func WithQueryParametersObligation(queryParameters []string, hf http.HandlerFunc) http.HandlerFunc {
	return queryParametersObligation(queryParameters)(hf)
}
func queryParametersObligation(queryParameters []string) middlewareFunc {
	return func(hf http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			query := r.URL.Query()
			for _, queryParameter := range queryParameters {
				if query.Get(queryParameter) == "" {
					writeJSONError(w, http.StatusBadRequest, "Query parameter "+queryParameter+" is required")
					return
				}
			}
			hf(w, r)
		}
	}
}

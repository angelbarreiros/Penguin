package router

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
)

type Router struct {
	mux    *http.ServeMux
	routes map[string]routeEntry
}

type routeEntry struct {
	handlers       map[HTTPMethod]http.HandlerFunc
	allowedMethods string
}

func (r *Router) StartServer(port string) error {
	return http.ListenAndServe(port, r.mux)
}

func InitRouter() *Router {
	routerOnce.Do(initRouter)
	return routerInstance
}

type Route struct {
	Path              string
	Method            HTTPMethod
	Handler           http.HandlerFunc
	AdditionalMethods []HTTPMethod
	// Deprecated: use AdditionalMethods.
	AditionalMethods []HTTPMethod
}

func (r *Router) NewRoute(route Route) {
	entry, exists := r.routes[route.Path]
	if !exists {
		entry = routeEntry{handlers: make(map[HTTPMethod]http.HandlerFunc)}
		r.mux.HandleFunc(route.Path, r.methodHandler(route.Path))
	}

	if _, exists := entry.handlers[route.Method]; exists {
		panic(fmt.Sprintf("Route already exists: %s %s", route.Method, route.Path))
	}

	entry.handlers[route.Method] = route.Handler

	additionalMethods := route.AdditionalMethods
	if len(additionalMethods) == 0 {
		additionalMethods = route.AditionalMethods
	}

	for _, method := range additionalMethods {
		if _, exists := entry.handlers[method]; exists {
			continue
		}
		entry.handlers[method] = route.Handler
	}

	entry.allowedMethods = buildAllowedMethodsHeader(entry.handlers)
	r.routes[route.Path] = entry
}

func buildAllowedMethodsHeader(handlers map[HTTPMethod]http.HandlerFunc) string {
	allowedMethods := make([]string, 0, len(handlers))
	for m := range handlers {
		allowedMethods = append(allowedMethods, string(m))
	}
	sort.Strings(allowedMethods)
	return strings.Join(allowedMethods, ", ")
}

func (r *Router) methodHandler(path string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		var method HTTPMethod = HTTPMethod(req.Method)
		route := r.routes[path]
		handlers := route.handlers
		w.Header().Set("Access-Control-Allow-Methods", route.allowedMethods)

		handler, exists := handlers[method]
		if !exists {
			w.Header().Set("Allow", route.allowedMethods)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"error": "Method not allowed"}`))
			return
		}
		if req.Method == http.MethodOptions {
			handler(w, req)
			return
		}

		if req.Method == http.MethodHead {
			var rec *responseRecorder = &responseRecorder{ResponseWriter: w}
			handler(rec, req)
			return
		}

		handler(w, req)
	}
}

type responseRecorder struct {
	http.ResponseWriter
}

func (rec *responseRecorder) Write(b []byte) (int, error) {
	return len(b), nil
}

func initRouter() {
	if nil == routerInstance {
		routerInstance = &Router{
			mux:    http.NewServeMux(),
			routes: make(map[string]routeEntry),
		}
	}

}

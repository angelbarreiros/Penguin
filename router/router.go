package router

import (
	"fmt"
	"net/http"
	"strings"
)

type HandleFunc func(http.ResponseWriter, *http.Request)

type Router struct {
	mux    *http.ServeMux
	routes map[string]map[HTTPMethod]HandleFunc
}

func (r *Router) StartServer(s string) {
	panic(http.ListenAndServe(s, r.mux))
}

func InitRouter() *Router {
	routerOnce.Do(initRouter)
	return routerInstance
}

type Route struct {
	Path             string
	Method           HTTPMethod
	Handler          HandleFunc
	AditionalMethods []HTTPMethod
}

func (r *Router) NewRoute(route Route) {
	if r.routes[route.Path] == nil {
		r.routes[route.Path] = make(map[HTTPMethod]HandleFunc)
		r.mux.HandleFunc(route.Path, r.methodHandler(route.Path))
	}

	if _, exists := r.routes[route.Path][route.Method]; exists {
		panic(fmt.Sprintf("Route already exists: %s %s", method, route.Path))
	}

	r.routes[route.Path][route.Method] = route.Handler

	for _, method := range route.AditionalMethods {
		if _, exists := r.routes[route.Path][method]; exists {
			continue
		}
		r.routes[route.Path][method] = route.Handler
	}
}
func (r *Router) methodHandler(path string) HandleFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		var method HTTPMethod = HTTPMethod(req.Method)
		handlers := r.routes[path]

		var allowedMethods []string
		for m := range handlers {
			allowedMethods = append(allowedMethods, string(m))
		}
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(allowedMethods, ", "))

		handler, exists := handlers[method]
		if !exists {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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
			routes: make(map[string]map[HTTPMethod]HandleFunc),
		}
	}

}

package router

import (
	"net/http"
	"slices"
	"strings"
)

type middlewareFunc func(handleFunc) handleFunc

type handleFunc func(http.ResponseWriter, *http.Request)

type router struct {
	mux    *http.ServeMux
	routes map[string]bool
}

func (r *router) StartServer(s string) {
	panic(http.ListenAndServe(s, r.mux))
}

func Router() *router {
	routerOnce.Do(initRouter)
	return routerInstance
}

type Route struct {
	Path             string
	Method           HTTPMethod
	Handler          handleFunc
	AditionalMethods []HTTPMethod
}

func (r *router) NewRoute(route Route) {
	var sb strings.Builder
	sb.WriteString(string(route.Method))
	sb.WriteString(" ")
	sb.WriteString(route.Path)
	if _, exists := r.routes[sb.String()]; exists {
		panic("Route already exists: " + sb.String())
	}
	r.routes[sb.String()] = true
	var handler handleFunc = methodHandler(route)
	r.mux.HandleFunc(route.Path, handler)
}

func methodHandler(route Route) handleFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var allowedMethod bool = slices.Contains(append(route.AditionalMethods, route.Method), HTTPMethod(req.Method))
		if !allowedMethod {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if req.Method == http.MethodOptions {
			route.Handler(w, req)
			return
		}
		if req.Method == http.MethodHead {
			var rec *responseRecorder = &responseRecorder{ResponseWriter: w}
			route.Handler(rec, req)
			return
		}

		route.Handler(w, req)
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
		routerInstance = &router{
			mux:    http.NewServeMux(),
			routes: make(map[string]bool),
		}
	}

}

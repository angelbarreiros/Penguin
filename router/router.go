package router

import (
	"net/http"
	"strings"
)

type path string
type middlewareFunc func(handleFunc) handleFunc
type middleware middlewareFunc

type handleFunc func(http.ResponseWriter, *http.Request)
type routeOptions func(*route)
type router struct {
	mux    *http.ServeMux
	routes map[string]handleFunc
}

type route struct {
	path              string
	usedMiddlewareMap map[middlewareType]bool
}

func Router() *router {
	routerOnce.Do(initRouter)
	return routerInstance
}
func (r *router) NewRoute(path string, httpMethod HTTPMethod, handleFunc handleFunc) {
	var sb strings.Builder
	sb.WriteString(string(httpMethod))
	sb.WriteString(" ")
	sb.WriteString(path)
	if condition := r.routes[sb.String()]; condition != nil {
		panic("Route already exists")
	}
	r.mux.HandleFunc(sb.String(), handleFunc)

}
func (r *router) StartServer(s string) {
	if err := http.ListenAndServe(s, r.mux); err != nil {
		panic(err)
	}
}

func initRouter() {
	if nil == routerInstance {
		routerInstance = &router{
			mux:    http.NewServeMux(),
			routes: make(map[string]handleFunc),
		}
	}

}

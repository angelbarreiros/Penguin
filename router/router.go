package router

import (
	"net/http"
)

type path string
type middlewareFunc func(handleFunc) handleFunc
type middleware struct {
	priority       uint
	middlewareFunc middlewareFunc
}
type handleFunc func(http.ResponseWriter, *http.Request)
type routeOptions func(*route)
type router struct {
	mux    *http.ServeMux
	routes map[string]handleFunc
}

type route struct {
	path              string
	middlewares       []*middleware
	usedMiddlewareMap map[middlewareType]bool
}

func NewRouter() *router {
	routerOnce.Do(initRouter)
	return routerInstance
}
func (r *router) NewRoute(path string, handleFunc handleFunc, opts ...routeOptions) {
	if condition := r.routes[path]; condition != nil {
		panic("Route already exists")
	}

	route := route{
		path:              path,
		usedMiddlewareMap: make(map[middlewareType]bool, 10),
	}
	for _, opt := range opts {
		opt(&route)
	}
	quickSortMiddlewares(route.middlewares)
	for _, middleware := range route.middlewares {
		handleFunc = middleware.middlewareFunc(handleFunc)
	}

	r.mux.HandleFunc(path, handleFunc)

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
func quickSortMiddlewares(middlewares []*middleware) {
	if len(middlewares) < 2 {
		return
	}

	left, right := 0, len(middlewares)-1

	pivot := len(middlewares) / 2

	middlewares[pivot], middlewares[right] = middlewares[right], middlewares[pivot]

	for i := range middlewares {
		if middlewares[i].priority < middlewares[right].priority {
			middlewares[i], middlewares[left] = middlewares[left], middlewares[i]
			left++
		}
	}

	middlewares[left], middlewares[right] = middlewares[right], middlewares[left]

	quickSortMiddlewares(middlewares[:left])
	quickSortMiddlewares(middlewares[left+1:])
}

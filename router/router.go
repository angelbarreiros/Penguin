package router

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

type Router struct {
	mux    *http.ServeMux
	mu     sync.RWMutex
	routes map[string]routeEntry

	serverMu sync.Mutex
	server   *http.Server
	config   ServerConfig
}

type routeEntry struct {
	handlers       map[HTTPMethod]http.HandlerFunc
	allowedMethods string
}

// ServerConfig holds the http.Server limits used by StartServer.
type ServerConfig struct {
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	MaxHeaderBytes    int
}

// DefaultServerConfig protects against slow-client (Slowloris) exhaustion.
var DefaultServerConfig = ServerConfig{
	ReadHeaderTimeout: 10 * time.Second,
	ReadTimeout:       30 * time.Second,
	WriteTimeout:      60 * time.Second,
	IdleTimeout:       120 * time.Second,
	MaxHeaderBytes:    1 << 20,
}

// SetServerConfig overrides the server limits; call it before StartServer.
func (r *Router) SetServerConfig(config ServerConfig) {
	r.serverMu.Lock()
	defer r.serverMu.Unlock()
	r.config = config
}

func (r *Router) StartServer(port string) error {
	r.serverMu.Lock()
	r.server = &http.Server{
		Addr:              port,
		Handler:           r.mux,
		ReadHeaderTimeout: r.config.ReadHeaderTimeout,
		ReadTimeout:       r.config.ReadTimeout,
		WriteTimeout:      r.config.WriteTimeout,
		IdleTimeout:       r.config.IdleTimeout,
		MaxHeaderBytes:    r.config.MaxHeaderBytes,
	}
	server := r.server
	r.serverMu.Unlock()
	return server.ListenAndServe()
}

// Shutdown gracefully stops a server started with StartServer.
func (r *Router) Shutdown(ctx context.Context) error {
	r.serverMu.Lock()
	server := r.server
	r.serverMu.Unlock()
	if server == nil {
		return nil
	}
	return server.Shutdown(ctx)
}

// Handler exposes the router as an http.Handler (e.g. for tests or a custom server).
func (r *Router) Handler() http.Handler {
	return r.mux
}

func InitRouter() *Router {
	routerOnce.Do(initRouter)
	return routerInstance
}

// NewRouter returns an independent router instead of the shared singleton.
func NewRouter() *Router {
	return &Router{
		mux:    http.NewServeMux(),
		routes: make(map[string]routeEntry),
		config: DefaultServerConfig,
	}
}

type Route struct {
	// Path is a net/http ServeMux path pattern such as "/users/{id}". It must
	// not contain a method prefix ("GET /x"): methods are given via Method.
	Path              string
	Method            HTTPMethod
	Handler           http.HandlerFunc
	AdditionalMethods []HTTPMethod
	// Deprecated: use AdditionalMethods.
	AditionalMethods []HTTPMethod
}

func (r *Router) NewRoute(route Route) {
	if strings.ContainsAny(route.Path, " \t") {
		panic(fmt.Sprintf("Route path must not contain a method or whitespace: %q", route.Path))
	}
	if route.Handler == nil {
		panic(fmt.Sprintf("Route handler is nil: %s %s", route.Method, route.Path))
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	entry, exists := r.routes[route.Path]
	if !exists {
		entry = routeEntry{handlers: make(map[HTTPMethod]http.HandlerFunc)}
		r.mux.HandleFunc(route.Path, r.methodHandler(route.Path))
	} else {
		// Copy-on-write: in-flight requests keep reading the previous map.
		entry.handlers = maps.Clone(entry.handlers)
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
	allowedMethods := make([]string, 0, len(handlers)+2)
	for m := range handlers {
		allowedMethods = append(allowedMethods, string(m))
	}
	if _, ok := handlers[GET]; ok {
		if _, ok := handlers[HEAD]; !ok {
			allowedMethods = append(allowedMethods, string(HEAD))
		}
	}
	if _, ok := handlers[OPTIONS]; !ok {
		allowedMethods = append(allowedMethods, string(OPTIONS))
	}
	sort.Strings(allowedMethods)
	return strings.Join(allowedMethods, ", ")
}

func (r *Router) methodHandler(path string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		r.mu.RLock()
		route := r.routes[path]
		r.mu.RUnlock()

		var method HTTPMethod = HTTPMethod(req.Method)
		handler, exists := route.handlers[method]
		// net/http already discards response bodies for HEAD requests.
		if !exists && method == HEAD {
			handler, exists = route.handlers[GET]
		}
		if !exists && method == OPTIONS {
			w.Header().Set("Allow", route.allowedMethods)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if !exists {
			w.Header().Set("Allow", route.allowedMethods)
			body, _ := json.Marshal(map[string]string{"error": "Method not allowed"})
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write(body)
			return
		}
		w.Header().Set("Access-Control-Allow-Methods", route.allowedMethods)
		handler(w, req)
	}
}

func initRouter() {
	routerInstance = NewRouter()
}

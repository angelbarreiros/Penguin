package main

import (
	"angelotero/commonBackend/logger"
	"angelotero/commonBackend/router"
	"angelotero/commonBackend/router/handlers"
	"net/http"
	"time"

	_ "github.com/lib/pq"
)

func handlerUser(w http.ResponseWriter, r *http.Request) {
	var cache = handlers.NewCleanerCacheInstance()
	var key string = handlers.GenerateCacheKey(r)
	var data any
	var ok bool
	if data, ok = cache.Get(w, key); !ok {
		time.Sleep(5 * time.Second)
		cache.Set(key, handlers.NewCacheItem("User data", 40*time.Second))
		w.Write([]byte("User data"))
		return
	}
	var dateItem = data.(handlers.CacheItem)
	var result = dateItem.GetValue().(string)
	w.Write([]byte(result))

}
func main() {
	r := router.Router()
	r.NewRoute(router.Route{
		Path:             "/",
		Method:           "GET",
		Handler:          router.WithRateLimiting(router.WithLoggingMiddleware(handlerUser)),
		AditionalMethods: []router.HTTPMethod{router.HEAD, router.OPTIONS},
	})
	var log = logger.GetFileLogger()
	log.Configure("logs", "app.log", 10, 20)
	for range 10000000 {
		log.Info("This is an info message")
		log.Debug("This is a debug message")
		log.Warn("This is a warning message")
	}

	r.StartServer(":8080")

}

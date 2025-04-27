package router

import (
	"angelotero/commonBackend/logger"
	"angelotero/commonBackend/router/auth"
	"angelotero/commonBackend/router/cors"
	"context"
	"math"
	"net/http"
	"runtime/debug"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

func WithAuthMiddleWare(auth auth.PlainAuthInterface, hf handleFunc) handleFunc {
	return authMiddleWareFunc(auth)(hf)
}

func authMiddleWareFunc(auth auth.PlainAuthInterface) middlewareFunc {
	return func(hf handleFunc) handleFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if auth == nil {
				hf(w, r)
				return
			}
			if authorize, err := auth.Authorize(r); !authorize || err != nil {
				http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
				return

			}
			user, err := auth.GetUser(r)

			if err != nil {
				http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
				return
			}
			var ctx, cancel = context.WithTimeout(context.Background(), auth.GetTimeout())
			defer cancel()
			ctx = context.WithValue(r.Context(), auth.GetContextKey(), user)
			r = r.WithContext(ctx)
			hf(w, r)
		}
	}
}
func WithCorsMiddleware(corrsConfig *cors.CORSConfig, hf handleFunc) handleFunc {
	return corsMiddleware(corrsConfig)(hf)
}

func corsMiddleware(corrsConfig *cors.CORSConfig) middlewareFunc {
	return func(hf handleFunc) handleFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if corrsConfig == nil {
				hf(w, r)
				return
			}
			if r.Method == http.MethodOptions {
				w.Header().Set("Access-Control-Allow-Origin", strings.Join(corrsConfig.AllowedOrigins(), ","))
				w.Header().Set("Access-Control-Allow-Headers", strings.Join(corrsConfig.AllowedHeaders(), ","))
				w.Header().Set("Access-Control-Max-Age", strconv.Itoa(corrsConfig.MaxAge()))
				if corrsConfig.AllowCredentials() {
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}
				if corrsConfig.ExposedHeaders() != nil {
					w.Header().Set("Access-Control-Expose-Headers", strings.Join(corrsConfig.ExposedHeaders(), ","))
				}
				if corrsConfig.OptionsPassthrough() {
					hf(w, r)
					return
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}

			var origin string = r.Header.Get("Origin")
			var allAllowedOrigin bool = slices.Contains(corrsConfig.AllowedOrigins(), cors.AllowAllOrigin)
			if !allAllowedOrigin {
				if strings.TrimSpace(origin) == "" || !slices.Contains(corrsConfig.AllowedOrigins(), origin) {
					http.Error(w, "Origin not allowed", http.StatusForbidden)
					return
				}

			}
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(corrsConfig.AllowedHeaders(), ","))
			w.Header().Set("Access-Control-Max-Age", strconv.Itoa(corrsConfig.MaxAge()))
			if corrsConfig.AllowCredentials() {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
			if corrsConfig.ExposedHeaders() != nil {
				w.Header().Set("Access-Control-Expose-Headers", strings.Join(corrsConfig.ExposedHeaders(), ","))
			}
			hf(w, r)
		}
	}
}

func WithAuthAndRBAC(authType auth.RBACAuthInterface, roles []string, hf handleFunc) handleFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if authorize, err := authType.Authorize(r); !authorize || err != nil {
			http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}

		user, err := authType.GetUser(r)
		if err != nil {
			http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), authType.GetTimeout())
		defer cancel()
		ctx = context.WithValue(ctx, authType.GetContextKey(), user)
		r = r.WithContext(ctx)
		if !authType.RBAC(roles) {
			http.Error(w, "Forbidden: You don't have the required role", http.StatusForbidden)
			return
		}

		hf(w, r)
	}
}
func WithQueryParametersObligation(queryParameters []string, hf handleFunc) handleFunc {
	return queryParametersObligation(queryParameters)(hf)
}
func queryParametersObligation(queryParameters []string) middlewareFunc {
	return func(hf handleFunc) handleFunc {
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

type bucketOption func(*tokenBucket)
type tokenBucket struct {
	tokens        int32
	lastTime      time.Time
	mu            sync.Mutex
	startingLimit int32
	limitPerSec   float64
}

func WithCustomStartingLimit(limit int32) bucketOption {
	return func(tb *tokenBucket) {
		tb.startingLimit = limit
		tb.tokens = limit
	}
}

func WithCustomLimitPerSecond(limit float64) bucketOption {
	return func(tb *tokenBucket) {
		tb.limitPerSec = limit
	}
}

func WithRateLimiting(hf handleFunc, opts ...bucketOption) handleFunc {
	return rateLimiting(opts...)(hf)
}

func rateLimiting(opts ...bucketOption) middlewareFunc {
	var (
		tokenBuckets sync.Map
	)

	return func(hf handleFunc) handleFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			var ip string = r.Header.Get("X-Real-IP")
			if ip == "" {
				ip = r.Header.Get("X-Forwarded-For")
			}
			if ip == "" {
				ip = r.RemoteAddr
			}

			var bucketKey string = ip + ":" + r.URL.Path
			var now time.Time = time.Now()

			var bucket *tokenBucket = &tokenBucket{
				tokens:        30,
				startingLimit: 30,
				limitPerSec:   1.0,
				lastTime:      now,
			}

			for _, opt := range opts {
				opt(bucket)
			}

			b, _ := tokenBuckets.LoadOrStore(bucketKey, bucket)
			currentBucket := b.(*tokenBucket)

			currentBucket.mu.Lock()
			defer currentBucket.mu.Unlock()

			var elapsed float64 = now.Sub(currentBucket.lastTime).Seconds()
			var tokensToAdd float64 = elapsed * currentBucket.limitPerSec
			var tokensSinceLastRequest int32 = int32(math.Floor(tokensToAdd))
			currentBucket.tokens = min(currentBucket.startingLimit, currentBucket.tokens+tokensSinceLastRequest)
			currentBucket.lastTime = now

			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(int(currentBucket.startingLimit)))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(int(currentBucket.tokens)))

			if currentBucket.tokens > 0 {
				currentBucket.tokens--
				hf(w, r)
			} else {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
		}
	}
}
func WithLoggingMiddleware(hf handleFunc) handleFunc {
	return loggingMiddleware()(hf)
}

func loggingMiddleware() middlewareFunc {
	return func(hf handleFunc) handleFunc {
		var l = logger.GetConsoleLogger()
		return func(w http.ResponseWriter, r *http.Request) {
			var start time.Time = time.Now()
			hf(w, r)
			var duration time.Duration = time.Since(start)
			var method string = r.Method
			var path string = r.URL.Path
			var ip string = r.Header.Get("X-Real-IP")
			if ip == "" {
				ip = r.Header.Get("X-Forwarded-For")
			}
			if ip == "" {
				ip = r.RemoteAddr
			}
			l.Info("Method: %s, Path: %s, IP: %s, Duration: %s", method, path, ip, duration)
		}
	}
}
func WithRecovery(hf handleFunc) handleFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer handlePanic(w)
		hf(w, r)
	}
}

func handlePanic(w http.ResponseWriter) {
	if err := recover(); err != nil {
		logger.GetConsoleLogger().Error("Panic recovered: %v\nStack: %s", err, debug.Stack())
		logger.GetFileLogger().Error("Panic recovered: %v\nStack: %s", err, debug.Stack())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

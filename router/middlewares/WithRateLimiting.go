package middlewares

import (
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type bucketOption func(*tokenBucket)
type tokenBucket struct {
	tokens        int32
	lastTime      time.Time
	mu            sync.Mutex
	startingLimit int32
	limitPerSec   float64
}

func RateLimitOptStartingLimit(limit int32) bucketOption {
	return func(tb *tokenBucket) {
		tb.startingLimit = limit
		tb.tokens = limit
	}
}

func RateLimitOptLimitPerSecond(limit float64) bucketOption {
	return func(tb *tokenBucket) {
		tb.limitPerSec = limit
	}
}

func WithRateLimiting(hf http.HandlerFunc, opts ...bucketOption) http.HandlerFunc {
	return rateLimiting(opts...)(hf)
}

func rateLimiting(opts ...bucketOption) middlewareFunc {
	var (
		tokenBuckets sync.Map
	)

	return func(hf http.HandlerFunc) http.HandlerFunc {
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

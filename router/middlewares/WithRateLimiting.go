package middlewares

import (
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/angelbarreiros/Penguin/router/internal/clientip"
)

type bucketOption func(*rateLimitConfig)

type rateLimitConfig struct {
	startingLimit  int32
	limitPerSec    float64
	maxEntries     int
	trustedProxies []string
}

type tokenBucket struct {
	tokens   float64
	lastTime time.Time
}

func RateLimitOptStartingLimit(limit int32) bucketOption {
	return func(c *rateLimitConfig) {
		c.startingLimit = limit
	}
}

func RateLimitOptLimitPerSecond(limit float64) bucketOption {
	return func(c *rateLimitConfig) {
		c.limitPerSec = limit
	}
}

// RateLimitOptMaxEntries caps the number of tracked clients. When the table is
// full after evicting idle buckets, new clients are rejected with 429.
func RateLimitOptMaxEntries(maxEntries int) bucketOption {
	return func(c *rateLimitConfig) {
		c.maxEntries = maxEntries
	}
}

// RateLimitOptTrustedProxies lists proxy IPs/CIDRs whose X-Forwarded-For and
// X-Real-IP headers are honoured. By default those headers are ignored.
func RateLimitOptTrustedProxies(proxies []string) bucketOption {
	return func(c *rateLimitConfig) {
		c.trustedProxies = proxies
	}
}

func WithRateLimiting(hf http.HandlerFunc, opts ...bucketOption) http.HandlerFunc {
	return rateLimiting(opts...)(hf)
}

type rateLimiter struct {
	config    rateLimitConfig
	resolver  *clientip.Resolver
	mu        sync.Mutex
	buckets   map[string]*tokenBucket
	lastSweep time.Time
	// idleAfter is the time an empty bucket needs to refill completely; after
	// that it is indistinguishable from a new one and can be dropped.
	idleAfter time.Duration
}

func rateLimiting(opts ...bucketOption) middlewareFunc {
	config := rateLimitConfig{
		startingLimit: 30,
		limitPerSec:   1.0,
		maxEntries:    100_000,
	}
	for _, opt := range opts {
		opt(&config)
	}
	if config.startingLimit <= 0 || config.limitPerSec <= 0 || config.maxEntries <= 0 {
		panic("middlewares: rate limit values must be greater than zero")
	}

	limiter := &rateLimiter{
		config:    config,
		resolver:  clientip.NewResolver(config.trustedProxies),
		buckets:   make(map[string]*tokenBucket),
		lastSweep: time.Now(),
		idleAfter: time.Duration(float64(config.startingLimit) / config.limitPerSec * float64(time.Second)),
	}

	return func(hf http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			route := r.Pattern
			if route == "" {
				route = r.URL.Path
			}
			allowed, remaining := limiter.take(limiter.resolver.ClientIP(r)+"|"+route, time.Now())

			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(int(config.startingLimit)))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			if !allowed {
				writeJSONError(w, http.StatusTooManyRequests, "Rate limit exceeded")
				return
			}
			hf(w, r)
		}
	}
}

func (l *rateLimiter) take(key string, now time.Time) (bool, int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	sinceSweep := now.Sub(l.lastSweep)
	if sinceSweep >= l.idleAfter || (len(l.buckets) >= l.config.maxEntries && sinceSweep >= time.Second) {
		l.sweep(now)
	}

	capacity := float64(l.config.startingLimit)
	bucket, exists := l.buckets[key]
	if !exists {
		if len(l.buckets) >= l.config.maxEntries {
			return false, 0
		}
		bucket = &tokenBucket{tokens: capacity, lastTime: now}
		l.buckets[key] = bucket
	}

	elapsed := now.Sub(bucket.lastTime).Seconds()
	if elapsed > 0 {
		bucket.tokens = math.Min(capacity, bucket.tokens+elapsed*l.config.limitPerSec)
		bucket.lastTime = now
	}

	if bucket.tokens < 1 {
		return false, 0
	}
	bucket.tokens--
	return true, int(bucket.tokens)
}

func (l *rateLimiter) sweep(now time.Time) {
	for key, bucket := range l.buckets {
		if now.Sub(bucket.lastTime) >= l.idleAfter {
			delete(l.buckets, key)
		}
	}
	l.lastSweep = now
}

package middlewares

import (
	"container/list"
	"math"
	"net/http"
	"net/netip"
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
	proxyHeader    []string
	ipv6PrefixLen  int
}

type tokenBucket struct {
	key      string
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
// full, the least recently seen client is forgotten (and starts again with a
// full bucket) so that filling the table never locks out new clients.
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

// RateLimitOptTrustedProxyHeader names the single forwarding header
// ("X-Forwarded-For" or "X-Real-IP") the trusted proxies set. Without it both
// are read; see clientip.NewResolver.
func RateLimitOptTrustedProxyHeader(header string) bucketOption {
	return func(c *rateLimitConfig) {
		c.proxyHeader = []string{header}
	}
}

// RateLimitOptIPv6PrefixLength groups IPv6 clients by network prefix
// (default 64), since a single host usually controls a whole /64.
func RateLimitOptIPv6PrefixLength(bits int) bucketOption {
	return func(c *rateLimitConfig) {
		c.ipv6PrefixLen = bits
	}
}

func WithRateLimiting(hf http.HandlerFunc, opts ...bucketOption) http.HandlerFunc {
	return rateLimiting(opts...)(hf)
}

type rateLimiter struct {
	config   rateLimitConfig
	resolver *clientip.Resolver
	mu       sync.Mutex
	buckets  map[string]*list.Element
	// lru orders buckets from most (front) to least (back) recently used.
	lru       *list.List
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
		ipv6PrefixLen: 64,
	}
	for _, opt := range opts {
		opt(&config)
	}
	if config.startingLimit <= 0 || config.limitPerSec <= 0 || config.maxEntries <= 0 {
		panic("middlewares: rate limit values must be greater than zero")
	}
	if config.ipv6PrefixLen <= 0 || config.ipv6PrefixLen > 128 {
		panic("middlewares: IPv6 prefix length must be between 1 and 128")
	}

	limiter := &rateLimiter{
		config:    config,
		resolver:  clientip.NewResolver(config.trustedProxies, config.proxyHeader...),
		buckets:   make(map[string]*list.Element),
		lru:       list.New(),
		lastSweep: time.Now(),
		idleAfter: time.Duration(float64(config.startingLimit) / config.limitPerSec * float64(time.Second)),
	}

	return func(hf http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			route := r.Pattern
			if route == "" {
				route = r.URL.Path
			}
			allowed, remaining := limiter.take(limiter.clientKey(r)+"|"+route, time.Now())

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

// clientKey is the client IP, with IPv6 addresses reduced to their network
// prefix so rotating through a subnet does not yield fresh buckets.
func (l *rateLimiter) clientKey(r *http.Request) string {
	ip := l.resolver.ClientIP(r)
	addr, err := netip.ParseAddr(ip)
	if err != nil || !addr.Is6() {
		return ip
	}
	prefix, err := addr.WithZone("").Prefix(l.config.ipv6PrefixLen)
	if err != nil {
		return ip
	}
	return prefix.String()
}

func (l *rateLimiter) take(key string, now time.Time) (bool, int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if now.Sub(l.lastSweep) >= l.idleAfter {
		l.sweep(now)
	}

	capacity := float64(l.config.startingLimit)
	var bucket *tokenBucket
	if element, exists := l.buckets[key]; exists {
		l.lru.MoveToFront(element)
		bucket = element.Value.(*tokenBucket)
	} else {
		if len(l.buckets) >= l.config.maxEntries {
			oldest := l.lru.Back()
			l.lru.Remove(oldest)
			delete(l.buckets, oldest.Value.(*tokenBucket).key)
		}
		bucket = &tokenBucket{key: key, tokens: capacity, lastTime: now}
		l.buckets[key] = l.lru.PushFront(bucket)
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

// sweep drops idle buckets, walking from the least recently used end.
func (l *rateLimiter) sweep(now time.Time) {
	for element := l.lru.Back(); element != nil; element = l.lru.Back() {
		bucket := element.Value.(*tokenBucket)
		if now.Sub(bucket.lastTime) < l.idleAfter {
			break
		}
		l.lru.Remove(element)
		delete(l.buckets, bucket.key)
	}
	l.lastSweep = now
}

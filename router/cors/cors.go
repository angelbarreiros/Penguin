package cors

import (
	"slices"
)

const (
	AllowAllOrigin = "*"
)

type CORSConfig struct {
	allowedOrigins     []string
	allowedHeaders     []string
	exposedHeaders     []string
	allowCredentials   bool
	maxAge             int
	optionsPassthrough bool
}

func (c *CORSConfig) AllowedOrigins() []string {
	return c.allowedOrigins
}

func (c *CORSConfig) AllowedHeaders() []string {
	return c.allowedHeaders
}

func (c *CORSConfig) ExposedHeaders() []string {
	return c.exposedHeaders
}

func (c *CORSConfig) AllowCredentials() bool {
	return c.allowCredentials
}

func (c *CORSConfig) MaxAge() int {
	return c.maxAge
}

func (c *CORSConfig) OptionsPassthrough() bool {
	return c.optionsPassthrough
}

// AllowsAllOrigins reports whether the wildcard origin is configured.
func (c *CORSConfig) AllowsAllOrigins() bool {
	return slices.Contains(c.allowedOrigins, AllowAllOrigin)
}

// IsOriginAllowed reports whether origin may access the resource.
func (c *CORSConfig) IsOriginAllowed(origin string) bool {
	if origin == "" {
		return false
	}
	return c.AllowsAllOrigins() || slices.Contains(c.allowedOrigins, origin)
}

// NewCORSConfig panics when AllowAllOrigin (*) is combined with
// AllowCredentials=true: browsers reject that combination, and the usual
// workaround (reflecting any origin with credentials) is a security hole.
func NewCORSConfig(options ...func(*CORSConfig)) *CORSConfig {
	var config *CORSConfig = &CORSConfig{
		allowedOrigins: []string{},

		allowedHeaders: []string{
			"Origin",
			"Accept",
			"X-Requested-With",
			"Content-Type",
			"Authorization",
		},
		exposedHeaders:     []string{},
		allowCredentials:   false,
		maxAge:             300, // 5 minutes
		optionsPassthrough: false,
	}

	for _, option := range options {
		option(config)
	}

	if config.allowCredentials && config.AllowsAllOrigins() {
		panic("cors: AllowAllOrigin (*) cannot be combined with AllowCredentials=true; list the allowed origins explicitly")
	}

	return config
}

func WithAllowedOrigins(origins []string) func(*CORSConfig) {
	return func(c *CORSConfig) {
		c.allowedOrigins = origins
	}
}

func WithAllowedHeaders(headers []string) func(*CORSConfig) {
	return func(c *CORSConfig) {
		c.allowedHeaders = headers
	}
}

func WithExposedHeaders(headers []string) func(*CORSConfig) {
	return func(c *CORSConfig) {
		c.exposedHeaders = headers
	}
}

func WithAllowCredentials(allow bool) func(*CORSConfig) {
	return func(c *CORSConfig) {
		c.allowCredentials = allow
	}
}

func WithMaxAge(age int) func(*CORSConfig) {
	return func(c *CORSConfig) {
		c.maxAge = age
	}
}

func WithOptionsPassthrough(passthrough bool) func(*CORSConfig) {
	return func(c *CORSConfig) {
		c.optionsPassthrough = passthrough
	}
}

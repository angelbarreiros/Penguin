package cors

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

func NewCORSConfig(options ...func(*CORSConfig)) *CORSConfig {
	var config *CORSConfig = &CORSConfig{
		allowedOrigins: []string{},

		allowedHeaders: []string{
			"Origin",
			"Accept",
			"X-Requested-With",
		},
		exposedHeaders:     []string{},
		allowCredentials:   false,
		maxAge:             300, // 5 minutes
		optionsPassthrough: false,
	}

	for _, option := range options {
		option(config)
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

# Penguin Library Documentation

This library focuses on routes, where each route is configured individually and in detail, allowing for precise control over HTTP handling.

## Table of Contents

- [Router Package](#router-package)
- [Logger Package](#logger-package)
- [Scheduler Package](#scheduler-package)

---

## Router Package

The `router` package provides an HTTP router for Go applications, allowing you to register routes and handle HTTP requests.

### Router Table of Contents

- [Auth Package](#auth-package)
- [CORS Package](#cors-package)
- [Middlewares Package](#middlewares-package)
- [Helpers Package](#helpers-package)
- [Types Package](#types-package)

### Functions

#### InitRouter()
Returns the router instance.

```go
router := router.InitRouter()
```

#### NewRoute(route Route)
Registers a new route with a specific path, HTTP method, and handler function. You can also specify additional methods for the same path.

Parameters:
- `route`: A `Route` struct containing:
  - `Path`: The URL path (string).
  - `Method`: The primary HTTP method (HTTPMethod).
  - `Handler`: The function to handle the request (HandleFunc).
  - `AditionalMethods`: Optional additional HTTP methods for the same path ([]HTTPMethod).

Example:
```go
route := router.Route{
    Path:    "/api/users",
    Method:  router.GET,
    Handler: func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Users list"))
    },
}
router.NewRoute(route)
```

With additional methods:
```go
route := router.Route{
    Path:             "/api/users",
    Method:           router.POST,
    Handler:          userHandler,
    AditionalMethods: []router.HTTPMethod{router.PUT, router.DELETE},
}
router.NewRoute(route)
```

#### StartServer(s string)
Starts the HTTP server on the specified address.

Parameters:
- `s`: The server address (e.g., ":8080").

Example:
```go
router.StartServer(":8080")
```

### HTTP Methods
Supported HTTP methods:
- `GET`
- `POST`
- `PUT`
- `DELETE`
- `PATCH`
- `OPTIONS`
- `HEAD`

---

## Auth Package

The `auth` package provides JWT-based authentication, supporting plain JWT and JWT with RBAC.

### Functions

#### LoadPrivateKeyFromFile(keyPem []byte) (*ecdsa.PrivateKey, error)
Loads an ECDSA private key from PEM bytes.

Example:
```go
keyPem := []byte("-----BEGIN EC PRIVATE KEY-----\n...\n-----END EC PRIVATE KEY-----")
privateKey, err := auth.LoadPrivateKeyFromFile(keyPem)
```

#### NewJwtAuth(secret *ecdsa.PrivateKey, claimsType plainClaimsInterface) *JwtAuth
Creates a JWT auth instance.

Example:
```go
jwtAuth := auth.NewJwtAuth(privateKey, &auth.PlainClaims{})
```

#### NewJwtAuthWithRbac(secret *ecdsa.PrivateKey, claimsType rBACClaimsInterface) *RBACJwtAuth
Creates a JWT auth instance with RBAC.

Example:
```go

rbacAuth := auth.NewJwtAuthWithRbac(privateKey, &auth.RBACClaims{})
```


### Creating Claims

#### PlainClaims
To create a `PlainClaims` instance:

```go
claims := &auth.PlainClaims{
    RegisteredClaims: jwt.RegisteredClaims{
        Issuer:    "your-issuer",
        Subject:   "user-id",
        ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
        IssuedAt:  jwt.NewNumericDate(time.Now()),
    },
}
```

#### RBACClaims
To create a `RBACClaims` instance with roles:

```go
claims := &auth.RBACClaims{
    RegisteredClaims: jwt.RegisteredClaims{
        Issuer:    "your-issuer",
        Subject:   "user-id",
        ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
        IssuedAt:  jwt.NewNumericDate(time.Now()),
    },
    Roles: []string{"admin", "user"},
}
```

---

## CORS Package

The `cors` package allows you to create CORS configurations for your routes.

### Functions

#### NewCORSConfig(options ...func(*CORSConfig)) *CORSConfig
Creates a new CORS configuration.

Example:
```go
config := cors.NewCORSConfig(
    cors.WithAllowedOrigins([]string{"http://localhost:3000"}),
    cors.WithAllowedHeaders([]string{"Content-Type", "Authorization"}),
)
```

#### WithAllowedOrigins(origins []string) func(*CORSConfig)
Sets allowed origins.

#### WithAllowedHeaders(headers []string) func(*CORSConfig)
Sets allowed headers.

#### WithExposedHeaders(headers []string) func(*CORSConfig)
Sets exposed headers.

#### WithAllowCredentials(allow bool) func(*CORSConfig)
Sets allow credentials.

#### WithMaxAge(age int) func(*CORSConfig)
Sets max age for preflight.

#### WithOptionsPassthrough(passthrough bool) func(*CORSConfig)
Sets options passthrough.

---

## Middlewares Package

The `middlewares` package provides abstract middlewares for common HTTP functionalities. For `WithAuthMiddleWare` and `WithCors`, you need to use the configurations from the `auth` and `cors` packages respectively.

### Functions

#### WithAuthMiddleWare(auth auth.PlainAuthInterface, hf handleFunc) handleFunc
Applies plain JWT authentication.

Example:
```go
handler := middlewares.WithAuthMiddleWare(jwtAuth, func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Authenticated"))
})
```

#### WithAuthAndRBAC(authType auth.RBACAuthInterface, roles []string, hf handleFunc) handleFunc
Applies JWT authentication with RBAC.

Example:
```go
handler := middlewares.WithAuthAndRBAC(rbacAuth, []string{"admin"}, func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Authorized"))
})
```

#### WithCors(corrsConfig *cors.CORSConfig, hf handleFunc) handleFunc
Applies CORS configuration.

Example:
```go
config := cors.NewCORSConfig(cors.WithAllowedOrigins([]string{"*"}))
handler := middlewares.WithCors(config, func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("CORS enabled"))
})
```

#### WithLogging(hf handleFunc) handleFunc
Logs HTTP requests.

Example:
```go
handler := middlewares.WithLogging(func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Logged"))
})
```

#### WithRecovery(hf handleFunc) handleFunc
Recovers from panics.

Example:
```go
handler := middlewares.WithRecovery(func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Safe"))
})
```

#### WithRateLimiting(hf handleFunc, opts ...bucketOption) handleFunc
Limits request rate.

Example:
```go
handler := middlewares.WithRateLimiting(func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Limited"))
}, middlewares.RateLimitOptStartingLimit(10), middlewares.RateLimitOptLimitPerSecond(2.0))
```

#### WithQueryParametersObligation(queryParameters []string, hf handleFunc) handleFunc
Requires query parameters.

Example:
```go
handler := middlewares.WithQueryParametersObligation([]string{"id"}, func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Params checked"))
})
```

#### WithFeatureEnable(env string, hf handleFunc) handleFunc
Enables feature based on env var. The feature name must be set as an environment variable on the machine.

Example:
```go
// Set env var: export ENABLE_FEATURE=true
handler := middlewares.WithFeatureEnable("ENABLE_FEATURE", func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Feature enabled"))
})
```

#### WithFeatureEnabledByHeader(header string, hf handleFunc) handleFunc
Enables feature based on header. The header must be present in the request.

Example:
```go
// Include header: X-Feature: true
handler := middlewares.WithFeatureEnabledByHeader("X-Feature", func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Header enabled"))
})
```

---

## Helpers Package

The `helpers` package provides utility functions for common HTTP operations.

### Cache

#### NewCacheItem(value any, expiration time.Duration) CacheItem
Creates a cache item.

Example:
```go
item := helpers.NewCacheItem(data, time.Hour)
```

#### NewCleanerCacheInstance() *cleanerCache
Gets singleton cache instance.

Example:
```go
cache := helpers.NewCleanerCacheInstance()
```

#### Set(key string, i CacheItem)
Stores item in cache with expiration.

Example:
```go
cache.Set("key", item)
```

#### Get(w http.ResponseWriter, key string) (CacheItem, bool)
Retrieves item from cache, sets HTTP headers.

Example:
```go
item, found := cache.Get(w, "key")
```

#### GenerateCacheKey(r *http.Request) string
Generates cache key from request.

Example:
```go
key := helpers.GenerateCacheKey(r)
```

### Body

#### DeserializeBodyWithLimit(r *http.Request, dto any, maxBytes int64) error
Deserializes JSON body with size limit.

Example:
```go
var user User
err := helpers.DeserializeBodyWithLimit(r, &user, 1024)
```

### Request

#### Context

##### GetContextValue[T any](r *http.Request, key string) (T, error)
Gets typed value from request context.

Example:
```go
user, err := helpers.GetContextValue[User](r, "user")
```

#### Pagination

##### GetPaginationParams(r *http.Request, defaultPageSize uint) PaginationParams
Gets pagination params from query.

Example:
```go
params := helpers.GetPaginationParams(r, 10)
```

#### Path Values

##### GetUintPathValue(identifier string, r *http.Request) (uint64, error)
Gets uint from path value.

Example:
```go
id, err := helpers.GetUintPathValue("id", r)
```

##### GetUUidPathValue(identifier string, r *http.Request) (uuid.UUID, error)
Gets UUID from path value.

Example:
```go
id, err := helpers.GetUUidPathValue("uuid", r)
```

##### GetStringPathValue(identifier string, r *http.Request, options ...MaxStringLengthOption) (string, error)
Gets sanitized string from path value.

Example:
```go
name, err := helpers.GetStringPathValue("name", r)
```

##### WithCustomLengthOption(length int) MaxStringLengthOption
Sets custom max length for string.

Example:
```go
name, err := helpers.GetStringPathValue("name", r, helpers.WithCustomLengthOption(100))
```

#### Query Params

##### GetNullUint64QueryParam(queryParameter string, r *http.Request) (sql.NullInt64, error)
Gets nullable uint64 from query.

Example:
```go
val, err := helpers.GetNullUint64QueryParam("limit", r)
```

##### GetNullInt64QueryParam(queryParameter string, r *http.Request) (sql.NullInt64, error)
Gets nullable int64 from query.

Example:
```go
val, err := helpers.GetNullInt64QueryParam("offset", r)
```

##### GetNullBoolQueryParam(queryParameter string, r *http.Request) (sql.NullBool, error)
Gets nullable bool from query.

Example:
```go
val, err := helpers.GetNullBoolQueryParam("active", r)
```

##### GetNullUUIDQueryParam(queryParameter string, r *http.Request) (uuid.NullUUID, error)
Gets nullable UUID from query.

Example:
```go
val, err := helpers.GetNullUUIDQueryParam("userId", r)
```

##### GetNullTimeQueryParam(queryParameter string, r *http.Request) (sql.NullTime, error)
Gets nullable time from query.

Example:
```go
val, err := helpers.GetNullTimeQueryParam("createdAt", r)
```

##### GetNullStringQueryParam(queryParameter string, r *http.Request) (sql.NullString, error)
Gets nullable string from query.

Example:
```go
val, err := helpers.GetNullStringQueryParam("search", r)
```

### Response

#### SendSuccessResponse(w http.ResponseWriter, data any)
Sends JSON success response.

Example:
```go
helpers.SendSuccessResponse(w, data)
```

#### SendValidationErrorResponse(w http.ResponseWriter, errors []string)
Sends validation errors.

Example:
```go
helpers.SendValidationErrorResponse(w, []string{"Invalid input"})
```

#### SendErrorResponse(w http.ResponseWriter, statusCode int, message string)
Sends error response.

Example:
```go
helpers.SendErrorResponse(w, 500, "Internal error")
```

#### SendNoContentResponse(w http.ResponseWriter)
Sends 204 No Content.

Example:
```go
helpers.SendNoContentResponse(w)
```

### Tokens

#### GenerateJwtToken(claims jwt.Claims, secret *ecdsa.PrivateKey) (string, error)
Generates JWT token.

Example:
```go
token, err := helpers.GenerateJwtToken(claims, privateKey)
```

---

*This documentation provides an overview of the Penguin library packages. For more details, refer to the source code.*

## Types Package

The `types` package (part of the `router` package) provides custom types for handling dates and times in JSON serialization.

### Types

#### Date
Represents a date with year, month, and day.

Fields:
- `Year int`
- `Month int`
- `Day int`

Methods:
- `ToString() string`: Converts to string format "2006-01-02" (makes it stringeable).
- `Marshal() ([]byte, error)`: Marshals to JSON (makes it JSON compatible).
- `Unmarshal(data []byte) error`: Unmarshals from JSON (makes it JSON compatible).

#### TimeStamp
Represents a timestamp with date, time, and timezone.

Fields:
- `Year int`
- `Month int`
- `Day int`
- `Hour int`
- `Minute int`
- `Second int`
- `Microsecond int`
- `Timezone string`

Methods:
- `ToString() string`: Converts to string format "2006-01-02 15:04:05.000000-07" (makes it stringeable).
- `Marshal() ([]byte, error)`: Marshals to JSON (makes it JSON compatible).
- `Unmarshal(data []byte) error`: Unmarshals from JSON (makes it JSON compatible).

#### Time
Represents a time with hour, minute, second, and millisecond.

Fields:
- `Hour int`
- `Minute int`
- `Second int`
- `Milisecond int`

Methods:
- `ToString() string`: Converts to string format "15:04:05.000" (makes it stringeable).
- `Marshal() ([]byte, error)`: Marshals to JSON (makes it JSON compatible).
- `Unmarshal(data []byte) error`: Unmarshals from JSON (makes it JSON compatible).

---

## Logger Package

The `logger` package provides logging functionality with file and console loggers.

### Functions

#### GetFileLogger() *FileLogger
Gets the singleton file logger instance.

#### GetConsoleLogger() *ConsoleLogger
Gets the singleton console logger instance.

### FileLogger Methods

#### Configuration Methods

##### Configure(logDir string, baseFileName string, maxFileSizeMB int, maxAgeDays int)
Configures the file logger.

##### SetLevel(level LogLevel)
Sets the log level.

#### Logging Methods

##### Debug(msg string, args ...any)
Logs a debug message.

##### Info(msg string, args ...any)
Logs an info message.

##### Warn(msg string, args ...any)
Logs a warning message.

##### Error(msg string, args ...any)
Logs an error message.

##### Fatal(msg string, args ...any)
Logs a fatal message and exits.

### ConsoleLogger Methods

#### Configuration Methods

##### SetLevel(level LogLevel)
Sets the log level.

##### EnableColors(enabled bool)
Enables or disables colored output.

#### Logging Methods

##### Debug(msg string, args ...any)
Logs a debug message to console.

##### Info(msg string, args ...any)
Logs an info message to console.

##### Warn(msg string, args ...any)
Logs a warning message to console.

##### Error(msg string, args ...any)
Logs an error message to console.

##### Fatal(msg string, args ...any)
Logs a fatal message to console and exits.

### Types

- `LogLevel`: Enum for log levels (DEBUG, INFO, WARN, ERROR, FATAL).
- `Logger`: Interface for loggers.

---

## Scheduler Package

The `scheduler` package provides job scheduling functionality.

### Functions

#### StartScheduler() *Scheduler
Starts the singleton scheduler instance.

#### JobFunction(f JobFuncInterface) *jobFunction
Creates a job function.

### Scheduler Methods

#### Configuration and Usage Methods

##### ScheduleJob(when time.Time, jobFunction *jobFunction) (uint64, error)
Schedules a one-time job.

##### ScheduleIntervalJob(interval time.Duration, jobFunction *jobFunction) (uint64, error)
Schedules a recurring job.

##### RemoveJob(id uint64) error
Removes a job.

##### PauseJob(id uint64) error
Pauses a job.

##### UnPauseJob(id uint64) error
Unpauses a job.

##### GetChannel(id uint64) chan []any
Gets the return channel for a job.

##### Stop()
Stops the scheduler.

##### IsRunning() bool
Checks if the scheduler is running.

##### Pause()
Pauses the scheduler.

##### UnPause()
Unpauses the scheduler.

### jobFunction Methods

#### WithReturnChannel() (*jobFunction, chan []any)
Adds a return channel to the job function.

### Types

- `JobFuncInterface`: Interface for job functions.










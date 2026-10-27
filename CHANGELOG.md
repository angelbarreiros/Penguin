# Changelog

All notable changes to this project are documented in this file. The format
is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the
project uses [Semantic Versioning](https://semver.org/).

## [v0.2.0] - 2026-10-27

This release fixes the findings of two security and bug audits.

### Breaking changes

- The auth middlewares store the user under the typed key
  `auth.UserContextKey` instead of the string `"user"`. To keep reading
  `r.Context().Value("user")`, pass
  `auth.WithCustomContextKey(auth.DefaultContextKey)` (or
  `auth.JwtAuthRbacWithCustomContextKey`). `DefaultContextKey` is deprecated.
- `RBACAuthInterface.RBAC` now takes `(user any, allowedRoles []string)`.
- `WithAuthMiddleware(nil, …)` and `WithAuthAndRBAC(nil, …)` panic.
- `NewCORSConfig` panics when `*` is combined with `WithAllowCredentials(true)`.
- Tokens without `exp`, or signed with an algorithm that does not match the
  key's curve, are rejected.
- `GetNullBytePathValue` parses a number from 0 to 255.

### Added

- `Route.CORS`: wraps the handler with `WithCors` and lets the router answer
  CORS preflights for the path.
- `RateLimitOptIPv6PrefixLength` (default 64) and
  `RateLimitOptTrustedProxyHeader`.
- `WithLoggingTrustedProxyHeader` to read only `X-Forwarded-For` or only
  `X-Real-IP`.
- `cleanerCache.SetMaxEntries` and `helpers.DefaultResponseCacheMaxEntries`
  (10,000) for the response cache.
- `GetPaginationParamsWithLimits` to clamp the page number.
- `ConsoleLogger.Flush`.
- `NewRouter`, `SetServerConfig`, `Shutdown` and `Handler` on the router, with
  server timeouts by default.
- Apache 2.0 license.

### Changed

- When the rate limiter's client table is full, the least recently seen client
  is forgotten instead of new clients getting 429.
- `GenerateCacheKey` includes the host and a hash of the `Cookie` header.
- The response cache sends `Cache-Control: private` and
  `Vary: Accept, Accept-Language, Authorization, Cookie` on misses as well as
  hits.
- Keyed caches use one sweep per minute instead of a scheduler job per entry.
  Closed caches ignore writes, and the shared getters return a new cache after
  a close.
- `WithCors` passes same-host requests through instead of returning 403.
- `Access-Control-Allow-Methods` is sent only on CORS preflights.
- The console logger writes from a background goroutine and drops lines when
  more than 4,096 are waiting.
- Scheduler start, sleep and wake messages are logged at DEBUG.

### Fixed

- Security: concurrent requests could swap identities in `JwtAuth` and
  `RBACJwtAuth`, which stored the token and claims on the shared struct.
- Security: attackers could fill the rate limiter's client table and lock out
  every new client.
- Security: rotating IPv6 addresses bypassed the per-IP rate limit.
- Security: the response cache could serve one user's or tenant's response to
  another.
- Security: the response cache could grow without bound from varied query
  strings.
- Security: forwarding headers were trusted from any peer. They are now read
  only from trusted proxies.
- CORS preflights never reached `WithCors` unless the route registered
  `OPTIONS`.
- i18n placeholders inside an argument were substituted again.
- `WithLogging` and `WithRecovery` hid `http.Flusher` and `http.Hijacker`,
  which broke SSE and WebSockets.
- Daily scheduler jobs ignored the time zone of `at` and shifted by an hour on
  DST change days.
- Removing or stopping a running scheduler job panicked, and a panicking job
  crashed the process.
- Also fixed: race-free route registration, early cache eviction and leaks,
  anchored numeric validation and uint64 overflow, body, response and recovery
  error handling, `DateRange` and PostgreSQL type fixes, i18n caching and
  q-values, and logger permissions, rotation and pruning.

[v0.2.0]: https://github.com/angelbarreiros/Penguin/compare/v0.1.41...v0.2.0

# gort/gin

Authentication and maintenance middleware for the [Gin](https://github.com/gin-gonic/gin) web framework.

## Installation

```
go get github.com/louvri/gort/gin
```

## Middleware

### `JWTAuthValidatorMiddleware`

```go
func JWTAuthValidatorMiddleware(key, unauthorizedErrorMessage string, symmetric, logErrorMessage bool) gin.HandlerFunc
```

Validates JWT bearer tokens on incoming requests. Returns `401 Unauthorized` if the token is missing, expired, or invalid.

- `key` — signing key (shared secret for HMAC, PEM public key for RSA)
- `unauthorizedErrorMessage` — message returned to the client on auth failure
- `symmetric` — `true` for HMAC, `false` for RSA
- `logErrorMessage` — `true` to log detailed errors to stdout

```go
r := gin.Default()

// Public routes with JWT auth
public := r.Group("/api")
public.Use(gortGin.JWTAuthValidatorMiddleware("your-secret", "Invalid token", true, true))
public.GET("/profile", profileHandler)
```

### `ServerKeyAuthValidatorMiddleware`

```go
func ServerKeyAuthValidatorMiddleware(headerKey, serverKey, expiringServerKey, unauthorizedErrorMessage string) gin.HandlerFunc
```

Validates requests using a static API key in a custom header. Supports key rotation via an expiring key.

- `headerKey` — header name to read the key from
- `serverKey` — primary API key
- `expiringServerKey` — secondary key for rotation (both are accepted)
- `unauthorizedErrorMessage` — message returned on auth failure

```go
// Internal routes with server key auth
internal := r.Group("/internal")
internal.Use(gortGin.ServerKeyAuthValidatorMiddleware(
    "X-Server-Token", "primary-key", "old-key", "Unauthorized",
))
internal.GET("/health", healthHandler)
```

### `ProbeMaintenanceMiddleware`

```go
func ProbeMaintenanceMiddleware(skippedPaths []string, errorMessage string, statusCode int, enabled bool) gin.HandlerFunc
```

Blocks all requests when maintenance mode is enabled, except for paths matching the skip list.

- `skippedPaths` — path prefixes that remain accessible during maintenance (e.g., `"/health"`)
- `errorMessage` — message returned to blocked requests
- `statusCode` — HTTP status code for blocked requests
- `enabled` — `true` to activate maintenance mode

```go
r.Use(gortGin.ProbeMaintenanceMiddleware(
    []string{"/health", "/ready"},
    "Service is under maintenance",
    http.StatusServiceUnavailable,
    true,
))
```

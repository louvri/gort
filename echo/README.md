# gort/echo

Authentication and maintenance middleware for the [Echo](https://github.com/labstack/echo) web framework.

## Installation

```
go get github.com/louvri/gort/echo
```

## Middleware

### `JWTAuthValidatorMiddleware`

```go
func JWTAuthValidatorMiddleware(key, unauthorizedErrorMessage string, symmetric, logErrorMessage bool) echo.MiddlewareFunc
```

Validates JWT bearer tokens on incoming requests. Returns `401 Unauthorized` if the token is missing, expired, or invalid.

- `key` — signing key (shared secret for HMAC, PEM public key for RSA)
- `unauthorizedErrorMessage` — message returned to the client on auth failure
- `symmetric` — `true` for HMAC, `false` for RSA
- `logErrorMessage` — `true` to log detailed errors to stdout

```go
e := echo.New()

// Public routes with JWT auth
public := e.Group("/api")
public.Use(gortEcho.JWTAuthValidatorMiddleware("your-secret", "Invalid token", true, true))
public.GET("/profile", profileHandler)
```

### `ServerKeyAuthValidatorMiddleware`

```go
func ServerKeyAuthValidatorMiddleware(headerKey, serverKey, expiringServerKey, unauthorizedErrorMessage string) echo.MiddlewareFunc
```

Validates requests using a static API key in a custom header. Supports key rotation via an expiring key.

- `headerKey` — header name to read the key from
- `serverKey` — primary API key
- `expiringServerKey` — secondary key for rotation (both are accepted)
- `unauthorizedErrorMessage` — message returned on auth failure

```go
// Internal routes with server key auth
internal := e.Group("/internal")
internal.Use(gortEcho.ServerKeyAuthValidatorMiddleware(
    "X-Server-Token", "primary-key", "old-key", "Unauthorized",
))
internal.GET("/health", healthHandler)
```

### `ProbeMaintenanceMiddleware`

```go
func ProbeMaintenanceMiddleware(skippedPaths []string, errorMessage string, statusCode int, enabled bool) echo.MiddlewareFunc
```

Blocks all requests when maintenance mode is enabled, except for paths matching the skip list.

- `skippedPaths` — path prefixes that remain accessible during maintenance (e.g., `"/health"`)
- `errorMessage` — message returned to blocked requests
- `statusCode` — HTTP status code for blocked requests
- `enabled` — `true` to activate maintenance mode

```go
e.Use(gortEcho.ProbeMaintenanceMiddleware(
    []string{"/health", "/ready"},
    "Service is under maintenance",
    http.StatusServiceUnavailable,
    true,
))
```

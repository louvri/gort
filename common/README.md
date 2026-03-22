# gort/common

Framework-agnostic utilities for JWT handling, bearer token extraction, and timezone parsing.

## Installation

```
go get github.com/louvri/gort/common
```

## API

### Authentication

#### `GetAuthorizationHeaderValue(r *http.Request) string`

Returns the value of the `Authorization` header.

#### `GetBearerToken(r *http.Request) string`

Extracts and returns the token from a `Bearer <token>` authorization header. Returns an empty string if the header is missing or uses a different scheme.

#### `JWTKeyFunc(key string, symmetric bool) jwt.Keyfunc`

Returns a `jwt.Keyfunc` for use with `jwt.Parse`. Validates that the signing algorithm matches the key type (HMAC for symmetric, RSA for asymmetric) to prevent algorithm confusion attacks.

- `symmetric=true` — expects HMAC-signed tokens, uses `key` as the shared secret
- `symmetric=false` — expects RSA-signed tokens, parses `key` as a PEM-encoded RSA public key

#### `GetMapClaimsFromJWT(key, bearerToken string, symmetric bool) (jwt.MapClaims, error)`

Parses and validates a JWT token, returning its claims as a map. Returns an error if the token is invalid, expired, or uses an unexpected signing algorithm.

```go
claims, err := common.GetMapClaimsFromJWT("your-secret", token, true)
if err != nil {
    // handle error
}
userID := claims["sub"].(string)
```

#### `GetMapClaimsFromJWTWithoutValidation(bearerToken string) jwt.MapClaims`

Extracts claims from a JWT without validating the signature. Useful for reading token metadata before validation. Returns `nil` for malformed tokens.

```go
claims := common.GetMapClaimsFromJWTWithoutValidation(token)
if claims != nil {
    userID := claims["sub"].(string)
}
```

### Datetime

#### `ParseTimeWithFallback(sourceTime, format, backupFormat string, location *time.Location) time.Time`

Parses a time string using the primary format, falling back to the backup format on failure. Returns zero time if both fail.

#### `ParseUTCTime(sourceTime string) time.Time`

Parses a UTC time string supporting both `"2006-01-02 15:04:05"` and `"2006-01-02T15:04:05Z"` formats.

#### `ExtractTimeZoneTextFromHeader(r *http.Request) string`

Returns the raw value of the `Timezone` header.

#### `ExtractTimeZoneInSecondsFromHeader(r *http.Request) int`

Parses the `Timezone` header (e.g., `GMT+07`, `UTC-0530`) and returns the offset in seconds.

#### `ExtractTimeZoneLocationFromHeader(r *http.Request) *time.Location`

Returns a `*time.Location` from the `Timezone` header. Defaults to `time.UTC` if the header is absent.

```go
loc := common.ExtractTimeZoneLocationFromHeader(r)
localTime := time.Now().In(loc)
```

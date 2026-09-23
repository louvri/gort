// Package echo provides authentication and maintenance-mode middleware for
// the Echo web framework.
package echo

import (
	"crypto/subtle"
	"log/slog"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// JWTAuthValidatorMiddleware rejects requests without a valid "Bearer" JWT
// with 401 and unauthorizedErrorMessage. The signing algorithm is pinned to
// the key type: symmetric expects HMAC with key as the shared secret;
// otherwise it expects RSA with key as a PEM-encoded public key.
// logErrorMessage logs parse and verification failures.
func JWTAuthValidatorMiddleware(key, unauthorizedErrorMessage string, symmetric, logErrorMessage bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			bearerToken := getBearerToken(c.Request())
			if bearerToken == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, unauthorizedErrorMessage)
			}

			token, err := jwt.Parse(bearerToken, jwtKeyFunc(key, symmetric))
			if err != nil {
				if logErrorMessage {
					slog.Error("authentication failed", "module", "jwt-auth", "error", err)
				}
				return echo.NewHTTPError(http.StatusUnauthorized, unauthorizedErrorMessage)
			}
			if !token.Valid {
				return echo.NewHTTPError(http.StatusUnauthorized, unauthorizedErrorMessage)
			}

			return next(c)
		}
	}
}

// ServerKeyAuthValidatorMiddleware admits requests whose headerKey header
// equals serverKey or expiringServerKey (the latter supports key rotation),
// compared in constant time; others get 401 with unauthorizedErrorMessage.
// Both keys must be non-empty: an empty key matches a missing header.
func ServerKeyAuthValidatorMiddleware(headerKey, serverKey, expiringServerKey, unauthorizedErrorMessage string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			headerValue := c.Request().Header.Get(headerKey)
			if subtle.ConstantTimeCompare([]byte(headerValue), []byte(serverKey)) == 1 ||
				subtle.ConstantTimeCompare([]byte(headerValue), []byte(expiringServerKey)) == 1 {
				return next(c)
			}

			return echo.NewHTTPError(http.StatusUnauthorized, unauthorizedErrorMessage)
		}
	}
}

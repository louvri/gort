package echo

import (
	"crypto/subtle"
	"log/slog"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

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

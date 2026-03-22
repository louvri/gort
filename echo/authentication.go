package echo

import (
	"crypto/subtle"
	"net/http"

	"github.com/golang-jwt/jwt"
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

			var errorMessage string
			switch vErr := err.(type) {
			case nil:
				if !token.Valid {
					errorMessage = "invalid token"
				}
			case *jwt.ValidationError:
				switch vErr.Errors {
				case jwt.ValidationErrorExpired:
					errorMessage = "token expired"
				default:
					errorMessage = "token ValidationError error: " + vErr.Error()
				}
			default:
				errorMessage = "token parse error: " + err.Error()
			}

			if errorMessage != "" {
				if logErrorMessage {
					logAuthError(err.Error())
				}
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

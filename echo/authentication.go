package echo

import (
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/louvri/gort/common"
	"net/http"
)

func JWTAuthValidatorMiddleware(key, unauthorizedErrorMessage string, symmetric bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			bearerToken := common.GetBearerToken(c.Request())
			if bearerToken == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, unauthorizedErrorMessage)
			}

			token, err := jwt.Parse(bearerToken, func(token *jwt.Token) (interface{}, error) {
				if symmetric {
					return []byte(key), nil
				} else {
					verifyKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(key))
					if err != nil {
						fmt.Printf("{\"module\":\"jwt-auth\", \"error\":\"%s\"}\n", err.Error())
						return nil, err
					}

					return verifyKey, nil
				}
			})

			var errorMessage string
			switch err.(type) {
			case nil: // no error
				if !token.Valid { // but may still be invalid
					errorMessage = "invalid token"
				}
			case *jwt.ValidationError: // something was wrong during the validation
				vErr := err.(*jwt.ValidationError)

				switch vErr.Errors {
				case jwt.ValidationErrorExpired:
					errorMessage = "token expired"
				default:
					errorMessage = "token ValidationError error: " + vErr.Error()
				}
			default: // something else went wrong
				errorMessage = "token parse error: " + err.Error()
			}

			if errorMessage != "" {
				return echo.NewHTTPError(http.StatusUnauthorized, unauthorizedErrorMessage)
			}

			return next(c)
		}
	}
}

func ServerKeyAuthValidatorMiddleware(headerKey, serverKey, expiringServerKey, unauthorizedErrorMessage string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			header := c.Request().Header[headerKey]
			if len(header) > 0 {
				if header[0] == serverKey || header[0] == expiringServerKey {
					return next(c)
				}
			}

			return echo.NewHTTPError(http.StatusUnauthorized, unauthorizedErrorMessage)
		}
	}
}

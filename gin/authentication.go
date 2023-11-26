package gin

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/louvri/gort/common"
	"net/http"
)

func JWTAuthValidatorMiddleware(key, unauthorizedErrorMessage string, symmetric, logErrorMessage bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		bearerToken := common.GetBearerToken(c.Request)
		if bearerToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": unauthorizedErrorMessage})
			c.Abort()
			return
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
			var vErr *jwt.ValidationError
			errors.As(err, &vErr)

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
			if logErrorMessage {
				fmt.Printf("{\"module\":\"jwt-auth\", \"error\":\"%s\"}\n", err.Error())
			}
			c.JSON(http.StatusUnauthorized, gin.H{"message": unauthorizedErrorMessage})
			c.Abort()
			return
		}

		c.Next()
		return
	}
}

func ServerKeyAuthValidatorMiddleware(headerKey, serverKey, expiringServerKey, unauthorizedErrorMessage string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.Request.Header[headerKey]
		if len(header) > 0 {
			if header[0] == serverKey || header[0] == expiringServerKey {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusUnauthorized, gin.H{"message": unauthorizedErrorMessage})
		c.Abort()
		return
	}
}

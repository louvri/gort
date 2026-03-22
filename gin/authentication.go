package gin

import (
	"crypto/subtle"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func JWTAuthValidatorMiddleware(key, unauthorizedErrorMessage string, symmetric, logErrorMessage bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		bearerToken := getBearerToken(c.Request)
		if bearerToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": unauthorizedErrorMessage})
			c.Abort()
			return
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
			c.JSON(http.StatusUnauthorized, gin.H{"message": unauthorizedErrorMessage})
			c.Abort()
			return
		}

		c.Next()
	}
}

func ServerKeyAuthValidatorMiddleware(headerKey, serverKey, expiringServerKey, unauthorizedErrorMessage string) gin.HandlerFunc {
	return func(c *gin.Context) {
		headerValue := c.Request.Header.Get(headerKey)
		if subtle.ConstantTimeCompare([]byte(headerValue), []byte(serverKey)) == 1 ||
			subtle.ConstantTimeCompare([]byte(headerValue), []byte(expiringServerKey)) == 1 {
			c.Next()
			return
		}

		c.JSON(http.StatusUnauthorized, gin.H{"message": unauthorizedErrorMessage})
		c.Abort()
	}
}

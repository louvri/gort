package gin

import (
	"crypto/subtle"
	"log/slog"
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
		if err != nil {
			if logErrorMessage {
				slog.Error("authentication failed", "module", "jwt-auth", "error", err)
			}
			c.JSON(http.StatusUnauthorized, gin.H{"message": unauthorizedErrorMessage})
			c.Abort()
			return
		}
		if !token.Valid {
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

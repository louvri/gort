// Package gin provides authentication and maintenance-mode middleware for
// the Gin web framework.
package gin

import (
	"crypto/subtle"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTAuthValidatorMiddleware rejects requests without a valid "Bearer" JWT
// with 401 and unauthorizedErrorMessage. The signing algorithm is pinned to
// the key type: symmetric expects HMAC with key as the shared secret;
// otherwise it expects RSA with key as a PEM-encoded public key.
// logErrorMessage logs parse and verification failures.
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

// ServerKeyAuthValidatorMiddleware admits requests whose headerKey header
// equals serverKey or expiringServerKey (the latter supports key rotation),
// compared in constant time; others get 401 with unauthorizedErrorMessage.
// Both keys must be non-empty: an empty key matches a missing header.
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

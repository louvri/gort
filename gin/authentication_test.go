package gin

import (
	"github.com/gin-gonic/gin"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServerKeyAuthValidatorMiddleware(t *testing.T) {
	e := gin.Default()
	e.Use(ServerKeyAuthValidatorMiddleware("X-Server-Token", "jnX771xQ4tlLwG9GxlkheY6hd", "uctiiAkDQNe7eB7SEU1z5Ot4T", "Invalid token/session"))
	e.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello World")
	})

	emptyReq := httptest.NewRequest(echo.GET, "/", nil)
	emptyRec := httptest.NewRecorder()
	e.ServeHTTP(emptyRec, emptyReq)

	assert.Equal(t, http.StatusUnauthorized, emptyRec.Code)
	assert.Equal(t, "{\"message\":\"Invalid token/session\"}", emptyRec.Body.String())

	expiringReq := httptest.NewRequest(echo.GET, "/", nil)
	expiringReq.Header.Set("X-Server-Token", "uctiiAkDQNe7eB7SEU1z5Ot4T")
	expiringRec := httptest.NewRecorder()
	e.ServeHTTP(expiringRec, expiringReq)

	assert.Equal(t, http.StatusOK, expiringRec.Code)
	assert.Equal(t, "Hello World", expiringRec.Body.String())

	validReq := httptest.NewRequest(echo.GET, "/", nil)
	validReq.Header.Set("X-Server-Token", "jnX771xQ4tlLwG9GxlkheY6hd")
	validRec := httptest.NewRecorder()
	e.ServeHTTP(validRec, validReq)

	assert.Equal(t, http.StatusOK, validRec.Code)
	assert.Equal(t, "Hello World", validRec.Body.String())
}

func TestJWTAuthValidatorMiddleware(t *testing.T) {
	e := gin.Default()
	e.Use(JWTAuthValidatorMiddleware("testing", "Invalid token/session", true))
	e.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello World")
	})

	emptyReq := httptest.NewRequest(echo.GET, "/", nil)
	emptyRec := httptest.NewRecorder()
	e.ServeHTTP(emptyRec, emptyReq)

	assert.Equal(t, http.StatusUnauthorized, emptyRec.Code)
	assert.Equal(t, "{\"message\":\"Invalid token/session\"}", emptyRec.Body.String())

	invalidReq := httptest.NewRequest(echo.GET, "/", nil)
	invalidReq.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.J5S9dkJXUz8iUu6epfgTBFtNCHXFCm2VUFQWKDd8JHI")
	invalidRec := httptest.NewRecorder()
	e.ServeHTTP(invalidRec, invalidReq)

	assert.Equal(t, http.StatusUnauthorized, invalidRec.Code)
	assert.Equal(t, "{\"message\":\"Invalid token/session\"}", invalidRec.Body.String())

	validReq := httptest.NewRequest(echo.GET, "/", nil)
	validReq.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.UJFGOSLW3ar5q9qUk8IOFrOYUsdL8pd9je3yV2Kp-9g")
	validRec := httptest.NewRecorder()
	e.ServeHTTP(validRec, validReq)

	assert.Equal(t, http.StatusOK, validRec.Code)
	assert.Equal(t, "Hello World", validRec.Body.String())
}

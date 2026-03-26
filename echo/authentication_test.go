package echo

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func signToken(claims jwt.MapClaims, key string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(key))
	return signed
}

func setupEchoJWT(key, errMsg string, symmetric, logError bool) *echo.Echo {
	e := echo.New()
	e.Use(JWTAuthValidatorMiddleware(key, errMsg, symmetric, logError))
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello World")
	})
	return e
}

func TestJWTAuthValidatorMiddleware(t *testing.T) {
	key := "testing"
	errMsg := "Invalid token/session"

	t.Run("missing token", func(t *testing.T) {
		e := setupEchoJWT(key, errMsg, true, false)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), errMsg)
	})

	t.Run("invalid signature", func(t *testing.T) {
		e := setupEchoJWT(key, errMsg, true, false)
		token := signToken(jwt.MapClaims{"sub": "123"}, "wrong-key")
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), errMsg)
	})

	t.Run("expired token", func(t *testing.T) {
		e := setupEchoJWT(key, errMsg, true, false)
		token := signToken(jwt.MapClaims{
			"sub": "123",
			"exp": float64(time.Now().Add(-1 * time.Hour).Unix()),
		}, key)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("valid token", func(t *testing.T) {
		e := setupEchoJWT(key, errMsg, true, false)
		token := signToken(jwt.MapClaims{
			"sub":  "1234567890",
			"name": "John Doe",
			"iat":  float64(1516239022),
		}, key)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "Hello World", rec.Body.String())
	})

	t.Run("malformed token", func(t *testing.T) {
		e := setupEchoJWT(key, errMsg, true, false)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer not-a-jwt")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("Basic scheme rejected", func(t *testing.T) {
		e := setupEchoJWT(key, errMsg, true, false)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("log error message enabled", func(t *testing.T) {
		e := setupEchoJWT(key, errMsg, true, true)
		token := signToken(jwt.MapClaims{"sub": "123"}, "wrong-key")
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

func TestServerKeyAuthValidatorMiddleware(t *testing.T) {
	setupEcho := func() *echo.Echo {
		e := echo.New()
		e.Use(ServerKeyAuthValidatorMiddleware("X-Server-Token", "primary-key", "expiring-key", "Invalid token/session"))
		e.GET("/", func(c echo.Context) error {
			return c.String(http.StatusOK, "Hello World")
		})
		return e
	}

	t.Run("missing header", func(t *testing.T) {
		e := setupEcho()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "Invalid token/session")
	})

	t.Run("wrong key", func(t *testing.T) {
		e := setupEcho()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Server-Token", "wrong-key")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("primary key", func(t *testing.T) {
		e := setupEcho()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Server-Token", "primary-key")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "Hello World", rec.Body.String())
	})

	t.Run("expiring key", func(t *testing.T) {
		e := setupEcho()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Server-Token", "expiring-key")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "Hello World", rec.Body.String())
	})

	t.Run("empty key value", func(t *testing.T) {
		e := setupEcho()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Server-Token", "")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

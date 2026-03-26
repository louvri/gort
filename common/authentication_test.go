package common

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func signToken(claims jwt.MapClaims, key string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(key))
	return signed
}

func TestGetAuthorizationHeaderValue(t *testing.T) {
	t.Run("empty header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		assert.Equal(t, "", GetAuthorizationHeaderValue(req))
	})

	t.Run("with bearer token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer abc123")
		assert.Equal(t, "Bearer abc123", GetAuthorizationHeaderValue(req))
	})

	t.Run("with basic auth", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
		assert.Equal(t, "Basic dXNlcjpwYXNz", GetAuthorizationHeaderValue(req))
	})
}

func TestGetBearerToken(t *testing.T) {
	t.Run("empty header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		assert.Equal(t, "", GetBearerToken(req))
	})

	t.Run("no Bearer prefix", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "abc123")
		assert.Equal(t, "", GetBearerToken(req))
	})

	t.Run("Basic scheme rejected", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
		assert.Equal(t, "", GetBearerToken(req))
	})

	t.Run("case sensitive Bearer", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "bearer abc123")
		assert.Equal(t, "", GetBearerToken(req))
	})

	t.Run("valid Bearer token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer mytoken123")
		assert.Equal(t, "mytoken123", GetBearerToken(req))
	})

	t.Run("Bearer with empty token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer ")
		assert.Equal(t, "", GetBearerToken(req))
	})
}

func TestJWTKeyFunc(t *testing.T) {
	t.Run("symmetric returns key bytes", func(t *testing.T) {
		keyFunc := JWTKeyFunc("mysecret", true)
		result, err := keyFunc(&jwt.Token{Method: jwt.SigningMethodHS256, Header: map[string]any{"alg": "HS256"}})
		assert.Nil(t, err)
		assert.Equal(t, []byte("mysecret"), result)
	})

	t.Run("symmetric rejects RSA algorithm", func(t *testing.T) {
		keyFunc := JWTKeyFunc("mysecret", true)
		result, err := keyFunc(&jwt.Token{Method: jwt.SigningMethodRS256, Header: map[string]any{"alg": "RS256"}})
		assert.NotNil(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "unexpected signing method")
	})

	t.Run("asymmetric rejects HMAC algorithm", func(t *testing.T) {
		keyFunc := JWTKeyFunc("not-a-valid-pem", false)
		result, err := keyFunc(&jwt.Token{Method: jwt.SigningMethodHS256, Header: map[string]any{"alg": "HS256"}})
		assert.NotNil(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "unexpected signing method")
	})

	t.Run("asymmetric with invalid PEM", func(t *testing.T) {
		keyFunc := JWTKeyFunc("not-a-valid-pem", false)
		result, err := keyFunc(&jwt.Token{Method: jwt.SigningMethodRS256, Header: map[string]any{"alg": "RS256"}})
		assert.NotNil(t, err)
		assert.Nil(t, result)
	})
}

func TestGetMapClaimsFromJWT(t *testing.T) {
	key := "testing"

	t.Run("valid token", func(t *testing.T) {
		token := signToken(jwt.MapClaims{
			"sub":  "1234567890",
			"name": "John Doe",
			"iat":  float64(1516239022),
		}, key)
		claims, err := GetMapClaimsFromJWT(key, token, true)
		assert.Nil(t, err)
		assert.Equal(t, "1234567890", claims["sub"])
		assert.Equal(t, "John Doe", claims["name"])
	})

	t.Run("wrong signing key", func(t *testing.T) {
		token := signToken(jwt.MapClaims{"sub": "123"}, "different-key")
		claims, err := GetMapClaimsFromJWT(key, token, true)
		assert.NotNil(t, err)
		assert.Nil(t, claims)
	})

	t.Run("expired token", func(t *testing.T) {
		token := signToken(jwt.MapClaims{
			"sub": "123",
			"exp": float64(time.Now().Add(-1 * time.Hour).Unix()),
		}, key)
		claims, err := GetMapClaimsFromJWT(key, token, true)
		assert.NotNil(t, err)
		assert.Nil(t, claims)
	})

	t.Run("empty bearer token", func(t *testing.T) {
		claims, err := GetMapClaimsFromJWT(key, "", true)
		assert.NotNil(t, err)
		assert.Nil(t, claims)
	})

	t.Run("malformed token", func(t *testing.T) {
		claims, err := GetMapClaimsFromJWT(key, "not.a.jwt.token", true)
		assert.NotNil(t, err)
		assert.Nil(t, claims)
	})

	t.Run("garbage string", func(t *testing.T) {
		claims, err := GetMapClaimsFromJWT(key, "garbage", true)
		assert.NotNil(t, err)
		assert.Nil(t, claims)
	})
}

func TestGetMapClaimsFromJWTWithoutValidation(t *testing.T) {
	t.Run("valid token", func(t *testing.T) {
		token := signToken(jwt.MapClaims{
			"sub":  "1234567890",
			"name": "John Doe",
		}, "anykey")
		claims := GetMapClaimsFromJWTWithoutValidation(token)
		assert.Equal(t, "1234567890", claims["sub"])
		assert.Equal(t, "John Doe", claims["name"])
	})

	t.Run("token signed with different key still returns claims", func(t *testing.T) {
		token := signToken(jwt.MapClaims{
			"sub": "123",
		}, "unknown-key")
		claims := GetMapClaimsFromJWTWithoutValidation(token)
		assert.NotNil(t, claims)
		assert.Equal(t, "123", claims["sub"])
	})

	t.Run("expired token still returns claims", func(t *testing.T) {
		token := signToken(jwt.MapClaims{
			"sub": "123",
			"exp": float64(time.Now().Add(-1 * time.Hour).Unix()),
		}, "anykey")
		claims := GetMapClaimsFromJWTWithoutValidation(token)
		assert.NotNil(t, claims)
		assert.Equal(t, "123", claims["sub"])
	})

	t.Run("empty string", func(t *testing.T) {
		claims := GetMapClaimsFromJWTWithoutValidation("")
		assert.Nil(t, claims)
	})

	t.Run("garbage string", func(t *testing.T) {
		claims := GetMapClaimsFromJWTWithoutValidation("garbage")
		assert.Nil(t, claims)
	})

	t.Run("malformed token", func(t *testing.T) {
		claims := GetMapClaimsFromJWTWithoutValidation("a.b.c")
		assert.Nil(t, claims)
	})
}

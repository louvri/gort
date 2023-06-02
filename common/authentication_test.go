package common

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetBearerToken(t *testing.T) {
	emptyHeaderReq := httptest.NewRequest(http.MethodGet, "/", nil)
	assert.Equal(t, "", GetBearerToken(emptyHeaderReq))

	invalidReq := httptest.NewRequest(http.MethodGet, "/", nil)
	invalidReq.Header.Set("Authorization", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.UJFGOSLW3ar5q9qUk8IOFrOYUsdL8pd9je3yV2Kp-9g")
	assert.Equal(t, "", GetBearerToken(invalidReq))

	validReq := httptest.NewRequest(http.MethodGet, "/", nil)
	validReq.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.UJFGOSLW3ar5q9qUk8IOFrOYUsdL8pd9je3yV2Kp-9g")
	assert.Equal(t, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.UJFGOSLW3ar5q9qUk8IOFrOYUsdL8pd9je3yV2Kp-9g", GetBearerToken(validReq))
}

func TestGetMapClaimsFromJWT(t *testing.T) {
	key := "testing"
	validBearerToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.UJFGOSLW3ar5q9qUk8IOFrOYUsdL8pd9je3yV2Kp-9g"
	validClaims, err := GetMapClaimsFromJWT(key, validBearerToken, true)
	assert.Nil(t, err)
	assert.Equal(t, validClaims["sub"], "1234567890")

	invalidBearerToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.ZXJyb3I.Eol_UQDKYG_1LyZcqBYd2_TGj_sEpfxxP-WkUzPflk4"
	invalidClaims, err := GetMapClaimsFromJWT(key, invalidBearerToken, true)
	assert.NotNil(t, err)
	assert.Nil(t, invalidClaims)
}

func TestGetMapClaimsFromJWTWithoutValidation(t *testing.T) {
	validBearerToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.UJFGOSLW3ar5q9qUk8IOFrOYUsdL8pd9je3yV2Kp-9g"
	validClaims := GetMapClaimsFromJWTWithoutValidation(validBearerToken)
	assert.Equal(t, validClaims["sub"], "1234567890")

	invalidBearerToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.ZXJyb3I.Eol_UQDKYG_1LyZcqBYd2_TGj_sEpfxxP-WkUzPflk4"
	invalidClaims := GetMapClaimsFromJWTWithoutValidation(invalidBearerToken)
	assert.Nil(t, invalidClaims)
}

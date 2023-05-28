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

package echo

import (
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProbeMaintenanceMiddlewareInMaintenance(t *testing.T) {
	e := echo.New()
	e.Use(ProbeMaintenanceMiddleware([]string{"api/allowed"}, "Server Undergo Maintenance", http.StatusUnprocessableEntity, true))
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello World")
	})
	e.GET("/api/allowed", func(c echo.Context) error {
		return c.String(http.StatusOK, "Allowed Hello World")
	})

	emptyReq := httptest.NewRequest(echo.GET, "/", nil)
	emptyRec := httptest.NewRecorder()
	e.ServeHTTP(emptyRec, emptyReq)

	assert.Equal(t, http.StatusUnprocessableEntity, emptyRec.Code)
	assert.Equal(t, "{\"message\":\"Server Undergo Maintenance\"}\n", emptyRec.Body.String())

	allowedReq := httptest.NewRequest(echo.GET, "/api/allowed", nil)
	allowedRec := httptest.NewRecorder()
	e.ServeHTTP(allowedRec, allowedReq)

	assert.Equal(t, http.StatusOK, allowedRec.Code)
	assert.Equal(t, "Allowed Hello World", allowedRec.Body.String())
}

func TestProbeMaintenanceMiddlewareNotInMaintenance(t *testing.T) {
	e := echo.New()
	e.Use(ProbeMaintenanceMiddleware([]string{}, "Server Undergo Maintenance", http.StatusUnprocessableEntity, false))
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello World")
	})

	emptyReq := httptest.NewRequest(echo.GET, "/", nil)
	emptyRec := httptest.NewRecorder()
	e.ServeHTTP(emptyRec, emptyReq)

	assert.Equal(t, http.StatusOK, emptyRec.Code)
	assert.Equal(t, "Hello World", emptyRec.Body.String())
}

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
	e.Use(ProbeMaintenanceMiddleware("Server Undergo Maintenance", http.StatusUnprocessableEntity, true))
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello World")
	})

	emptyReq := httptest.NewRequest(echo.GET, "/", nil)
	emptyRec := httptest.NewRecorder()
	e.ServeHTTP(emptyRec, emptyReq)

	assert.Equal(t, http.StatusUnprocessableEntity, emptyRec.Code)
	assert.Equal(t, "{\"message\":\"Server Undergo Maintenance\"}\n", emptyRec.Body.String())
}

func TestProbeMaintenanceMiddlewareNotInMaintenance(t *testing.T) {
	e := echo.New()
	e.Use(ProbeMaintenanceMiddleware("Server Undergo Maintenance", http.StatusUnprocessableEntity, false))
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello World")
	})

	emptyReq := httptest.NewRequest(echo.GET, "/", nil)
	emptyRec := httptest.NewRecorder()
	e.ServeHTTP(emptyRec, emptyReq)

	assert.Equal(t, http.StatusOK, emptyRec.Code)
	assert.Equal(t, "Hello World", emptyRec.Body.String())
}

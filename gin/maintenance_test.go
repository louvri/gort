package gin

import (
	"github.com/gin-gonic/gin"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProbeMaintenanceMiddlewareInMaintenance(t *testing.T) {
	e := gin.Default()
	e.Use(ProbeMaintenanceMiddleware("Server Undergo Maintenance", http.StatusUnprocessableEntity, true))
	e.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello World")
	})

	emptyReq := httptest.NewRequest(echo.GET, "/", nil)
	emptyRec := httptest.NewRecorder()
	e.ServeHTTP(emptyRec, emptyReq)

	assert.Equal(t, http.StatusUnprocessableEntity, emptyRec.Code)
	assert.Equal(t, "{\"message\":\"Server Undergo Maintenance\"}", emptyRec.Body.String())
}

func TestProbeMaintenanceMiddlewareNotInMaintenance(t *testing.T) {
	e := gin.Default()
	e.Use(ProbeMaintenanceMiddleware("Server Undergo Maintenance", http.StatusUnprocessableEntity, false))
	e.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello World")
	})

	emptyReq := httptest.NewRequest(echo.GET, "/", nil)
	emptyRec := httptest.NewRecorder()
	e.ServeHTTP(emptyRec, emptyReq)

	assert.Equal(t, http.StatusOK, emptyRec.Code)
	assert.Equal(t, "Hello World", emptyRec.Body.String())
}

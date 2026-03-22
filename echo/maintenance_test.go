package echo

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func setupEchoMaintenance(skippedPaths []string, errMsg string, statusCode int, enabled bool) *echo.Echo {
	e := echo.New()
	e.Use(ProbeMaintenanceMiddleware(skippedPaths, errMsg, statusCode, enabled))
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello World")
	})
	e.GET("/api/allowed", func(c echo.Context) error {
		return c.String(http.StatusOK, "Allowed")
	})
	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})
	return e
}

func TestProbeMaintenanceMiddleware(t *testing.T) {
	t.Run("maintenance enabled blocks requests", func(t *testing.T) {
		e := setupEchoMaintenance(nil, "Under Maintenance", http.StatusServiceUnavailable, true)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
		assert.Contains(t, rec.Body.String(), "Under Maintenance")
	})

	t.Run("maintenance enabled allows skipped path", func(t *testing.T) {
		e := setupEchoMaintenance([]string{"/api/allowed"}, "Under Maintenance", http.StatusServiceUnavailable, true)
		req := httptest.NewRequest(http.MethodGet, "/api/allowed", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "Allowed", rec.Body.String())
	})

	t.Run("maintenance enabled blocks non-skipped path", func(t *testing.T) {
		e := setupEchoMaintenance([]string{"/api/allowed"}, "Under Maintenance", http.StatusServiceUnavailable, true)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	})

	t.Run("maintenance disabled passes through", func(t *testing.T) {
		e := setupEchoMaintenance(nil, "Under Maintenance", http.StatusServiceUnavailable, false)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "Hello World", rec.Body.String())
	})

	t.Run("multiple skipped paths", func(t *testing.T) {
		e := setupEchoMaintenance([]string{"/api/allowed", "/health"}, "Under Maintenance", http.StatusServiceUnavailable, true)

		req1 := httptest.NewRequest(http.MethodGet, "/api/allowed", nil)
		rec1 := httptest.NewRecorder()
		e.ServeHTTP(rec1, req1)
		assert.Equal(t, http.StatusOK, rec1.Code)

		req2 := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec2 := httptest.NewRecorder()
		e.ServeHTTP(rec2, req2)
		assert.Equal(t, http.StatusOK, rec2.Code)

		req3 := httptest.NewRequest(http.MethodGet, "/", nil)
		rec3 := httptest.NewRecorder()
		e.ServeHTTP(rec3, req3)
		assert.Equal(t, http.StatusServiceUnavailable, rec3.Code)
	})

	t.Run("custom status code", func(t *testing.T) {
		e := setupEchoMaintenance(nil, "Try Later", http.StatusUnprocessableEntity, true)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	})

	t.Run("empty skipped paths blocks all in maintenance", func(t *testing.T) {
		e := setupEchoMaintenance([]string{}, "Under Maintenance", http.StatusServiceUnavailable, true)

		req1 := httptest.NewRequest(http.MethodGet, "/", nil)
		rec1 := httptest.NewRecorder()
		e.ServeHTTP(rec1, req1)
		assert.Equal(t, http.StatusServiceUnavailable, rec1.Code)

		req2 := httptest.NewRequest(http.MethodGet, "/api/allowed", nil)
		rec2 := httptest.NewRecorder()
		e.ServeHTTP(rec2, req2)
		assert.Equal(t, http.StatusServiceUnavailable, rec2.Code)
	})
}

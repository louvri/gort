package gin

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupGinMaintenance(skippedPaths []string, errMsg string, statusCode int, enabled bool) *gin.Engine {
	e := gin.New()
	e.Use(ProbeMaintenanceMiddleware(skippedPaths, errMsg, statusCode, enabled))
	e.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello World")
	})
	e.GET("/api/allowed", func(c *gin.Context) {
		c.String(http.StatusOK, "Allowed")
	})
	e.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})
	return e
}

func TestProbeMaintenanceMiddleware(t *testing.T) {
	t.Run("maintenance enabled blocks requests", func(t *testing.T) {
		e := setupGinMaintenance(nil, "Under Maintenance", http.StatusServiceUnavailable, true)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
		assert.Contains(t, rec.Body.String(), "Under Maintenance")
	})

	t.Run("maintenance enabled allows skipped path", func(t *testing.T) {
		e := setupGinMaintenance([]string{"/api/allowed"}, "Under Maintenance", http.StatusServiceUnavailable, true)
		req := httptest.NewRequest(http.MethodGet, "/api/allowed", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "Allowed", rec.Body.String())
	})

	t.Run("maintenance enabled blocks non-skipped path", func(t *testing.T) {
		e := setupGinMaintenance([]string{"/api/allowed"}, "Under Maintenance", http.StatusServiceUnavailable, true)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	})

	t.Run("maintenance disabled passes through", func(t *testing.T) {
		e := setupGinMaintenance(nil, "Under Maintenance", http.StatusServiceUnavailable, false)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "Hello World", rec.Body.String())
	})

	t.Run("multiple skipped paths", func(t *testing.T) {
		e := setupGinMaintenance([]string{"/api/allowed", "/health"}, "Under Maintenance", http.StatusServiceUnavailable, true)

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
		e := setupGinMaintenance(nil, "Try Later", http.StatusUnprocessableEntity, true)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	})

	t.Run("empty skipped paths blocks all in maintenance", func(t *testing.T) {
		e := setupGinMaintenance([]string{}, "Under Maintenance", http.StatusServiceUnavailable, true)

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

func setupGinScheduledMaintenance(skippedPaths []string, errMsg string, statusCode int, from, to time.Time) *gin.Engine {
	e := gin.New()
	e.Use(ScheduledMaintenanceMiddleware(skippedPaths, errMsg, statusCode, from, to))
	e.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello World")
	})
	e.GET("/api/allowed", func(c *gin.Context) {
		c.String(http.StatusOK, "Allowed")
	})
	e.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})
	return e
}

func TestScheduledMaintenanceMiddleware(t *testing.T) {
	t.Run("before window passes through", func(t *testing.T) {
		from := time.Now().Add(1 * time.Hour)
		to := time.Now().Add(2 * time.Hour)
		e := setupGinScheduledMaintenance(nil, "Under Maintenance", http.StatusServiceUnavailable, from, to)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "Hello World", rec.Body.String())
	})

	t.Run("after window passes through", func(t *testing.T) {
		from := time.Now().Add(-2 * time.Hour)
		to := time.Now().Add(-1 * time.Hour)
		e := setupGinScheduledMaintenance(nil, "Under Maintenance", http.StatusServiceUnavailable, from, to)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "Hello World", rec.Body.String())
	})

	t.Run("within window blocks requests", func(t *testing.T) {
		from := time.Now().Add(-1 * time.Hour)
		to := time.Now().Add(1 * time.Hour)
		e := setupGinScheduledMaintenance(nil, "Under Maintenance", http.StatusServiceUnavailable, from, to)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
		assert.Contains(t, rec.Body.String(), "Under Maintenance")
	})

	t.Run("within window allows skipped path", func(t *testing.T) {
		from := time.Now().Add(-1 * time.Hour)
		to := time.Now().Add(1 * time.Hour)
		e := setupGinScheduledMaintenance([]string{"/health"}, "Under Maintenance", http.StatusServiceUnavailable, from, to)

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "OK", rec.Body.String())

		req2 := httptest.NewRequest(http.MethodGet, "/", nil)
		rec2 := httptest.NewRecorder()
		e.ServeHTTP(rec2, req2)
		assert.Equal(t, http.StatusServiceUnavailable, rec2.Code)
	})

	t.Run("custom status code", func(t *testing.T) {
		from := time.Now().Add(-1 * time.Hour)
		to := time.Now().Add(1 * time.Hour)
		e := setupGinScheduledMaintenance(nil, "Try Later", http.StatusUnprocessableEntity, from, to)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	})

	t.Run("both zero times passes through", func(t *testing.T) {
		e := setupGinScheduledMaintenance(nil, "Under Maintenance", http.StatusServiceUnavailable, time.Time{}, time.Time{})
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("zero from with future to blocks", func(t *testing.T) {
		to := time.Now().Add(1 * time.Hour)
		e := setupGinScheduledMaintenance(nil, "Under Maintenance", http.StatusServiceUnavailable, time.Time{}, to)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	})

	t.Run("zero from with past to passes through", func(t *testing.T) {
		to := time.Now().Add(-1 * time.Hour)
		e := setupGinScheduledMaintenance(nil, "Under Maintenance", http.StatusServiceUnavailable, time.Time{}, to)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("past from with zero to blocks indefinitely", func(t *testing.T) {
		from := time.Now().Add(-1 * time.Hour)
		e := setupGinScheduledMaintenance(nil, "Under Maintenance", http.StatusServiceUnavailable, from, time.Time{})
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	})

	t.Run("future from with zero to passes through", func(t *testing.T) {
		from := time.Now().Add(1 * time.Hour)
		e := setupGinScheduledMaintenance(nil, "Under Maintenance", http.StatusServiceUnavailable, from, time.Time{})
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

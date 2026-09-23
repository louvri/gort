package gin

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ProbeMaintenanceMiddleware rejects every request with statusCode and
// errorMessage while enabled, except those whose URL path starts with one of
// skippedPaths.
func ProbeMaintenanceMiddleware(skippedPaths []string, errorMessage string, statusCode int, enabled bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !enabled {
			c.Next()
			return
		}
		for _, skippedPath := range skippedPaths {
			if strings.HasPrefix(c.Request.URL.Path, skippedPath) {
				c.Next()
				return
			}
		}
		c.JSON(statusCode, gin.H{"message": errorMessage})
		c.Abort()
	}
}

// ScheduledMaintenanceMiddleware behaves like ProbeMaintenanceMiddleware while
// the current time is within [from, to]. A zero from starts the window
// immediately, a zero to leaves it open-ended, and both zero disables it.
func ScheduledMaintenanceMiddleware(skippedPaths []string, errorMessage string, statusCode int, from, to time.Time) gin.HandlerFunc {
	return func(c *gin.Context) {
		now := time.Now()
		if (from.IsZero() && to.IsZero()) || now.Before(from) || (!to.IsZero() && now.After(to)) {
			c.Next()
			return
		}
		for _, skippedPath := range skippedPaths {
			if strings.HasPrefix(c.Request.URL.Path, skippedPath) {
				c.Next()
				return
			}
		}
		c.JSON(statusCode, gin.H{"message": errorMessage})
		c.Abort()
	}
}

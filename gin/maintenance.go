package gin

import (
	"github.com/gin-gonic/gin"
	"strings"
)

func ProbeMaintenanceMiddleware(errorMessage, skippedPath string, statusCode int, enabled bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if enabled && !strings.Contains(c.Request.URL.Path, skippedPath) {
			c.JSON(statusCode, gin.H{"message": errorMessage})
			c.Abort()
			return
		}
		c.Next()
	}
}

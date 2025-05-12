package gin

import (
	"github.com/gin-gonic/gin"
	"strings"
)

func ProbeMaintenanceMiddleware(skippedPaths []string, errorMessage string, statusCode int, enabled bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if enabled {
			for _, skippedPath := range skippedPaths {
				if strings.Contains(c.Request.URL.Path, skippedPath) {
					c.Next()
					return
				}
			}
			c.JSON(statusCode, gin.H{"message": errorMessage})
			c.Abort()
			return
		}
		c.Next()
	}
}

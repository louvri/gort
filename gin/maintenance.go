package gin

import (
	"github.com/gin-gonic/gin"
)

func ProbeMaintenanceMiddleware(errorMessage string, statusCode int, enabled bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if enabled {
			c.JSON(statusCode, gin.H{"message": errorMessage})
			c.Abort()
			return
		}
		c.Next()
	}
}

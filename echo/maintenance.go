package echo

import (
	"github.com/labstack/echo/v4"
	"strings"
)

func ProbeMaintenanceMiddleware(skippedPaths []string, errorMessage string, statusCode int, enabled bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if enabled {
				for _, skippedPath := range skippedPaths {
					if strings.Contains(c.Request().URL.Path, skippedPath) {
						return next(c)
					}
				}
				return echo.NewHTTPError(statusCode, map[string]string{"message": errorMessage})
			}

			return next(c)
		}
	}
}

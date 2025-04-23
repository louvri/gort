package echo

import (
	"github.com/labstack/echo/v4"
)

func ProbeMaintenanceMiddleware(errorMessage string, statusCode int, enabled bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if enabled {
				return echo.NewHTTPError(statusCode, map[string]string{"message": errorMessage})
			}

			return next(c)
		}
	}
}

package echo

import (
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

// ProbeMaintenanceMiddleware rejects every request with statusCode and
// errorMessage while enabled, except those whose URL path starts with one of
// skippedPaths.
func ProbeMaintenanceMiddleware(skippedPaths []string, errorMessage string, statusCode int, enabled bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !enabled {
				return next(c)
			}
			for _, skippedPath := range skippedPaths {
				if strings.HasPrefix(c.Request().URL.Path, skippedPath) {
					return next(c)
				}
			}
			return echo.NewHTTPError(statusCode, map[string]string{"message": errorMessage})
		}
	}
}

// ScheduledMaintenanceMiddleware behaves like ProbeMaintenanceMiddleware while
// the current time is within [from, to]. A zero from starts the window
// immediately, a zero to leaves it open-ended, and both zero disables it.
func ScheduledMaintenanceMiddleware(skippedPaths []string, errorMessage string, statusCode int, from, to time.Time) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			now := time.Now()
			if (from.IsZero() && to.IsZero()) || now.Before(from) || (!to.IsZero() && now.After(to)) {
				return next(c)
			}
			for _, skippedPath := range skippedPaths {
				if strings.HasPrefix(c.Request().URL.Path, skippedPath) {
					return next(c)
				}
			}
			return echo.NewHTTPError(statusCode, map[string]string{"message": errorMessage})
		}
	}
}

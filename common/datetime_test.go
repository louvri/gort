package common

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExtractTimeZoneTextFromHeader(t *testing.T) {
	t.Run("empty header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		assert.Equal(t, "", ExtractTimeZoneTextFromHeader(req))
	})

	t.Run("GMT offset", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Timezone", "GMT+07")
		assert.Equal(t, "GMT+07", ExtractTimeZoneTextFromHeader(req))
	})

	t.Run("UTC offset", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Timezone", "UTC+05")
		assert.Equal(t, "UTC+05", ExtractTimeZoneTextFromHeader(req))
	})

	t.Run("named timezone", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Timezone", "Asia/Jakarta")
		assert.Equal(t, "Asia/Jakarta", ExtractTimeZoneTextFromHeader(req))
	})
}

func TestExtractTimeZoneInSecondsFromHeader(t *testing.T) {
	t.Run("empty header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		assert.Equal(t, 0, ExtractTimeZoneInSecondsFromHeader(req))
	})

	t.Run("GMT negative hour", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Timezone", "GMT-07")
		assert.Equal(t, -25200, ExtractTimeZoneInSecondsFromHeader(req))
	})

	t.Run("GMT positive hour", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Timezone", "GMT+07")
		assert.Equal(t, 25200, ExtractTimeZoneInSecondsFromHeader(req))
	})

	t.Run("GMT long format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Timezone", "GMT+0700")
		assert.Equal(t, 25200, ExtractTimeZoneInSecondsFromHeader(req))
	})

	t.Run("GMT negative with minutes", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Timezone", "GMT-0130")
		assert.Equal(t, -5400, ExtractTimeZoneInSecondsFromHeader(req))
	})

	t.Run("GMT positive with minutes", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Timezone", "GMT+0730")
		assert.Equal(t, 27000, ExtractTimeZoneInSecondsFromHeader(req))
	})

	t.Run("UTC prefix", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Timezone", "UTC+05")
		assert.Equal(t, 18000, ExtractTimeZoneInSecondsFromHeader(req))
	})

	t.Run("UTC long format with minutes", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Timezone", "UTC+0545")
		assert.Equal(t, 20700, ExtractTimeZoneInSecondsFromHeader(req))
	})

	t.Run("invalid named timezone", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Timezone", "Asia/Jakarta")
		assert.Equal(t, 0, ExtractTimeZoneInSecondsFromHeader(req))
	})

	t.Run("GMT with no offset", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Timezone", "GMT")
		assert.Equal(t, 0, ExtractTimeZoneInSecondsFromHeader(req))
	})

	t.Run("single character after prefix", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Timezone", "GMT+")
		assert.Equal(t, 0, ExtractTimeZoneInSecondsFromHeader(req))
	})
}

func TestExtractTimeZoneLocationFromHeader(t *testing.T) {
	t.Run("empty header returns UTC", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		loc := ExtractTimeZoneLocationFromHeader(req)
		assert.Equal(t, time.UTC, loc)
	})

	t.Run("GMT positive offset", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Timezone", "GMT+07")
		loc := ExtractTimeZoneLocationFromHeader(req)
		assert.Equal(t, "GMT+07", loc.String())
		_, offset := time.Now().In(loc).Zone()
		assert.Equal(t, 25200, offset)
	})

	t.Run("GMT negative offset", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Timezone", "GMT-05")
		loc := ExtractTimeZoneLocationFromHeader(req)
		assert.Equal(t, "GMT-05", loc.String())
		_, offset := time.Now().In(loc).Zone()
		assert.Equal(t, -18000, offset)
	})

	t.Run("UTC prefix", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Timezone", "UTC+0530")
		loc := ExtractTimeZoneLocationFromHeader(req)
		assert.Equal(t, "UTC+0530", loc.String())
		_, offset := time.Now().In(loc).Zone()
		assert.Equal(t, 19800, offset)
	})

	t.Run("invalid timezone returns zero offset", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Timezone", "Asia/Jakarta")
		loc := ExtractTimeZoneLocationFromHeader(req)
		_, offset := time.Now().In(loc).Zone()
		assert.Equal(t, 0, offset)
	})
}

func TestParseTimeWithFallback(t *testing.T) {
	t.Run("primary format succeeds", func(t *testing.T) {
		result := ParseTimeWithFallback("2024-01-15 10:30:00", "2006-01-02 15:04:05", "2006-01-02T15:04:05Z", time.UTC)
		assert.Equal(t, 2024, result.Year())
		assert.Equal(t, time.January, result.Month())
		assert.Equal(t, 15, result.Day())
		assert.Equal(t, 10, result.Hour())
		assert.Equal(t, 30, result.Minute())
	})

	t.Run("falls back to backup format", func(t *testing.T) {
		result := ParseTimeWithFallback("2024-01-15T10:30:00Z", "2006-01-02 15:04:05", "2006-01-02T15:04:05Z", time.UTC)
		assert.Equal(t, 2024, result.Year())
		assert.Equal(t, 10, result.Hour())
	})

	t.Run("both formats fail returns zero time", func(t *testing.T) {
		result := ParseTimeWithFallback("not-a-date", "2006-01-02 15:04:05", "2006-01-02T15:04:05Z", time.UTC)
		assert.True(t, result.IsZero())
	})

	t.Run("with non-UTC location", func(t *testing.T) {
		loc := time.FixedZone("WIB", 7*3600)
		result := ParseTimeWithFallback("2024-01-15 10:30:00", "2006-01-02 15:04:05", "2006-01-02T15:04:05Z", loc)
		assert.Equal(t, "WIB", result.Location().String())
	})
}

func TestParseUTCTime(t *testing.T) {
	t.Run("space separated format", func(t *testing.T) {
		result := ParseUTCTime("2024-06-15 14:30:00")
		assert.Equal(t, 2024, result.Year())
		assert.Equal(t, time.June, result.Month())
		assert.Equal(t, 15, result.Day())
		assert.Equal(t, 14, result.Hour())
		assert.Equal(t, 30, result.Minute())
		assert.Equal(t, time.UTC, result.Location())
	})

	t.Run("ISO 8601 format", func(t *testing.T) {
		result := ParseUTCTime("2024-06-15T14:30:00Z")
		assert.Equal(t, 2024, result.Year())
		assert.Equal(t, 14, result.Hour())
		assert.Equal(t, time.UTC, result.Location())
	})

	t.Run("invalid format returns zero time", func(t *testing.T) {
		result := ParseUTCTime("invalid")
		assert.True(t, result.IsZero())
	})

	t.Run("empty string returns zero time", func(t *testing.T) {
		result := ParseUTCTime("")
		assert.True(t, result.IsZero())
	})
}

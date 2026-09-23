package common

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ParseTimeWithFallback parses sourceTime in location using format, retrying
// with backupFormat on failure. It returns the zero time if both fail.
func ParseTimeWithFallback(sourceTime, format, backupFormat string, location *time.Location) time.Time {
	result, err := time.ParseInLocation(format, sourceTime, location)
	if err != nil {
		result, _ = time.ParseInLocation(backupFormat, sourceTime, location)
	}
	return result
}

// ParseUTCTime parses sourceTime as UTC in either "2006-01-02 15:04:05" or
// "2006-01-02T15:04:05Z" form, returning the zero time if neither matches.
func ParseUTCTime(sourceTime string) time.Time {
	return ParseTimeWithFallback(sourceTime, "2006-01-02 15:04:05", "2006-01-02T15:04:05Z", time.UTC)
}

// ExtractTimeZoneTextFromHeader returns the raw Timezone header of r.
func ExtractTimeZoneTextFromHeader(r *http.Request) string {
	return r.Header.Get("Timezone")
}

// ExtractTimeZoneLocationFromHeader returns a fixed zone named after the
// Timezone header with the offset from ExtractTimeZoneInSecondsFromHeader, or
// time.UTC when the header is absent.
func ExtractTimeZoneLocationFromHeader(r *http.Request) *time.Location {
	name := ExtractTimeZoneTextFromHeader(r)
	offset := ExtractTimeZoneInSecondsFromHeader(r)
	if name == "" {
		return time.UTC
	}
	return time.FixedZone(name, offset)
}

// ExtractTimeZoneInSecondsFromHeader parses the Timezone header as a UTC
// offset in seconds. A "GMT" or "UTC" prefix is ignored, and the remainder is
// read as hours ("+7", "-05") or as HHMM ("+0530"). It returns 0 when the
// header is missing or unparseable.
func ExtractTimeZoneInSecondsFromHeader(r *http.Request) int {
	timezone := strings.ToUpper(ExtractTimeZoneTextFromHeader(r))
	timezone = strings.ReplaceAll(strings.ReplaceAll(timezone, "GMT", ""), "UTC", "")
	if len(timezone) < 2 {
		return 0
	}
	rawData, err := strconv.Atoi(timezone)
	if err != nil {
		return 0
	}
	if len(timezone) < 4 {
		return rawData * 3600
	}
	return ((rawData / 100) * 3600) + ((rawData % 100) * 60)
}

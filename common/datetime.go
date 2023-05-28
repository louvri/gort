package common

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

func ParseTimeWithFallback(sourceTime, format, backupFormat string, location *time.Location) (result time.Time) {
	result, err := time.ParseInLocation(format, sourceTime, location)
	if err != nil {
		result, _ = time.ParseInLocation(backupFormat, sourceTime, location)
	}
	return result
}

func ParseUTCTime(sourceTime string) (result time.Time) {
	return ParseTimeWithFallback(sourceTime, "2006-01-02 15:04:05", "2006-01-02T15:04:05Z", time.UTC)
}

func ExtractTimeZoneTextFromHeader(r *http.Request) string {
	header := r.Header["Timezone"]
	if len(header) < 1 {
		header = r.Header["timezone"]
		if len(header) < 1 {
			return ""
		}
	}

	return header[0]
}

func ExtractTimeZoneInSecondsFromHeader(r *http.Request) int {
	timezone := strings.ToUpper(ExtractTimeZoneTextFromHeader(r))
	timezone = strings.ReplaceAll(strings.ReplaceAll(timezone, "GMT", ""), "UTC", "")
	if len(timezone) < 2 {
		return 0
	} else if len(timezone) < 4 {
		rawData, err := strconv.Atoi(timezone)
		if err != nil {
			return 0
		}
		return rawData * 3600
	} else {
		rawData, err := strconv.Atoi(timezone)
		if err != nil {
			return 0
		}
		return ((rawData / 100) * 3600) + ((rawData % 100) * 60)
	}
}

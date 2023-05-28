package common

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestExtractTimeZoneInSecondsFromHeader(t *testing.T) {
	negativeHourOnlyReq := httptest.NewRequest(http.MethodGet, "/", nil)
	negativeHourOnlyReq.Header.Set("Timezone", "GMT-07")
	assert.Equal(t, -25200, ExtractTimeZoneInSecondsFromHeader(negativeHourOnlyReq))

	positiveHourOnlyReq := httptest.NewRequest(http.MethodGet, "/", nil)
	positiveHourOnlyReq.Header.Set("Timezone", "GMT+07")
	assert.Equal(t, 25200, ExtractTimeZoneInSecondsFromHeader(positiveHourOnlyReq))

	hourWithLongFormatReq := httptest.NewRequest(http.MethodGet, "/", nil)
	hourWithLongFormatReq.Header.Set("Timezone", "GMT+0700")
	assert.Equal(t, 25200, ExtractTimeZoneInSecondsFromHeader(hourWithLongFormatReq))

	negativeHourWithMinutesReq := httptest.NewRequest(http.MethodGet, "/", nil)
	negativeHourWithMinutesReq.Header.Set("Timezone", "GMT-0130")
	assert.Equal(t, -5400, ExtractTimeZoneInSecondsFromHeader(negativeHourWithMinutesReq))

	positiveHourWithMinutesReq := httptest.NewRequest(http.MethodGet, "/", nil)
	positiveHourWithMinutesReq.Header.Set("Timezone", "GMT+0730")
	assert.Equal(t, 27000, ExtractTimeZoneInSecondsFromHeader(positiveHourWithMinutesReq))

	invalidReq := httptest.NewRequest(http.MethodGet, "/", nil)
	invalidReq.Header.Set("Timezone", "Asia/Jakarta")
	assert.Equal(t, 0, ExtractTimeZoneInSecondsFromHeader(invalidReq))
}

func TestParseUTCTime(t *testing.T) {
	sourceTime := time.Now().UTC().Add(5 * time.Hour)
	expectedTimeText := sourceTime.Format("2006-01-02 15:04:05")
	parsedTime := ParseUTCTime(expectedTimeText)
	assert.NotNil(t, parsedTime)
	assert.Equal(t, expectedTimeText, parsedTime.Format("2006-01-02 15:04:05"))

	expectedTimeWithDifferentFormatText := sourceTime.Format("2006-01-02T15:04:05Z")
	parsedTimeWithDifferentFormat := ParseUTCTime(expectedTimeWithDifferentFormatText)
	assert.NotNil(t, parsedTimeWithDifferentFormat)
	assert.Equal(t, expectedTimeText, parsedTimeWithDifferentFormat.Format("2006-01-02 15:04:05"))
}

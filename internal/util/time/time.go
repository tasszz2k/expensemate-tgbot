package time

import (
	"time"
)

const (
	// APIFormat is the format for API timestamps
	APIFormat = "02/01/2006 15:04:05"
	// DateOnlyFormat is the format for date-only values
	DateOnlyFormat = "2/1/2006"
	// DateTimeFormat is the format for datetime values in sheets
	DateTimeFormat = "2/1/2006 15:04"
)

// LocalLocation is the default timezone (Asia/Ho_Chi_Minh)
var LocalLocation *time.Location

func init() {
	loc, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		// Fallback to UTC if timezone not available
		loc = time.UTC
	}
	LocalLocation = loc
}

// FormatAPI formats time using APIFormat
func FormatAPI(t time.Time) string {
	return t.Format(APIFormat)
}

// FormatDateOnly formats time using DateOnlyFormat
func FormatDateOnly(t time.Time) string {
	return t.Format(DateOnlyFormat)
}

// FormatDateTime formats time using DateTimeFormat
func FormatDateTime(t time.Time) string {
	return t.In(LocalLocation).Format(DateTimeFormat)
}

// ParseAPILocal parses a string using APIFormat in local timezone
func ParseAPILocal(input string) (time.Time, error) {
	return time.ParseInLocation(APIFormat, input, LocalLocation)
}

// ParseDateOnly parses a string using DateOnlyFormat in local timezone
func ParseDateOnly(input string) (time.Time, error) {
	return time.ParseInLocation(DateOnlyFormat, input, LocalLocation)
}

// GetCurrentDay returns today's date at midnight
func GetCurrentDay() time.Time {
	return time.Now().Truncate(24 * time.Hour)
}

// GetDayAfter returns the day after the given time
func GetDayAfter(day time.Time) time.Time {
	return day.Add(24 * time.Hour)
}

// Now returns current time in local timezone
func Now() time.Time {
	return time.Now().In(LocalLocation)
}

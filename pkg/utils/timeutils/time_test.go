package timeutils_test

import (
	"testing"
	"time"

	"github.com/smartystreets/assertions"

	"zlpsaas-api/v2/pkg/utils/timeutils"
)

func TestGetCurrentDay(t *testing.T) {
	t.Parallel()
	// Get the current day
	currentDay := timeutils.GetCurrentDay()

	// Verify that the time is truncated to the beginning of the day
	assertions.ShouldBeTrue(
		t,
		currentDay.Hour() == 0 && currentDay.Minute() == 0 && currentDay.Second() == 0 &&
			currentDay.Nanosecond() == 0,
	)
}

func TestGetADayAfter(t *testing.T) {
	t.Parallel()
	// Create a specific day
	startDay := time.Date(2022, time.January, 1, 12, 0, 0, 0, time.UTC)

	// Get a day after the specified day
	nextDay := timeutils.GetADayAfter(startDay)

	// Verify that the next day is exactly 24 hours after the start day
	assertions.ShouldBeTrue(t, nextDay.Sub(startDay) == 24*time.Hour)
}

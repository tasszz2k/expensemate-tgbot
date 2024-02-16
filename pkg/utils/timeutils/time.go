package timeutils

import "time"

const (
	APIFormat      = "02/01/2006 15:04:05"
	DateOnlyFormat = "2/1/2006"
)

var LocLocal *time.Location

func init() {
	l, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		panic(err)
	}
	LocLocal = l
}

func FormatAPI(t time.Time) string {
	return t.Format(APIFormat)
}

func FormatDateOnly(t time.Time) string {
	return t.Format(DateOnlyFormat)
}

func ParseAPILocal(input string) (time.Time, error) {
	return time.ParseInLocation(APIFormat, input, LocLocal)
}

func GetCurrentDay() time.Time {
	return time.Now().Truncate(24 * time.Hour)
}

func GetADayAfter(day time.Time) time.Time {
	return day.Add(24 * time.Hour)
}

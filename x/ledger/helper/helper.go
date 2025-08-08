package helper

import "time"

// StrPtr returns a pointer to the string s.
func StrPtr(s string) *string {
	return &s
}

func DaysSinceEpoch(date time.Time) int32 {
	return int32(date.Sub(time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)).Hours() / 24)
}

func EpochDaysToISO8601(days int32) string {
	t := time.Unix(int64(days)*86400, 0).UTC()
	return t.Format("2006-01-02")
}

func StrToDate(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}

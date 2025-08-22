package helper

import "time"

// DaysSinceEpoch calculates the number of days between the given date and the Unix epoch (January 1, 1970).
func DaysSinceEpoch(date time.Time) int32 {
	d := date.Sub(time.Unix(0, 0))
	return int32(d / (24 * time.Hour)) //nolint:gosec // 24 * time.Hour is 86.4e12, so this always fits.
}

// EpochDaysToYMD converts a number of days since the Unix epoch (January 1, 1970) to a date string (YYYY-MM-DD).
func EpochDaysToYMD(days int32) string {
	t := time.Unix(int64(days)*86400, 0).UTC()
	return t.Format("2006-01-02")
}

// ParseYMD parses a date string in the format "YYYY-MM-DD" and returns a time.Time object or an error if invalid.
func ParseYMD(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}

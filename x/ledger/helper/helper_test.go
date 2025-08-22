package helper

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/provenance-io/provenance/testutil/assertions"
)

func TestDaysSinceEpoch(t *testing.T) {
	// Not all machines have all the timezone info. So, to get a location other than UTC,
	// I'm going to parse a string with an offset into a time, then just get that location.
	// Doing it in a function mostly so that the err variable isn't scoped to the whole test.
	nonUTCLocation := func() *time.Location {
		nonUTCTimeStr := "2021-04-20T16:20:00+04:00"
		nonUTCTime, err := time.Parse(time.RFC3339, nonUTCTimeStr)
		require.NoError(t, err, "time.Parse(time.RFC3339, %q", nonUTCTimeStr)
		return nonUTCTime.Location()
	}()

	tests := []struct {
		name string
		date time.Time
		exp  int32
	}{
		{
			name: "before epoch",
			date: time.Date(1969, 12, 25, 0, 0, 0, 0, time.UTC),
			exp:  -7,
		},
		{
			name: "on epoch",
			date: time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
			exp:  0,
		},
		{
			name: "after epoch",
			date: time.Date(1970, 1, 18, 0, 0, 0, 0, time.UTC),
			exp:  17,
		},
		{
			name: "utc with 00:00",
			date: time.Date(2025, 6, 9, 0, 0, 0, 0, time.UTC),
			exp:  20248,
		},
		{
			name: "utc with 00:01",
			date: time.Date(2025, 6, 9, 0, 1, 0, 0, time.UTC),
			exp:  20248,
		},
		{
			name: "utc with 01:00",
			date: time.Date(2025, 6, 9, 1, 0, 0, 0, time.UTC),
			exp:  20248,
		},
		{
			name: "utc with 12:00",
			date: time.Date(2025, 6, 9, 12, 0, 0, 0, time.UTC),
			exp:  20248,
		},
		{
			name: "utc with 23:00",
			date: time.Date(2025, 6, 9, 23, 0, 0, 0, time.UTC),
			exp:  20248,
		},
		{
			name: "utc with 23:59",
			date: time.Date(2025, 6, 9, 23, 59, 0, 0, time.UTC),
			exp:  20248,
		},
		{
			name: "non-utc with 00:00",
			date: time.Date(2025, 6, 9, 0, 0, 0, 0, nonUTCLocation),
			exp:  20247,
		},
		{
			name: "non-utc with 00:01",
			date: time.Date(2025, 6, 9, 0, 1, 0, 0, nonUTCLocation),
			exp:  20247,
		},
		{
			name: "non-utc with 01:00",
			date: time.Date(2025, 6, 9, 1, 0, 0, 0, nonUTCLocation),
			exp:  20247,
		},
		{
			name: "non-utc with 03:00",
			date: time.Date(2025, 6, 9, 3, 0, 0, 0, nonUTCLocation),
			exp:  20247,
		},
		{
			name: "non-utc with 03:59",
			date: time.Date(2025, 6, 9, 3, 59, 0, 0, nonUTCLocation),
			exp:  20247,
		},
		{
			name: "non-utc with 03:59:59",
			date: time.Date(2025, 6, 9, 3, 59, 0, 0, nonUTCLocation),
			exp:  20247,
		},
		{
			name: "non-utc with 04:00",
			date: time.Date(2025, 6, 9, 4, 0, 0, 0, nonUTCLocation),
			exp:  20248,
		},
		{
			name: "non-utc with 04:00:01",
			date: time.Date(2025, 6, 9, 4, 0, 1, 0, nonUTCLocation),
			exp:  20248,
		},
		{
			name: "non-utc with 12:00",
			date: time.Date(2025, 6, 9, 12, 0, 0, 0, nonUTCLocation),
			exp:  20248,
		},
		{
			name: "non-utc with 23:00",
			date: time.Date(2025, 6, 9, 23, 0, 0, 0, nonUTCLocation),
			exp:  20248,
		},
		{
			name: "non-utc with 23:59",
			date: time.Date(2025, 6, 9, 23, 59, 0, 0, nonUTCLocation),
			exp:  20248,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act int32
			testFunc := func() {
				act = DaysSinceEpoch(tc.date)
			}
			require.NotPanics(t, testFunc, "DaysSinceEpoch(%s)", tc.date)
			assert.Equal(t, int(tc.exp), int(act), "DaysSinceEpoch(%s) result", tc.date)
		})
	}
}

func TestEpochDaysToYMD(t *testing.T) {
	tests := []struct {
		days int32
		exp  string
	}{
		{days: -7, exp: "1969-12-25"},
		{days: 0, exp: "1970-01-01"},
		{days: 17, exp: "1970-01-18"},
		{days: 18737, exp: "2021-04-20"},
		{days: 20247, exp: "2025-06-08"},
		{days: 20248, exp: "2025-06-09"},
	}

	for _, tc := range tests {
		t.Run(strconv.Itoa(int(tc.days)), func(t *testing.T) {
			var act string
			testFunc := func() {
				act = EpochDaysToYMD(tc.days)
			}
			require.NotPanics(t, testFunc, "EpochDaysToYMD(%d)", tc.days)
			assert.Equal(t, tc.exp, act, "EpochDaysToYMD(%d) result", tc.days)
		})
	}
}

func TestParseYMD(t *testing.T) {
	tests := []struct {
		name    string
		dateStr string
		exp     time.Time
		expErr  string
	}{
		{
			name:    "two digit year",
			dateStr: "23-04-14",
			expErr:  "parsing time \"23-04-14\" as \"2006-01-02\": cannot parse \"23-04-14\" as \"2006\"",
		},
		{
			name:    "slashes",
			dateStr: "2014/07/30",
			expErr:  "parsing time \"2014/07/30\" as \"2006-01-02\": cannot parse \"/07/30\" as \"-\"",
		},
		{
			name:    "mdy",
			dateStr: "02-04-2016",
			expErr:  "parsing time \"02-04-2016\" as \"2006-01-02\": cannot parse \"02-04-2016\" as \"2006\"",
		},
		{
			name:    "month name",
			dateStr: "2019-April-14",
			expErr:  "parsing time \"2019-April-14\" as \"2006-01-02\": cannot parse \"April-14\" as \"01\"",
		},
		{
			name:    "pre-epoch",
			dateStr: "1969-12-25",
			exp:     time.Date(1969, 12, 25, 0, 0, 0, 0, time.UTC),
		},
		{
			name:    "epoch",
			dateStr: "1970-01-01",
			exp:     time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:    "mainnet block 0",
			dateStr: "2021-04-20",
			exp:     time.Date(2021, 4, 20, 0, 0, 0, 0, time.UTC),
		},
		{
			name:    "post-epoch",
			dateStr: "2025-06-09",
			exp:     time.Date(2025, 6, 9, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act time.Time
			var err error
			testFunc := func() {
				act, err = ParseYMD(tc.dateStr)
			}
			require.NotPanics(t, testFunc, "ParseYMD(%q)", tc.dateStr)
			assertions.AssertErrorValue(t, err, tc.expErr, "ParseYMD(%q) error", tc.dateStr)
			assert.Equal(t, tc.exp.String(), act.String(), "ParseYMD(%q) result", tc.dateStr)
		})
	}
}

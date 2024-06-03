package rand

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/provenance-io/provenance/testutil/assertions"
)

func TestSelectEntries(t *testing.T) {
	entries := make([]string, 3)
	for i := range entries {
		entries[i] = fmt.Sprintf("entry_%02d", i)
	}

	// Seeds are chosen through trial and error for some tests so that the results are what I want for that test case.
	tests := []struct {
		name        string
		seed        int64 // Defaults to 1 if not defined.
		entries     []string
		count       int
		entriesType string
		expected    []string
		expErr      string
	}{
		{
			name:     "nil entries, count 0",
			entries:  nil,
			count:    0,
			expected: nil,
		},
		{
			name:        "nil entries, count 1",
			entries:     nil,
			count:       1,
			entriesType: "entries",
			expErr:      "cannot choose 1 entries because there are only 0",
		},
		{
			name:        "nil entries, count 2",
			entries:     nil,
			count:       2,
			entriesType: "bananas",
			expErr:      "cannot choose 2 bananas because there are only 0",
		},
		{
			name:        "nil entries, count 3",
			entries:     nil,
			count:       3,
			entriesType: "cars",
			expErr:      "cannot choose 3 cars because there are only 0",
		},
		{
			name:     "empty entries, count 0",
			entries:  []string{},
			count:    0,
			expected: nil,
		},
		{
			name:        "empty entries, count 1",
			entries:     []string{},
			count:       1,
			entriesType: "thingies",
			expErr:      "cannot choose 1 thingies because there are only 0",
		},
		{
			name:        "empty entries, count 2",
			entries:     []string{},
			count:       2,
			entriesType: "red pandas",
			expErr:      "cannot choose 2 red pandas because there are only 0",
		},
		{
			name:        "empty entries, count 3",
			entries:     []string{},
			count:       3,
			entriesType: "legos",
			expErr:      "cannot choose 3 legos because there are only 0",
		},
		{
			name:     "one entry, count 0",
			entries:  entries[:1],
			count:    0,
			expected: nil,
		},
		{
			name:     "one entry, count 1",
			entries:  entries[1:2],
			count:    1,
			expected: entries[1:2],
		},
		{
			name:        "one entry, count 2",
			entries:     entries[2:3],
			count:       2,
			entriesType: "mice",
			expErr:      "cannot choose 2 mice because there are only 1",
		},
		{
			name:        "one entry, count 3",
			entries:     entries[:1],
			count:       3,
			entriesType: "remotes",
			expErr:      "cannot choose 3 remotes because there are only 1",
		},
		{
			name:     "two entries, count 0",
			seed:     2,
			entries:  entries[:2],
			count:    0,
			expected: nil,
		},
		{
			name:     "two entries, count 1, seed 1",
			seed:     1,
			entries:  entries[:2],
			count:    1,
			expected: []string{entries[1]},
		},
		{
			name:     "two entries, count 1, seed 2",
			seed:     2,
			entries:  entries[:2],
			count:    1,
			expected: []string{entries[0]},
		},
		{
			name:     "two entries, count 2",
			seed:     2,
			entries:  entries[:2],
			count:    2,
			expected: []string{entries[1], entries[0]},
		},
		{
			name:        "two entries, count 3",
			entries:     entries[:2],
			count:       3,
			entriesType: "shoelaces",
			expErr:      "cannot choose 3 shoelaces because there are only 2",
		},
		{
			name:     "three entries, count 0",
			entries:  entries[:3],
			count:    0,
			expected: nil,
		},
		{
			name:     "three entries, count 1",
			seed:     2,
			entries:  entries[:3],
			count:    1,
			expected: []string{entries[1]},
		},
		{
			name:     "three entries, count 2",
			seed:     10,
			entries:  entries[:3],
			count:    2,
			expected: []string{entries[2], entries[0]},
		},
		{
			name:     "three entries, count 3, seed 10",
			seed:     10,
			entries:  entries[:3],
			count:    3,
			expected: []string{entries[2], entries[0], entries[1]},
		},
		{
			name:     "three entries, count 3, seed 20",
			seed:     20,
			entries:  entries[:3],
			count:    3,
			expected: []string{entries[1], entries[2], entries[0]},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.seed == 0 {
				tc.seed = 1
			}
			if len(tc.entriesType) == 0 {
				tc.entriesType = "strings"
			}
			r := rand.New(rand.NewSource(tc.seed))
			var actual []string
			var err error
			testFunc := func() {
				actual, err = SelectEntries(r, tc.entries, tc.count, tc.entriesType)
			}
			require.NotPanics(t, testFunc, "SelectRandomEntries")
			assertions.AssertErrorValue(t, err, tc.expErr, "SelectRandomEntries error")
			assert.Equal(t, tc.expected, actual, "SelectRandomEntries result")
		})
	}
}

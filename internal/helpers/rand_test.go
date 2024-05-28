package helpers

import (
	"fmt"
	"math/rand"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/provenance-io/provenance/testutil/assertions"
)

func TestRandIntBetween(t *testing.T) {
	tests := []struct {
		min      int
		max      int
		expPanic bool
	}{
		{min: 0, max: 0},
		{min: 0, max: 10},
		{min: 1, max: 1},
		{min: -1, max: -1},
		{min: 1, max: 0, expPanic: true},
		{min: 0, max: -1, expPanic: true},
		{min: 1, max: -1, expPanic: true},
		{min: 10, max: -20, expPanic: true},
		{min: 10, max: -11, expPanic: true},
		{min: 10, max: -10, expPanic: true},
		{min: 10, max: -9, expPanic: true},
		{min: 10, max: 9, expPanic: true},
		{min: 10, max: 10},
		{min: 10, max: 11},
		{min: 10, max: 20},
		{min: -10, max: -11, expPanic: true},
		{min: -10, max: -10},
		{min: -10, max: -9},
		{min: -10, max: 9},
		{min: -10, max: 10},
		{min: -10, max: 11},
		{min: -20, max: -1},
		{min: -20, max: 0},
		{min: -20, max: 1},
		{min: 1001, max: 1100},
		{min: -1778, max: -1670},
	}

	for _, tc := range tests {
		name := fmt.Sprintf("RandIntBetween(%d, %d)", tc.min, tc.max)
		if tc.expPanic {
			name += " panics"
		}

		t.Run(name, func(t *testing.T) {
			r := rand.New(rand.NewSource(1))
			// Check for panic for the first try.
			seen := make(map[int]bool)
			testFunc := func() {
				val := RandIntBetween(r, tc.min, tc.max)
				seen[val] = true
			}
			if tc.expPanic {
				require.PanicsWithValue(t, "invalid argument to Intn", testFunc)
				return
			}
			require.NotPanics(t, testFunc)

			count := tc.max - tc.min + 1
			expected := make([]int, 0, count)
			for i := tc.min; i <= tc.max; i++ {
				expected = append(expected, i)
			}

			// Run it a bunch of times, trying to get it to return all possible values.
			// I chose count*100 to essentially give each value 100 chances to be chosen, but be
			// low enough to still finish pretty quickly if one or more values never gets returned.
			for i := 0; i < count*100 && len(seen) < count; i++ {
				testFunc()
			}
			// Make sure both the min and max were returned at some point.
			assert.True(t, seen[tc.min], "minimum value %d in seen map", tc.min)
			assert.True(t, seen[tc.max], "maximum value %d in seen map", tc.max)

			seenVals := Keys(seen)
			slices.Sort(seenVals)
			// Make sure the smallest and largest are as expected.
			assert.Equal(t, tc.min, seenVals[0], "smallest number generated")
			assert.Equal(t, tc.max, seenVals[len(seenVals)-1], "largest number generated")
			// Make sure all values were generated. This check technically covers the previous ones,
			// but I've got them split out like this for friendlier test failure messages.
			assert.Equal(t, expected, seenVals, "values generated")
		})
	}
}

func TestSelectRandomEntries(t *testing.T) {
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
				actual, err = SelectRandomEntries(r, tc.entries, tc.count, tc.entriesType)
			}
			require.NotPanics(t, testFunc, "SelectRandomEntries")
			assertions.AssertErrorValue(t, err, tc.expErr, "SelectRandomEntries error")
			assert.Equal(t, tc.expected, actual, "SelectRandomEntries result")
		})
	}
}

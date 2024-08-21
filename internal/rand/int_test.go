package rand

import (
	"fmt"
	"maps"
	"math/rand"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntBetween(t *testing.T) {
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
				val := IntBetween(r, tc.min, tc.max)
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

			seenVals := slices.Collect(maps.Keys(seen))
			// Make sure the smallest and largest are as expected.
			assert.Equal(t, tc.min, seenVals[0], "smallest number generated")
			assert.Equal(t, tc.max, seenVals[len(seenVals)-1], "largest number generated")
			// Make sure all values were generated. This check technically covers the previous ones,
			// but I've got them split out like this for friendlier test failure messages.
			assert.Equal(t, expected, seenVals, "values generated")
		})
	}
}

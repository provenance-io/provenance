package provutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// stringSame is a string with an IsSameAs(stringSame) function.
type stringSame string

// IsSameAs returns true if this stringSame is the same as the provided one.
func (s stringSame) IsSameAs(c stringSame) bool {
	return string(s) == string(c)
}

// newStringSames converts a slice of strings to a slice of stringEqs.
// nil in => nil out. empty in => empty out.
func newStringSames(strs []string) []stringSame {
	if strs == nil {
		return nil
	}
	rv := make([]stringSame, len(strs), cap(strs))
	for i, str := range strs {
		rv[i] = stringSame(str)
	}
	return rv
}

// stringSameR is a string with an IsSameAs(stringSameC) function (note the other type there).
type stringSameR string

// stringSameC is a string that can be provided to the stringSameR IsSameAs function.
type stringSameC string

// IsSameAs returns true if this stringSameC is the same as the provided stringSameC.
func (s stringSameR) IsSameAs(c stringSameC) bool {
	return string(s) == string(c)
}

// newStringSameRs converts a slice of strings to a slice of stringEqRs.
// nil in => nil out. empty in => empty out.
func newStringSameRs(strs []string) []stringSameR {
	if strs == nil {
		return nil
	}
	rv := make([]stringSameR, len(strs), cap(strs))
	for i, str := range strs {
		rv[i] = stringSameR(str)
	}
	return rv
}

// newStringSameCs converts a slice of strings to a slice of stringEqCs.
// nil in => nil out. empty in => empty out.
func newStringSameCs(strs []string) []stringSameC {
	if strs == nil {
		return nil
	}
	rv := make([]stringSameC, len(strs), cap(strs))
	for i, str := range strs {
		rv[i] = stringSameC(str)
	}
	return rv
}

type testCaseFindMissing struct {
	name     string
	required []string
	toCheck  []string
	expected []string
}

func testCasesForFindMissing() []testCaseFindMissing {
	return []testCaseFindMissing{
		{
			name:     "nil required - nil toCheck - nil out",
			required: nil,
			toCheck:  nil,
			expected: nil,
		},
		{
			name:     "empty required - nil toCheck - nil out",
			required: []string{},
			toCheck:  nil,
			expected: nil,
		},
		{
			name:     "nil required - empty toCheck - nil out",
			required: nil,
			toCheck:  []string{},
			expected: nil,
		},
		{
			name:     "empty required - empty toCheck - nil out",
			required: []string{},
			toCheck:  []string{},
			expected: nil,
		},
		{
			name:     "nil required - 2 toCheck - nil out",
			required: nil,
			toCheck:  []string{"one", "two"},
			expected: nil,
		},
		{
			name:     "empty required - 2 toCheck - nil out",
			required: []string{},
			toCheck:  []string{"one", "two"},
			expected: nil,
		},
		{
			name:     "1 required - is only toCheck - nil out",
			required: []string{"one"},
			toCheck:  []string{"one"},
			expected: nil,
		},
		{
			name:     "1 required - is 1st of 2 toCheck - nil out",
			required: []string{"one"},
			toCheck:  []string{"one", "two"},
			expected: nil,
		},
		{
			name:     "1 required - is 2nd of 2 toCheck - nil out",
			required: []string{"one"},
			toCheck:  []string{"two", "one"},
			expected: nil,
		},
		{
			name:     "1 required -  nil toCheck - required out",
			required: []string{"one"},
			toCheck:  nil,
			expected: []string{"one"},
		},
		{
			name:     "1 required - empty toCheck - required out",
			required: []string{"one"},
			toCheck:  []string{},
			expected: []string{"one"},
		},
		{
			name:     "1 required - 1 other in toCheck - required out",
			required: []string{"one"},
			toCheck:  []string{"two"},
			expected: []string{"one"},
		},
		{
			name:     "1 required - 2 other in toCheck - required out",
			required: []string{"one"},
			toCheck:  []string{"two", "three"},
			expected: []string{"one"},
		},
		{
			name:     "2 required - both in toCheck - nil out",
			required: []string{"one", "two"},
			toCheck:  []string{"one", "two"},
			expected: nil,
		},
		{
			name:     "2 required - reversed in toCheck - nil out",
			required: []string{"one", "two"},
			toCheck:  []string{"two", "one"},
			expected: nil,
		},
		{
			name:     "2 required - only 1st in toCheck - 2nd out",
			required: []string{"one", "two"},
			toCheck:  []string{"one"},
			expected: []string{"two"},
		},
		{
			name:     "2 required - only 2nd in toCheck - 1st out",
			required: []string{"one", "two"},
			toCheck:  []string{"two"},
			expected: []string{"one"},
		},
		{
			name:     "2 required - 1st and other in toCheck - 2nd out",
			required: []string{"one", "two"},
			toCheck:  []string{"one", "other"},
			expected: []string{"two"},
		},
		{
			name:     "2 required - 2nd and other in toCheck - 1st out",
			required: []string{"one", "two"},
			toCheck:  []string{"two", "other"},
			expected: []string{"one"},
		},
		{
			name:     "2 required - nil toCheck - required out",
			required: []string{"one", "two"},
			toCheck:  nil,
			expected: []string{"one", "two"},
		},
		{
			name:     "2 required - empty toCheck - required out",
			required: []string{"one", "two"},
			toCheck:  []string{},
			expected: []string{"one", "two"},
		},
		{
			name:     "2 required - neither in 1 toCheck - required out",
			required: []string{"one", "two"},
			toCheck:  []string{"neither"},
			expected: []string{"one", "two"},
		},
		{
			name:     "2 required - neither in 3 toCheck - required out",
			required: []string{"one", "two"},
			toCheck:  []string{"neither", "nor", "nothing"},
			expected: []string{"one", "two"},
		},
		{
			name:     "2 required - 1st not in 3 toCheck 2nd at 0 - 1st out",
			required: []string{"one", "two"},
			toCheck:  []string{"two", "nor", "nothing"},
			expected: []string{"one"},
		},
		{
			name:     "2 required - 1st not in 3 toCheck 2nd at 1 - 1st out",
			required: []string{"one", "two"},
			toCheck:  []string{"neither", "two", "nothing"},
			expected: []string{"one"},
		},
		{
			name:     "2 required - 1s5 not in 3 toCheck 2nd at 2 - 1st out",
			required: []string{"one", "two"},
			toCheck:  []string{"neither", "nor", "two"},
			expected: []string{"one"},
		},
		{
			name:     "2 required - 2nd not in 3 toCheck 1st at 0 - 2nd out",
			required: []string{"one", "two"},
			toCheck:  []string{"one", "nor", "nothing"},
			expected: []string{"two"},
		},
		{
			name:     "2 required - 2nd not in 3 toCheck 1st at 1 - 2nd out",
			required: []string{"one", "two"},
			toCheck:  []string{"neither", "one", "nothing"},
			expected: []string{"two"},
		},
		{
			name:     "2 required - 2nd not in 3 toCheck 1st at 2 - 2nd out",
			required: []string{"one", "two"},
			toCheck:  []string{"neither", "nor", "one"},
			expected: []string{"two"},
		},

		{
			name:     "3 required - none in 5 toCheck - required out",
			required: []string{"one", "two", "three"},
			toCheck:  []string{"other1", "other2", "other3", "other4", "other5"},
			expected: []string{"one", "two", "three"},
		},
		{
			name:     "3 required - only 1st in 5 toCheck - 2nd 3rd out",
			required: []string{"one", "two", "three"},
			toCheck:  []string{"other1", "other2", "one", "other4", "other5"},
			expected: []string{"two", "three"},
		},
		{
			name:     "3 required - only 2nd in 5 toCheck - 1st 3rd out",
			required: []string{"one", "two", "three"},
			toCheck:  []string{"other1", "two", "other3", "other4", "other5"},
			expected: []string{"one", "three"},
		},
		{
			name:     "3 required - only 3rd in 5 toCheck - 1st 2nd out",
			required: []string{"one", "two", "three"},
			toCheck:  []string{"other1", "other2", "other3", "three", "other5"},
			expected: []string{"one", "two"},
		},
		{
			name:     "3 required - 1st 2nd in 5 toCheck - 3rd out",
			required: []string{"one", "two", "three"},
			toCheck:  []string{"other1", "two", "other3", "one", "other5"},
			expected: []string{"three"},
		},
		{
			name:     "3 required - 1st 3nd in 5 toCheck - 2nd out",
			required: []string{"one", "two", "three"},
			toCheck:  []string{"three", "other2", "other3", "other4", "one"},
			expected: []string{"two"},
		},
		{
			name:     "3 required - 2nd 3rd in 5 toCheck - 1st out",
			required: []string{"one", "two", "three"},
			toCheck:  []string{"other1", "other2", "two", "three", "other5"},
			expected: []string{"one"},
		},
		{
			name:     "3 required - all in 5 toCheck - nil out",
			required: []string{"one", "two", "three"},
			toCheck:  []string{"two", "other2", "one", "three", "other5"},
			expected: nil,
		},
		{
			name:     "3 required with dup - all in toCheck - nil out",
			required: []string{"one", "two", "one"},
			toCheck:  []string{"one", "two"},
			expected: nil,
		},
		{
			name:     "3 required with dup - dup not in toCheck - dups out",
			required: []string{"one", "two", "one"},
			toCheck:  []string{"two"},
			expected: []string{"one", "one"},
		},
		{
			name:     "3 required with dup - other not in toCheck - other out",
			required: []string{"one", "two", "one"},
			toCheck:  []string{"one"},
			expected: []string{"two"},
		},
		{
			name:     "3 required all dup - in toCheck - nil out",
			required: []string{"one", "one", "one"},
			toCheck:  []string{"one"},
			expected: nil,
		},
		{
			name:     "3 required all dup - not in toCheck - all 3 out",
			required: []string{"one", "one", "one"},
			toCheck:  []string{"two"},
			expected: []string{"one", "one", "one"},
		},
	}
}

func TestFindMissing(t *testing.T) {
	for _, tc := range testCasesForFindMissing() {
		t.Run(tc.name, func(t *testing.T) {
			actual := FindMissing(tc.required, tc.toCheck)
			assert.Equal(t, tc.expected, actual, "findMissing")
		})
	}
}

func TestFindMissingFunc(t *testing.T) {
	t.Run("equals equals", func(t *testing.T) {
		equals := func(r, c string) bool {
			return r == c
		}
		for _, tc := range testCasesForFindMissing() {
			t.Run(tc.name, func(t *testing.T) {
				actual := FindMissingFunc(tc.required, tc.toCheck, equals)
				assert.Equal(t, tc.expected, actual, "FindMissingFunc")
			})
		}
	})

	t.Run("is same as same types", func(t *testing.T) {
		equals := func(r, c stringSame) bool {
			return r.IsSameAs(c)
		}
		for _, tc := range testCasesForFindMissing() {
			t.Run(tc.name, func(t *testing.T) {
				required := newStringSames(tc.required)
				toCheck := newStringSames(tc.toCheck)
				expected := newStringSames(tc.expected)
				actual := FindMissingFunc(required, toCheck, equals)
				assert.Equal(t, expected, actual, "FindMissingFunc")
			})
		}
	})

	t.Run("is same as different types", func(t *testing.T) {
		equals := func(r stringSameR, c stringSameC) bool {
			return r.IsSameAs(c)
		}
		for _, tc := range testCasesForFindMissing() {
			t.Run(tc.name, func(t *testing.T) {
				required := newStringSameRs(tc.required)
				toCheck := newStringSameCs(tc.toCheck)
				expected := newStringSameRs(tc.expected)
				actual := FindMissingFunc(required, toCheck, equals)
				assert.Equal(t, expected, actual, "FindMissingFunc")
			})
		}
	})

	t.Run("string lengths", func(t *testing.T) {
		equals := func(r string, c int) bool {
			return len(r) == c
		}
		req := []string{"a", "bb", "ccc", "dddd", "eeeee"}
		checks := []struct {
			name     string
			toCheck  []int
			expected []string
		}{
			{name: "all there", toCheck: []int{1, 2, 3, 4, 5}, expected: nil},
			{name: "missing len 1", toCheck: []int{2, 3, 4, 5}, expected: []string{"a"}},
			{name: "missing len 2", toCheck: []int{1, 3, 4, 5}, expected: []string{"bb"}},
			{name: "missing len 3", toCheck: []int{1, 2, 4, 5}, expected: []string{"ccc"}},
			{name: "missing len 4", toCheck: []int{1, 2, 3, 5}, expected: []string{"dddd"}},
			{name: "missing len 5", toCheck: []int{1, 2, 3, 4}, expected: []string{"eeeee"}},
			{name: "none there", toCheck: []int{0, 6}, expected: req},
		}
		for _, tc := range checks {
			t.Run(tc.name, func(t *testing.T) {
				actual := FindMissingFunc(req, tc.toCheck, equals)
				assert.Equal(t, tc.expected, actual, "FindMissingFunc equals returns true if the string is the right length")
			})
		}
	})

	t.Run("div two", func(t *testing.T) {
		equals := func(r int, c int) bool {
			return r/2 == c
		}
		req := []int{1, 2, 3, 4, 5}
		checks := []struct {
			name     string
			toCheck  []int
			expected []int
		}{
			{name: "all there", toCheck: []int{0, 1, 2}, expected: nil},
			{name: "missing 0", toCheck: []int{1, 2}, expected: []int{1}},
			{name: "missing 1", toCheck: []int{0, 2}, expected: []int{2, 3}},
			{name: "missing 2", toCheck: []int{0, 1}, expected: []int{4, 5}},
			{name: "none there", toCheck: []int{-1, 3}, expected: req},
		}
		for _, tc := range checks {
			t.Run(tc.name, func(t *testing.T) {
				actual := FindMissingFunc(req, tc.toCheck, equals)
				assert.Equal(t, tc.expected, actual, "FindMissingFunc equals returns true if r/2 == c")
			})
		}
	})

	t.Run("all true", func(t *testing.T) {
		equals := func(r, c string) bool {
			return true
		}
		for _, tc := range testCasesForFindMissing() {
			t.Run(tc.name, func(t *testing.T) {
				var expected []string
				// required entries are only marked as found after being compared to something.
				// So if there's nothing in the toCheck list, all the required will be returned.
				// But if tc.required is an empty slice, we still expect to get nil back, so we don't
				// set expected = tc.required in that case.
				if len(tc.toCheck) == 0 && len(tc.required) > 0 {
					expected = tc.required
				}
				actual := FindMissingFunc(tc.required, tc.toCheck, equals)
				assert.Equal(t, expected, actual, "FindMissingFunc equals always returns true")
			})
		}
	})

	t.Run("all false", func(t *testing.T) {
		equals := func(r, c string) bool {
			return false
		}
		for _, tc := range testCasesForFindMissing() {
			t.Run(tc.name, func(t *testing.T) {
				// If tc.required is nil, or an empty slice, we expect nil, otherwise, we always expect tc.required back.
				var expected []string
				if len(tc.required) > 0 {
					expected = tc.required
				}
				actual := FindMissingFunc(tc.required, tc.toCheck, equals)
				assert.Equal(t, expected, actual, "FindMissingFunc equals always returns false")
			})
		}
	})
}

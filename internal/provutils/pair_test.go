package provutils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// pairTestCase is a struct that defines a test case to run involving the Pair type.
type pairTestCase[KA any, KB any] struct {
	// A is the A value to give the initial pair.
	A KA
	// B is the B value to give the initial pair.
	B KB
	// ExpStr is the expected value of the initial pair.String().
	ExpStr string

	// NewA is a pointer to an A value to provide to SetA. Leave nil to not call SetA.
	NewA *KA
	// NewB is a pointer to a B value to provide to SetB. Leave nil to not call SetB.
	NewB *KB
	// ExpNewStr is the expected value of pair.String() after calling SetA and/or SetB.
	ExpNewStr string
}

// runPairTest runs a subtest that creates a pair and makes sure it behaves as expected.
func runPairTest[KA any, KB any](t *testing.T, tc pairTestCase[KA, KB]) bool {
	t.Helper()
	name := tc.ExpStr
	if len(tc.ExpNewStr) > 0 {
		name += " to " + tc.ExpNewStr
	}
	rv := t.Run(name, func(t *testing.T) {
		var pair *Pair[KA, KB]
		testNewPair := func() {
			pair = NewPair(tc.A, tc.B)
		}
		require.NotPanics(t, testNewPair, "NewPair")
		require.NotNil(t, pair, "NewPair result")

		assertPair(t, pair, tc.A, tc.B, tc.ExpStr, "")

		if tc.NewA == nil && tc.NewB == nil {
			return // No desired changes, so no reason to do the rest.
		}

		expA, expB := tc.A, tc.B
		setAOK, setBOK := true, true
		if tc.NewA != nil {
			expA = *tc.NewA
			testSetA := func() {
				pair.SetA(expA)
			}
			setAOK = assert.NotPanics(t, testSetA, "SetA(...)")
		}
		if tc.NewB != nil {
			expB = *tc.NewB
			testSetB := func() {
				pair.SetB(expB)
			}
			setBOK = assert.NotPanics(t, testSetB, "SetB(...)")
		}

		if setAOK && setBOK {
			assertPair(t, pair, expA, expB, tc.ExpNewStr, " after being changed")
		}
	})
	return rv
}

// assertPair runs several assertions against the provided pair, ensuring it is and behaves as expected.
// Returns true if it passes all assertions, false if there's something wrong.
func assertPair[KA any, KB any](t *testing.T, pair *Pair[KA, KB], expA KA, expB KB, expStr string, msgSuffix string) {
	t.Helper()
	assert.Equal(t, expA, pair.A, "A field%s", msgSuffix)
	assert.Equal(t, expB, pair.B, "B field%s", msgSuffix)

	var actA KA
	testGetA := func() {
		actA = pair.GetA()
	}
	if assert.NotPanics(t, testGetA, "GetA()%s", msgSuffix) {
		assert.Equal(t, expA, actA, "GetA() result%s", msgSuffix)
	}

	var actB KB
	testGetB := func() {
		actB = pair.GetB()
	}
	if assert.NotPanics(t, testGetB, "GetB()%s", msgSuffix) {
		assert.Equal(t, expB, actB, "GetB() result%s", msgSuffix)
	}

	var actString string
	testString := func() {
		actString = pair.String()
	}
	if assert.NotPanics(t, testString, "String()%s", msgSuffix) {
		assert.Equal(t, expStr, actString, "String() result%s", msgSuffix)
	}

	var actAV KA
	var actBV KB
	testValues := func() {
		actAV, actBV = pair.Values()
	}
	if assert.NotPanics(t, testValues, "Values()%s", msgSuffix) {
		assert.Equal(t, expA, actAV, "A value returned from Values()%s", msgSuffix)
		assert.Equal(t, expB, actBV, "B value returned from Values()%s", msgSuffix)
	}
}

// ptrTo returns a pointer to the provided value. It's like & but this works on built-in data types.
func ptrTo[V any](v V) *V {
	return &v
}

// testType is just a simple dumb struct that I can use to test some pair stuff.
type testType struct {
	Id string
}

// newTT creates a new testType with the given id.
func newTT(id string) *testType {
	return &testType{Id: id}
}

// String implements the fmt.Stringer interface so that I know what %v looks like for a testType.
func (t *testType) String() string {
	if t == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%T=%s", t, t.Id)
}

func TestPair(t *testing.T) {
	tests := []pairTestCase[int, string]{
		{A: 0, B: "", ExpStr: "<0:>", NewA: ptrTo(1), NewB: ptrTo("one"), ExpNewStr: "<1:one>"},
		{A: 1, B: "two", ExpStr: "<1:two>", NewA: ptrTo(555), NewB: ptrTo("seven"), ExpNewStr: "<555:seven>"},
		{A: 33, B: "three", ExpStr: "<33:three>", NewA: ptrTo(0), NewB: nil, ExpNewStr: "<0:three>"},
		{A: 4321, B: "four", ExpStr: "<4321:four>", NewA: nil, NewB: ptrTo(""), ExpNewStr: "<4321:>"},
	}

	for _, tc := range tests {
		runPairTest(t, tc)
	}

	runPairTest(t, pairTestCase[string, *testType]{A: "one", B: newTT("ONE"), ExpStr: "<one:*provutils.testType=ONE>"})
	runPairTest(t, pairTestCase[*testType, string]{A: newTT("two"), B: "TWO", ExpStr: "<*provutils.testType=two:TWO>"})
	runPairTest(t, pairTestCase[*testType, *testType]{
		A: newTT("three"), B: newTT("FOUR"), ExpStr: "<*provutils.testType=three:*provutils.testType=FOUR>",
		NewA: ptrTo(newTT("tHrEe")), NewB: ptrTo(newTT("FoUr")), ExpNewStr: "<*provutils.testType=tHrEe:*provutils.testType=FoUr>",
	})
	runPairTest(t, pairTestCase[*testType, *testType]{
		A: newTT("five"), B: nil, ExpStr: "<*provutils.testType=five:<nil>>",
		NewA: ptrTo((*testType)(nil)), NewB: ptrTo(newTT("sIx")), ExpNewStr: "<<nil>:*provutils.testType=sIx>",
	})
	runPairTest(t, pairTestCase[[]byte, string]{
		A: []byte{108, 101, 102, 116}, B: "right", ExpStr: "<[108 101 102 116]:right>",
	})
}

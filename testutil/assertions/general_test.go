package assertions

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	errorLead = "\n\tError:      \t"
	blankLead = "\n\t            \t"
)

// mockTB is a TB to use for testing since there's no way to create
// a divorced T to run tests on things that take in a T.
// It keeps track of output and failure status.
type mockTB struct {
	panic    any
	isFailed bool
	output   string
}

var _ TB = (*mockTB)(nil)

func (t *mockTB) Helper() {}
func (t *mockTB) Errorf(format string, args ...any) {
	t.isFailed = true
	t.output += fmt.Sprintf(format, args...)
}
func (t *mockTB) Logf(format string, args ...any) {
	t.output += fmt.Sprintf(format, args...)
}
func (t *mockTB) FailNow() {
	t.isFailed = true
	runtime.Goexit()
}

// mockRun runs the provided function giving it a new mockTB.
// Returns that mockTB after it's done.
func mockRun(t *testing.T, fn func(TB)) *mockTB {
	signal := make(chan (bool))
	rv := &mockTB{}
	go func() {
		defer func() {
			rv.panic = recover()
		}()
		defer func() {
			signal <- true
		}()
		fn(rv)
	}()
	// Wait for it to finish.
	_ = <-signal
	if rv.panic != nil {
		t.Fatalf("mock test panicked:\n%v", rv.panic)
	}
	return rv
}

// assertMockRunAssertResult applies a set of assertions that check the results of an Assert... function.
func assertMockRunAssertResult(t *testing.T, funcName string, tb *mockTB, success bool, expOutput []string, expMsgAndArgs string) {
	if len(expOutput) > 0 {
		assert.False(t, success, "%s result", funcName)
		assert.True(t, tb.isFailed, "%s is failed", funcName)
		for _, exp := range expOutput {
			assert.Contains(t, tb.output, exp, "%s output", funcName)
		}
		assert.Contains(t, tb.output, expMsgAndArgs, "%s output has msgAndArgs", funcName)
	} else {
		assert.True(t, success, "%s result", funcName)
		assert.False(t, tb.isFailed, "%s is failed", funcName)
		assert.Equal(t, tb.output, strings.Join(expOutput, ""), "%s output", funcName)
	}
}

// assertMockRunAssertResult applies a set of assertions that check the results of an Require... function.
func assertMockRunRequireResult(t *testing.T, funcName string, tb *mockTB, exited bool, expOutput []string, expMsgAndArgs string) {
	if len(expOutput) > 0 {
		assert.True(t, tb.isFailed, "%s is failed", funcName)
		for _, exp := range expOutput {
			assert.Contains(t, tb.output, exp, "%s output", funcName)
		}
		assert.Contains(t, tb.output, expMsgAndArgs, "%s output has msgAndArgs", funcName)
		assert.True(t, exited, "%s exited", funcName)
	} else {
		assert.False(t, tb.isFailed, "%s is failed", funcName)
		assert.Equal(t, tb.output, strings.Join(expOutput, ""), "%s output", funcName)
		assert.False(t, exited, "%s exited", funcName)
	}
}

// notPanicsNoErrorTestCase is a test case for the [Assert|Require]NotPanicsNoError[f]? functions.
type notPanicsNoErrorTestCase struct {
	name      string
	f         ErrorTestFunc
	fCalled   bool
	expOutput []string
}

// newNotPanicsNoError creates a new notPanicsNoErrorTestCase using the info in the provided base.
// The base.f function is wrapped so that fCalled is updated when base.f is called.
func newNotPanicsNoError(base notPanicsNoErrorTestCase) *notPanicsNoErrorTestCase {
	rv := &notPanicsNoErrorTestCase{
		name:      base.name,
		expOutput: base.expOutput,
	}
	rv.f = func() error {
		rv.fCalled = true
		return base.f()
	}
	return rv
}

// getNotPanicsNoErrorsTestCases returns all the tests cases for the [Assert|Require]NotPanicsNoError[f]? functions.
func getNotPanicsNoErrorsTestCases() []*notPanicsNoErrorTestCase {
	return []*notPanicsNoErrorTestCase{
		newNotPanicsNoError(notPanicsNoErrorTestCase{
			name: "no panic no error",
			f: func() error {
				return nil
			},
			expOutput: nil,
		}),
		newNotPanicsNoError(notPanicsNoErrorTestCase{
			name: "panics",
			f: func() error {
				panic("my pen has gone dry")
			},
			expOutput: []string{
				errorLead + "func (assert.PanicTestFunc)",
				"should not panic",
				blankLead + "\tPanic value:\tmy pen has gone dry",
				blankLead + "\tPanic stack:\t",
				"assertions.AssertNotPanicsNoError",
			},
		}),
		newNotPanicsNoError(notPanicsNoErrorTestCase{
			name: "returns an error",
			f: func() error {
				return errors.New("my sock fell on the roof")
			},
			expOutput: []string{
				errorLead + "Received unexpected error:",
				blankLead + "my sock fell on the roof",
			},
		}),
	}
}

func TestAssertNotPanicsNoError(t *testing.T) {
	funcName := "AssertNotPanicsNoError"
	for _, tc := range getNotPanicsNoErrorsTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			msg := "SetupFunc(%s, %q)"
			args := []interface{}{sdk.AccAddress("test_address________"), sdk.NewCoins(sdk.NewInt64Coin("cats", 3))}
			msgAndArgs := append([]interface{}{msg}, args...)
			expMsgAndArgs := "Messages:   \t" + fmt.Sprintf(msg, args...)

			var success bool
			testFunc := func(testTB TB) {
				success = AssertNotPanicsNoError(testTB, tc.f, msgAndArgs...)
			}
			tb := mockRun(t, testFunc)
			require.True(t, tc.fCalled, "%s called test func", funcName)

			assertMockRunAssertResult(t, funcName, tb, success, tc.expOutput, expMsgAndArgs)
		})
	}
}

func TestRequireNotPanicsNoError(t *testing.T) {
	funcName := "RequireNotPanicsNoError"
	for _, tc := range getNotPanicsNoErrorsTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			msg := "%d args: %s, %s"
			args := []interface{}{2, "arg 1", "arg 2"}
			msgAndArgs := append([]interface{}{msg}, args...)
			expMsgAndArgs := "Messages:   \t" + fmt.Sprintf(msg, args...)

			exited := true
			testFunc := func(testTB TB) {
				RequireNotPanicsNoError(testTB, tc.f, msgAndArgs...)
				exited = false
			}
			tb := mockRun(t, testFunc)
			require.True(t, tc.fCalled, "%s called test func", funcName)

			assertMockRunRequireResult(t, funcName, tb, exited, tc.expOutput, expMsgAndArgs)
		})
	}
}

func TestAssertNotPanicsNoErrorf(t *testing.T) {
	funcName := "AssertNotPanicsNoErrorf"
	for _, tc := range getNotPanicsNoErrorsTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			msg := "SetupFunc(%s, %q)"
			args := []interface{}{sdk.AccAddress("test_address________"), sdk.NewCoins(sdk.NewInt64Coin("cats", 3))}
			expMsgAndArgs := "Messages:   \t" + fmt.Sprintf(msg, args...)

			var success bool
			testFunc := func(testTB TB) {
				success = AssertNotPanicsNoErrorf(testTB, tc.f, msg, args...)
			}
			tb := mockRun(t, testFunc)
			require.True(t, tc.fCalled, "%s called test func", funcName)

			assertMockRunAssertResult(t, funcName, tb, success, tc.expOutput, expMsgAndArgs)
		})
	}
}

func TestRequireNotPanicsNoErrorf(t *testing.T) {
	funcName := "RequireNotPanicsNoErrorf"
	for _, tc := range getNotPanicsNoErrorsTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			msg := "%d args: %s, %s"
			args := []interface{}{2, "arg 1", "arg 2"}
			expMsgAndArgs := "Messages:   \t" + fmt.Sprintf(msg, args...)

			exited := true
			testFunc := func(testTB TB) {
				RequireNotPanicsNoErrorf(testTB, tc.f, msg, args...)
				exited = false
			}
			tb := mockRun(t, testFunc)
			require.True(t, tc.fCalled, "%s called test func", funcName)

			assertMockRunRequireResult(t, funcName, tb, exited, tc.expOutput, expMsgAndArgs)
		})
	}
}

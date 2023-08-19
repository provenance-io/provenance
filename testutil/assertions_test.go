package testutil

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/hold"
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

// errorContentsTestCase is a test case for the [Assert|Require]ErrorContents[f]? functions.
type errorContentsTestCase struct {
	name      string
	theError  error
	contains  []string
	expOutput []string
}

// getErrorContentsTestCases returns all the tests cases for the [Assert|Require]ErrorContents[f]? functions.
func getErrorContentsTestCases() []errorContentsTestCase {
	return []errorContentsTestCase{
		{
			name:      "nil error nil contains",
			theError:  nil,
			contains:  nil,
			expOutput: nil,
		},
		{
			name:      "nil error one contains",
			theError:  nil,
			contains:  []string{"not in error"},
			expOutput: []string{"An error is expected but got nil.", "Error was expected to contain: \"not in error\""},
		},
		{
			name:      "nil error two contains",
			theError:  nil,
			contains:  []string{"missing", "fiddlesticks"},
			expOutput: []string{"An error is expected but got nil.", "Error was expected to contain:\n\t\t\"missing\"\n\t\t\"fiddlesticks\""},
		},
		{
			name:      "with error nil contains",
			theError:  fmt.Errorf("test error"),
			contains:  nil,
			expOutput: []string{"Received unexpected error:", "test error"},
		},
		{
			name:      "with error one contains in error",
			theError:  fmt.Errorf("test error"),
			contains:  []string{"test"},
			expOutput: nil,
		},
		{
			name:     "with error one contains not in error",
			theError: fmt.Errorf("test error"),
			contains: []string{"bananas"},
			expOutput: []string{
				errorLead + "1 expected substring(s) missing from error.",
				blankLead + "Actual:\t\"test error\"",
				blankLead + "Missing:\n\t",
				blankLead + "\t1: \"bananas\"",
			},
		},
		{
			name:      "with error three contains all in error",
			theError:  fmt.Errorf("this is a test error"),
			contains:  []string{"test", "error", "this"},
			expOutput: nil,
		},
		{
			name:     "with error three contains first not in error",
			theError: fmt.Errorf("this is a test error"),
			contains: []string{"bananas", "error", "this"},
			expOutput: []string{
				errorLead + "1 expected substring(s) missing from error.",
				blankLead + "Actual:\t\"this is a test error\"",
				blankLead + "Missing:\n\t",
				blankLead + "\t1: \"bananas\"",
			},
		},
		{
			name:     "with error three contains second not in error",
			theError: fmt.Errorf("this is a test error"),
			contains: []string{"this", "bananas", "this"},
			expOutput: []string{
				errorLead + "1 expected substring(s) missing from error.",
				blankLead + "Actual:\t\"this is a test error\"",
				blankLead + "Missing:\n\t",
				blankLead + "\t1: \"bananas\"",
			},
		},
		{
			name:     "with error three contains last not in error",
			theError: fmt.Errorf("this is a test error"),
			contains: []string{"this", "error", "bananas"},
			expOutput: []string{
				errorLead + "1 expected substring(s) missing from error.",
				blankLead + "Actual:\t\"this is a test error\"",
				blankLead + "Missing:\n\t",
				blankLead + "\t1: \"bananas\"",
			},
		},
		{
			name:     "with error three contains none in error",
			theError: fmt.Errorf("test error"),
			contains: []string{"bananas", "not in the error", "farcical"},
			expOutput: []string{
				errorLead + "3 expected substring(s) missing from error.",
				blankLead + "Actual:\t\"test error\"",
				blankLead + "Missing:\n\t",
				blankLead + "\t1: \"bananas\"",
				blankLead + "\t2: \"not in the error\"",
				blankLead + "\t3: \"farcical\"",
			},
		},
	}
}

func TestAssertErrorContents(t *testing.T) {
	funcName := "AssertErrorContents"
	for _, tc := range getErrorContentsTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			msg := "a message with %d args: %v %v %v"
			args := []interface{}{3, "arg 1", "arg 2", "arg 3"}
			msgAndArgs := append([]interface{}{msg}, args...)
			expMsgAndArgs := "Messages:   \t" + fmt.Sprintf(msg, args...)

			var success bool
			testFunc := func(testTB TB) {
				success = AssertErrorContents(testTB, tc.theError, tc.contains, msgAndArgs...)
			}
			tb := mockRun(t, testFunc)

			assertMockRunAssertResult(t, funcName, tb, success, tc.expOutput, expMsgAndArgs)
		})
	}
}

func TestRequireErrorContents(t *testing.T) {
	funcName := "RequireErrorContents"
	for _, tc := range getErrorContentsTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			msg := "a message with %d args: %v %v %v"
			args := []interface{}{3, "arg 1", "arg 2", "arg 3"}
			msgAndArgs := append([]interface{}{msg}, args...)
			expMsgAndArgs := "Messages:   \t" + fmt.Sprintf(msg, args...)

			exited := true
			testFunc := func(testTB TB) {
				RequireErrorContents(testTB, tc.theError, tc.contains, msgAndArgs...)
				exited = false
			}
			tb := mockRun(t, testFunc)

			assertMockRunRequireResult(t, funcName, tb, exited, tc.expOutput, expMsgAndArgs)
		})
	}
}

func TestAssertErrorContentsf(t *testing.T) {
	funcName := "AssertErrorContentsf"
	for _, tc := range getErrorContentsTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			msg := "a message with %d args: %v %v %v"
			args := []interface{}{3, "arg 1", "arg 2", "arg 3"}
			expMsgAndArgs := "Messages:   \t" + fmt.Sprintf(msg, args...)

			var success bool
			testFunc := func(testTB TB) {
				success = AssertErrorContentsf(testTB, tc.theError, tc.contains, msg, args...)
			}
			tb := mockRun(t, testFunc)

			assertMockRunAssertResult(t, funcName, tb, success, tc.expOutput, expMsgAndArgs)
		})
	}
}

func TestRequireErrorContentsf(t *testing.T) {
	funcName := "RequireErrorContentsf"
	for _, tc := range getErrorContentsTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			msg := "a message with %d args: %v %v %v"
			args := []interface{}{3, "arg 1", "arg 2", "arg 3"}
			expMsgAndArgs := "Messages:   \t" + fmt.Sprintf(msg, args...)

			exited := true
			testFunc := func(testTB TB) {
				RequireErrorContentsf(testTB, tc.theError, tc.contains, msg, args...)
				exited = false
			}
			tb := mockRun(t, testFunc)

			assertMockRunRequireResult(t, funcName, tb, exited, tc.expOutput, expMsgAndArgs)
		})
	}
}

// errorValueTestCase is a test case for the [Assert|Require]ErrorValue[f]? functions.
type errorValueTestCase struct {
	name      string
	theError  error
	expected  string
	expOutput []string
}

// getErrorValueTestCases returns all the tests cases for the [Assert|Require]ErrorValue[f]? functions.
func getErrorValueTestCases() []errorValueTestCase {
	return []errorValueTestCase{
		{
			name:      "nil error empty expected",
			theError:  nil,
			expected:  "",
			expOutput: nil,
		},
		{
			name:      "nil error with expected",
			theError:  nil,
			expected:  "bananas",
			expOutput: []string{errorLead + "An error is expected but got nil."},
		},
		{
			name:     "with error empty expected",
			theError: fmt.Errorf("test error"),
			expected: "",
			expOutput: []string{
				errorLead + "Received unexpected error:",
				blankLead + "test error",
			},
		},
		{
			name:      "with error and same expected",
			theError:  fmt.Errorf("test error"),
			expected:  "test error",
			expOutput: nil,
		},
		{
			name:     "with error and different expected",
			theError: fmt.Errorf("this error is bananas"),
			expected: "bananas",
			expOutput: []string{
				errorLead + "Error message not equal",
				blankLead + "expected: \"bananas\"",
				blankLead + "actual  : \"this error is bananas\"",
			},
		},
	}
}

func TestAssertErrorValue(t *testing.T) {
	funcName := "AssertErrorValue"
	for _, tc := range getErrorValueTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			msg := "%d args: %s, %s"
			args := []interface{}{2, "arg 1", "arg 2"}
			msgAndArgs := append([]interface{}{msg}, args...)
			expMsgAndArgs := "Messages:   \t" + fmt.Sprintf(msg, args...)

			var success bool
			testFunc := func(testTB TB) {
				success = AssertErrorValue(testTB, tc.theError, tc.expected, msgAndArgs...)
			}
			tb := mockRun(t, testFunc)

			assertMockRunAssertResult(t, funcName, tb, success, tc.expOutput, expMsgAndArgs)
		})
	}
}

func TestRequireErrorValue(t *testing.T) {
	funcName := "RequireErrorValue"
	for _, tc := range getErrorValueTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			msg := "%d args: %s, %s"
			args := []interface{}{2, "arg 1", "arg 2"}
			msgAndArgs := append([]interface{}{msg}, args...)
			expMsgAndArgs := "Messages:   \t" + fmt.Sprintf(msg, args...)

			exited := true
			testFunc := func(testTB TB) {
				RequireErrorValue(testTB, tc.theError, tc.expected, msgAndArgs...)
				exited = false
			}
			tb := mockRun(t, testFunc)

			assertMockRunRequireResult(t, funcName, tb, exited, tc.expOutput, expMsgAndArgs)
		})
	}
}

func TestAssertErrorValuef(t *testing.T) {
	funcName := "AssertErrorValuef"
	for _, tc := range getErrorValueTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			msg := "%d args: %s, %s"
			args := []interface{}{2, "arg 1", "arg 2"}
			expMsgAndArgs := "Messages:   \t" + fmt.Sprintf(msg, args...)

			var success bool
			testFunc := func(testTB TB) {
				success = AssertErrorValuef(testTB, tc.theError, tc.expected, msg, args...)
			}
			tb := mockRun(t, testFunc)

			assertMockRunAssertResult(t, funcName, tb, success, tc.expOutput, expMsgAndArgs)
		})
	}
}

func TestRequireErrorValuef(t *testing.T) {
	funcName := "RequireErrorValuef"
	for _, tc := range getErrorValueTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			msg := "%d args: %s, %s"
			args := []interface{}{2, "arg 1", "arg 2"}
			expMsgAndArgs := "Messages:   \t" + fmt.Sprintf(msg, args...)

			exited := true
			testFunc := func(testTB TB) {
				RequireErrorValuef(testTB, tc.theError, tc.expected, msg, args...)
				exited = false
			}
			tb := mockRun(t, testFunc)

			assertMockRunRequireResult(t, funcName, tb, exited, tc.expOutput, expMsgAndArgs)
		})
	}
}

func TestDidPanic(t *testing.T) {
	var called bool

	tests := []struct {
		name        string
		f           PanicTestFunc
		expPanicked bool
		expMessage  interface{}
	}{
		{
			name: "does not panic",
			f: func() {
				called = true
			},
			expPanicked: false,
			expMessage:  nil,
		},
		{
			name: "panic with error",
			f: func() {
				called = true
				panic(errors.New("hair should not be on fire"))
			},
			expPanicked: true,
			expMessage:  errors.New("hair should not be on fire"),
		},
		{
			name: "panic with string",
			f: func() {
				called = true
				panic("my hair is on fire")
			},
			expPanicked: true,
			expMessage:  "my hair is on fire",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			called = false
			panicked, message, stack := didPanic(tc.f)
			assert.True(t, called, "test function called")
			assert.Equal(t, tc.expPanicked, panicked, "panicked")
			assert.Equal(t, tc.expMessage, message, "message")
			if panicked {
				assert.Contains(t, stack, "runtime/debug.Stack()", "stack trace")
				assert.Contains(t, stack, "testutil.didPanic.func1()", "stack trace")
				assert.Contains(t, stack, "runtime/panic.go", "stack trace")
				assert.Contains(t, stack, "testutil.TestDidPanic", "stack trace")
				assert.Contains(t, stack, "testutil/assertions_test.go", "stack trace")
			}
		})
	}
}

// panicFuncWithError returns a function that, when called, will panic with the given error.
func panicFuncWithError(msg string) PanicTestFunc {
	return func() {
		panic(errors.New(msg))
	}
}

// panicFuncWithString returns a function that, when called, will panic with the given string.
func panicFuncWithString(msg string) PanicTestFunc {
	return func() {
		panic(msg)
	}
}

// panicContentsTestCase is a test case for the [Assert|Require]PanicContents[f]? functions.
type panicContentsTestCase struct {
	name      string
	f         PanicTestFunc
	contains  []string
	fCalled   bool
	expOutput []string
}

// newPanicContentsTestCase creates a new panicContentsTestCase using the info in the provided base.
// The base.f function is wrapped so that fCalled is updated when base.f is called.
func newPanicContentsTestCase(base panicContentsTestCase) *panicContentsTestCase {
	rv := &panicContentsTestCase{
		name:      base.name,
		contains:  base.contains,
		expOutput: base.expOutput,
	}
	rv.f = func() {
		rv.fCalled = true
		base.f()
	}
	return rv
}

// getPanicContentsTestCases returns all the tests cases for the [Assert|Require]PanicContents[f]? functions.
func getPanicContentsTestCases() []*panicContentsTestCase {
	return []*panicContentsTestCase{
		newPanicContentsTestCase(panicContentsTestCase{
			name:      "does not panic no contains",
			f:         func() {},
			contains:  nil,
			expOutput: nil,
		}),
		newPanicContentsTestCase(panicContentsTestCase{
			name:     "does not panic one contains",
			f:        func() {},
			contains: []string{"bananas"},
			expOutput: []string{
				errorLead + "func (testutil.PanicTestFunc)",
				"should panic, but did not.",
				blankLead + "\tExpected to contain:\t\"bananas\"",
			},
		}),
		newPanicContentsTestCase(panicContentsTestCase{
			name:     "panics no contains",
			f:        panicFuncWithString("arthur dent"),
			contains: nil,
			expOutput: []string{
				errorLead + "func (testutil.PanicTestFunc)",
				"should not panic, but did.",
				blankLead + "\tPanic message:\t\"arthur dent\"",
			},
		}),
		newPanicContentsTestCase(panicContentsTestCase{
			name:      "panics with error as expected",
			f:         panicFuncWithError("this is a panic error"),
			contains:  []string{"this is a", "a panic error"},
			expOutput: nil,
		}),
		newPanicContentsTestCase(panicContentsTestCase{
			name:      "panics with string as expected",
			f:         panicFuncWithString("this is a panic string"),
			contains:  []string{"this is a", "a panic string"},
			expOutput: nil,
		}),
		newPanicContentsTestCase(panicContentsTestCase{
			name:     "error panic missing first contains",
			f:        panicFuncWithError("there is a fly in my soup"),
			contains: []string{"bananas", "fly", "soup"},
			expOutput: []string{
				errorLead + "func (testutil.PanicTestFunc)",
				"panic message incorrect.",
				blankLead + "\t   Panic message:\t\"there is a fly in my soup\"",
				blankLead + "\tDoes not contain:\t\"bananas\"",
			},
		}),
		newPanicContentsTestCase(panicContentsTestCase{
			name:     "error panic missing second contains",
			f:        panicFuncWithError("there is a fly in my soup"),
			contains: []string{"there is a", "bananas", "soup"},
			expOutput: []string{
				errorLead + "func (testutil.PanicTestFunc)",
				"panic message incorrect.",
				blankLead + "\t   Panic message:\t\"there is a fly in my soup\"",
				blankLead + "\tDoes not contain:\t\"bananas\"",
			},
		}),
		newPanicContentsTestCase(panicContentsTestCase{
			name:     "error panic missing third contains",
			f:        panicFuncWithError("there is a fly in my soup"),
			contains: []string{"there is a", "fly", "bananas"},
			expOutput: []string{
				errorLead + "func (testutil.PanicTestFunc)",
				"panic message incorrect.",
				blankLead + "\t   Panic message:\t\"there is a fly in my soup\"",
				blankLead + "\tDoes not contain:\t\"bananas\"",
			},
		}),
		newPanicContentsTestCase(panicContentsTestCase{
			name:      "error panic with three contains",
			f:         panicFuncWithError("there is a fly in my soup"),
			contains:  []string{"there is a", "fly", "soup"},
			expOutput: nil,
		}),
		newPanicContentsTestCase(panicContentsTestCase{
			name:     "string panic missing first contains",
			f:        panicFuncWithString("there is a fly in my soup"),
			contains: []string{"bananas", "fly", "soup"},
			expOutput: []string{
				errorLead + "func (testutil.PanicTestFunc)",
				"panic message incorrect.",
				blankLead + "\t   Panic message:\t\"there is a fly in my soup\"",
				blankLead + "\tDoes not contain:\t\"bananas\"",
			},
		}),
		newPanicContentsTestCase(panicContentsTestCase{
			name:     "string panic missing second contains",
			f:        panicFuncWithString("there is a fly in my soup"),
			contains: []string{"there is a", "bananas", "soup"},
			expOutput: []string{
				errorLead + "func (testutil.PanicTestFunc)",
				"panic message incorrect.",
				blankLead + "\t   Panic message:\t\"there is a fly in my soup\"",
				blankLead + "\tDoes not contain:\t\"bananas\"",
			},
		}),
		newPanicContentsTestCase(panicContentsTestCase{
			name:     "string panic missing third contains",
			f:        panicFuncWithString("there is a fly in my soup"),
			contains: []string{"there is a", "fly", "bananas"},
			expOutput: []string{
				errorLead + "func (testutil.PanicTestFunc)",
				"panic message incorrect.",
				blankLead + "\t   Panic message:\t\"there is a fly in my soup\"",
				blankLead + "\tDoes not contain:\t\"bananas\"",
			},
		}),
		newPanicContentsTestCase(panicContentsTestCase{
			name:      "string panic with three contains",
			f:         panicFuncWithString("there is a fly in my soup"),
			contains:  []string{"there is a", "fly", "soup"},
			expOutput: nil,
		}),
	}
}

func TestAssertPanicContents(t *testing.T) {
	funcName := "AssertPanicContents"
	for _, tc := range getPanicContentsTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			msg := "%d args: %s, %s"
			args := []interface{}{2, "arg 1", "arg 2"}
			msgAndArgs := append([]interface{}{msg}, args...)
			expMsgAndArgs := "Messages:   \t" + fmt.Sprintf(msg, args...)

			var success bool
			testFunc := func(testTB TB) {
				success = AssertPanicContents(testTB, tc.f, tc.contains, msgAndArgs...)
			}
			tb := mockRun(t, testFunc)
			require.True(t, tc.fCalled, "%s called test func", funcName)

			assertMockRunAssertResult(t, funcName, tb, success, tc.expOutput, expMsgAndArgs)
		})
	}
}

func TestRequirePanicContents(t *testing.T) {
	funcName := "RequirePanicContents"
	for _, tc := range getPanicContentsTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			msg := "%d args: %s, %s"
			args := []interface{}{2, "arg 1", "arg 2"}
			msgAndArgs := append([]interface{}{msg}, args...)
			expMsgAndArgs := "Messages:   \t" + fmt.Sprintf(msg, args...)

			exited := true
			testFunc := func(testTB TB) {
				RequirePanicContents(testTB, tc.f, tc.contains, msgAndArgs...)
				exited = false
			}
			tb := mockRun(t, testFunc)
			require.True(t, tc.fCalled, "%s called test func", funcName)

			assertMockRunRequireResult(t, funcName, tb, exited, tc.expOutput, expMsgAndArgs)
		})
	}
}

func TestAssertPanicContentsf(t *testing.T) {
	funcName := "AssertPanicContentsf"
	for _, tc := range getPanicContentsTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			msg := "%d args: %s, %s"
			args := []interface{}{2, "arg 1", "arg 2"}
			expMsgAndArgs := "Messages:   \t" + fmt.Sprintf(msg, args...)

			var success bool
			testFunc := func(testTB TB) {
				success = AssertPanicContentsf(testTB, tc.f, tc.contains, msg, args...)
			}
			tb := mockRun(t, testFunc)
			require.True(t, tc.fCalled, "%s called test func", funcName)

			assertMockRunAssertResult(t, funcName, tb, success, tc.expOutput, expMsgAndArgs)
		})
	}
}

func TestRequirePanicContentsf(t *testing.T) {
	funcName := "RequirePanicContentsf"
	for _, tc := range getPanicContentsTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			msg := "%d args: %s, %s"
			args := []interface{}{2, "arg 1", "arg 2"}
			expMsgAndArgs := "Messages:   \t" + fmt.Sprintf(msg, args...)

			exited := true
			testFunc := func(testTB TB) {
				RequirePanicContentsf(testTB, tc.f, tc.contains, msg, args...)
				exited = false
			}
			tb := mockRun(t, testFunc)
			require.True(t, tc.fCalled, "%s called test func", funcName)

			assertMockRunRequireResult(t, funcName, tb, exited, tc.expOutput, expMsgAndArgs)
		})
	}
}

// panicContentsTestCase is a test case for the [Assert|Require]PanicEquals[f]? functions.
type panicEqualsTestCase struct {
	name      string
	f         PanicTestFunc
	expected  string
	fCalled   bool
	expOutput []string
}

// newPanicEqualsTestCase creates a new panicEqualsTestCase using the info in the provided base.
// The base.f function is wrapped so that fCalled is updated when base.f is called.
func newPanicEqualsTestCase(base panicEqualsTestCase) *panicEqualsTestCase {
	rv := &panicEqualsTestCase{
		name:      base.name,
		expected:  base.expected,
		expOutput: base.expOutput,
	}
	rv.f = func() {
		rv.fCalled = true
		base.f()
	}
	return rv
}

// getPanicEqualsTestCases returns all the tests cases for the [Assert|Require]PanicEquals[f]? functions.
func getPanicEqualsTestCases() []*panicEqualsTestCase {
	return []*panicEqualsTestCase{
		newPanicEqualsTestCase(panicEqualsTestCase{
			name:      "no panic none expected",
			f:         func() {},
			expected:  "",
			expOutput: nil,
		}),
		newPanicEqualsTestCase(panicEqualsTestCase{
			name:     "no panic but was expecting it to",
			f:        func() {},
			expected: "bananas",
			expOutput: []string{
				errorLead + "func (testutil.PanicTestFunc)",
				"did not panic, but should.",
				blankLead + "\tExpected message:\t\"bananas\"",
			},
		}),
		newPanicEqualsTestCase(panicEqualsTestCase{
			name:     "panics but not expected",
			f:        panicFuncWithError("this dip is tepid"),
			expected: "",
			expOutput: []string{
				errorLead + "func (testutil.PanicTestFunc)",
				"should not panic, but did.",
				blankLead + "\tPanic message:\t\"this dip is tepid\"",
				blankLead + "\t  Panic value:\t&errors.errorString{s:\"this dip is tepid\"}",
				blankLead + "\t  Panic stack:",
				"testutil.AssertPanicEquals",
			},
		}),
		newPanicEqualsTestCase(panicEqualsTestCase{
			name:      "panics with error as expected",
			f:         panicFuncWithError("this dip is tepid"),
			expected:  "this dip is tepid",
			expOutput: nil,
		}),
		newPanicEqualsTestCase(panicEqualsTestCase{
			name:      "panics with string as expected",
			f:         panicFuncWithString("this dip is tepid"),
			expected:  "this dip is tepid",
			expOutput: nil,
		}),
		newPanicEqualsTestCase(panicEqualsTestCase{
			name:     "panics with error different from expected",
			f:        panicFuncWithError("this dip is tepid"),
			expected: "this dip is ice cold",
			expOutput: []string{
				errorLead + "func (testutil.PanicTestFunc)",
				"panic message incorrect.",
				blankLead + "\t   Panic message:\t\"this dip is tepid\"",
				blankLead + "\tExpected message:\t\"this dip is ice cold\"",
				blankLead + "\tPanic value:\t&errors.errorString{s:\"this dip is tepid\"}",
				blankLead + "\tPanic stack:",
				"testutil.AssertPanicEquals",
			},
		}),
		newPanicEqualsTestCase(panicEqualsTestCase{
			name:     "panics with string different from expected",
			f:        panicFuncWithString("this dip is tepid"),
			expected: "this dip is ice cold",
			expOutput: []string{
				errorLead + "func (testutil.PanicTestFunc)",
				"panic message incorrect.",
				blankLead + "\t   Panic message:\t\"this dip is tepid\"",
				blankLead + "\tExpected message:\t\"this dip is ice cold\"",
				blankLead + "\tPanic value:\t\"this dip is tepid\"",
				blankLead + "\tPanic stack:",
				"testutil.AssertPanicEquals",
			},
		}),
	}
}

func TestAssertPanicEquals(t *testing.T) {
	funcName := "AssertPanicEquals"
	for _, tc := range getPanicEqualsTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			msg := "%d args: %s, %s"
			args := []interface{}{2, "arg 1", "arg 2"}
			msgAndArgs := append([]interface{}{msg}, args...)
			expMsgAndArgs := "Messages:   \t" + fmt.Sprintf(msg, args...)

			var success bool
			testFunc := func(testTB TB) {
				success = AssertPanicEquals(testTB, tc.f, tc.expected, msgAndArgs...)
			}
			tb := mockRun(t, testFunc)
			require.True(t, tc.fCalled, "%s called test func", funcName)

			assertMockRunAssertResult(t, funcName, tb, success, tc.expOutput, expMsgAndArgs)
		})
	}
}

func TestRequirePanicEquals(t *testing.T) {
	funcName := "RequirePanicEquals"
	for _, tc := range getPanicEqualsTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			msg := "%d args: %s, %s"
			args := []interface{}{2, "arg 1", "arg 2"}
			msgAndArgs := append([]interface{}{msg}, args...)
			expMsgAndArgs := "Messages:   \t" + fmt.Sprintf(msg, args...)

			exited := true
			testFunc := func(testTB TB) {
				RequirePanicEquals(testTB, tc.f, tc.expected, msgAndArgs...)
				exited = false
			}
			tb := mockRun(t, testFunc)
			require.True(t, tc.fCalled, "%s called test func", funcName)

			assertMockRunRequireResult(t, funcName, tb, exited, tc.expOutput, expMsgAndArgs)
		})
	}
}

func TestAssertPanicEqualsf(t *testing.T) {
	funcName := "AssertPanicEqualsf"
	for _, tc := range getPanicEqualsTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			msg := "%d args: %s, %s"
			args := []interface{}{2, "arg 1", "arg 2"}
			expMsgAndArgs := "Messages:   \t" + fmt.Sprintf(msg, args...)

			var success bool
			testFunc := func(testTB TB) {
				success = AssertPanicEqualsf(testTB, tc.f, tc.expected, msg, args...)
			}
			tb := mockRun(t, testFunc)
			require.True(t, tc.fCalled, "%s called test func", funcName)

			assertMockRunAssertResult(t, funcName, tb, success, tc.expOutput, expMsgAndArgs)
		})
	}
}

func TestRequirePanicEqualsf(t *testing.T) {
	funcName := "RequirePanicEqualsf"
	for _, tc := range getPanicEqualsTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			msg := "%d args: %s, %s"
			args := []interface{}{2, "arg 1", "arg 2"}
			expMsgAndArgs := "Messages:   \t" + fmt.Sprintf(msg, args...)

			exited := true
			testFunc := func(testTB TB) {
				RequirePanicEqualsf(testTB, tc.f, tc.expected, msg, args...)
				exited = false
			}
			tb := mockRun(t, testFunc)
			require.True(t, tc.fCalled, "%s called test func", funcName)

			assertMockRunRequireResult(t, funcName, tb, exited, tc.expOutput, expMsgAndArgs)
		})
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
				"testutil.AssertNotPanicsNoError",
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

func TestEventsToStrings(t *testing.T) {
	t.Run("no events", func(t *testing.T) {
		var events sdk.Events
		actual := eventsToStrings(events)
		assert.Empty(t, actual, "eventsToStrings(nil)")
	})

	t.Run("two events", func(t *testing.T) {
		// This test is just making sure that the strings generated by eventsToStrings have
		// all the needed info in them for accurate comparisons.
		coins := func(amounts string) sdk.Coins {
			rv, err := sdk.ParseCoinsNormalized(amounts)
			require.NoError(t, err, "ParseCoinsNormalized(%q)", amounts)
			return rv
		}

		addrAdd := sdk.AccAddress("address_add_event___")
		coinsAdd := coins("97acorn,12banana")
		reason := "just some test reason"
		eventAddT := hold.NewEventHoldAdded(addrAdd, coinsAdd, reason)
		eventAdd, err := sdk.TypedEventToEvent(eventAddT)
		require.NoError(t, err, "TypedEventToEvent EventHoldAdded")

		addrRel := sdk.AccAddress("address_rel_event___")
		coinsRel := coins("13cucumber,81dill")
		eventRelT := hold.NewEventHoldReleased(addrRel, coinsRel)
		eventRel, err := sdk.TypedEventToEvent(eventRelT)
		require.NoError(t, err, "TypedEventToEvent EventHoldReleased")

		events := sdk.Events{
			eventAdd,
			eventRel,
		}

		// Set the index flag on the first attribute of the first event so we make sure that makes a difference.
		events[0].Attributes[0].Index = true

		expected := []string{
			fmt.Sprintf("[0]provenance.hold.v1.EventHoldAdded[0]: \"address\" = \"\\\"%s\\\"\" (indexed)", addrAdd.String()),
			fmt.Sprintf("[0]provenance.hold.v1.EventHoldAdded[1]: \"amount\" = \"\\\"%s\\\"\"", coinsAdd.String()),
			fmt.Sprintf("[0]provenance.hold.v1.EventHoldAdded[2]: \"reason\" = \"\\\"%s\\\"\"", reason),
			fmt.Sprintf("[1]provenance.hold.v1.EventHoldReleased[0]: \"address\" = \"\\\"%s\\\"\"", addrRel.String()),
			fmt.Sprintf("[1]provenance.hold.v1.EventHoldReleased[1]: \"amount\" = \"\\\"%s\\\"\"", coinsRel.String()),
		}

		actual := eventsToStrings(events)
		assert.Equal(t, expected, actual, "events strings")
	})
}

// equalEventsTestCase is a test case for the [Assert|Require]EqualEvents[f]? functions.
type equalEventsTestCase struct {
	name      string
	expected  sdk.Events
	actual    sdk.Events
	expOutput []string
}

// getEqualEventsTestCases returns all the tests cases for the [Assert|Require]EqualEvents[f]? functions.
func getEqualEventsTestCases() []equalEventsTestCase {
	attr := func(key, value string) abci.EventAttribute {
		return abci.EventAttribute{
			Key:   []byte(key),
			Value: []byte(value),
		}
	}
	attri := func(key, value string) abci.EventAttribute {
		return abci.EventAttribute{
			Key:   []byte(key),
			Value: []byte(value),
			Index: true,
		}
	}
	event := func(name string, attrs ...abci.EventAttribute) sdk.Event {
		return sdk.Event{
			Type:       name,
			Attributes: attrs,
		}
	}

	return []equalEventsTestCase{
		{
			name:      "nil nil",
			expected:  nil,
			actual:    nil,
			expOutput: nil,
		},
		{
			name:      "nil empty",
			expected:  nil,
			actual:    sdk.Events{},
			expOutput: nil,
		},
		{
			name:      "empty nil",
			expected:  sdk.Events{},
			actual:    nil,
			expOutput: nil,
		},
		{
			name:      "empty empty",
			expected:  sdk.Events{},
			actual:    sdk.Events{},
			expOutput: nil,
		},
		{
			name:     "nil one",
			expected: nil,
			actual:   sdk.Events{event("missing", attr("mk", "mv"))},
			expOutput: []string{
				errorLead + "Not equal:",
				blankLead + "expected: []string(nil)",
				blankLead + "actual  : []string{\"[0]missing[0]: \\\"mk\\\" = \\\"mv\\\"\"}",
				"+ (string) (len=26) \"[0]missing[0]: \\\"mk\\\" = \\\"mv\\\"\"",
			},
		},
		{
			name:     "one nil",
			expected: sdk.Events{event("missing", attr("mk", "mv"))},
			actual:   nil,
			expOutput: []string{
				errorLead + "Not equal:",
				blankLead + "expected: []string{\"[0]missing[0]: \\\"mk\\\" = \\\"mv\\\"\"}",
				blankLead + "actual  : []string(nil)",
				"- (string) (len=26) \"[0]missing[0]: \\\"mk\\\" = \\\"mv\\\"\"",
			},
		},
		{
			name:      "one and same",
			expected:  sdk.Events{event("found", attr("key1", "value1"))},
			actual:    sdk.Events{event("found", attr("key1", "value1"))},
			expOutput: nil,
		},
		{
			name:     "one with different attribute key",
			expected: sdk.Events{event("found", attr("key1", "value1"))},
			actual:   sdk.Events{event("found", attr("key2", "value1"))},
			expOutput: []string{
				errorLead + "Not equal:",
				blankLead + "expected: []string{\"[0]found[0]: \\\"key1\\\" = \\\"value1\\\"\"}",
				blankLead + "actual  : []string{\"[0]found[0]: \\\"key2\\\" = \\\"value1\\\"\"}",
				"- (string) (len=30) \"[0]found[0]: \\\"key1\\\" = \\\"value1\\\"\"",
				"+ (string) (len=30) \"[0]found[0]: \\\"key2\\\" = \\\"value1\\\"\"",
			},
		},
		{
			name:     "one with different attribute value",
			expected: sdk.Events{event("found", attr("key1", "value1"))},
			actual:   sdk.Events{event("found", attr("key1", "value2"))},
			expOutput: []string{
				errorLead + "Not equal:",
				blankLead + "expected: []string{\"[0]found[0]: \\\"key1\\\" = \\\"value1\\\"\"}",
				blankLead + "actual  : []string{\"[0]found[0]: \\\"key1\\\" = \\\"value2\\\"\"}",
				"- (string) (len=30) \"[0]found[0]: \\\"key1\\\" = \\\"value1\\\"\"",
				"+ (string) (len=30) \"[0]found[0]: \\\"key1\\\" = \\\"value2\\\"\"",
			},
		},
		{
			name:     "one expected index",
			expected: sdk.Events{event("found", attri("key1", "value1"))},
			actual:   sdk.Events{event("found", attr("key1", "value1"))},
			expOutput: []string{
				errorLead + "Not equal:",
				blankLead + "expected: []string{\"[0]found[0]: \\\"key1\\\" = \\\"value1\\\" (indexed)\"}",
				blankLead + "actual  : []string{\"[0]found[0]: \\\"key1\\\" = \\\"value1\\\"\"}",
				"- (string) (len=40) \"[0]found[0]: \\\"key1\\\" = \\\"value1\\\" (indexed)\"",
				"+ (string) (len=30) \"[0]found[0]: \\\"key1\\\" = \\\"value1\\\"\"",
			},
		},
		{
			name:     "one actual index",
			expected: sdk.Events{event("found", attr("key1", "value1"))},
			actual:   sdk.Events{event("found", attri("key1", "value1"))},
			expOutput: []string{
				errorLead + "Not equal:",
				blankLead + "expected: []string{\"[0]found[0]: \\\"key1\\\" = \\\"value1\\\"\"}",
				blankLead + "actual  : []string{\"[0]found[0]: \\\"key1\\\" = \\\"value1\\\" (indexed)\"}",
				"- (string) (len=30) \"[0]found[0]: \\\"key1\\\" = \\\"value1\\\"\"",
				"+ (string) (len=40) \"[0]found[0]: \\\"key1\\\" = \\\"value1\\\" (indexed)\"",
			},
		},
		{
			name:     "one expected extra attribute",
			expected: sdk.Events{event("found", attr("key1", "value1"), attr("key2", "value2"))},
			actual:   sdk.Events{event("found", attr("key1", "value1"))},
			expOutput: []string{
				errorLead + "Not equal:",
				blankLead + "expected: []string{\"[0]found[0]: \\\"key1\\\" = \\\"value1\\\"\", \"[0]found[1]: \\\"key2\\\" = \\\"value2\\\"\"}",
				blankLead + "actual  : []string{\"[0]found[0]: \\\"key1\\\" = \\\"value1\\\"\"}",
				"- (string) (len=30) \"[0]found[1]: \\\"key2\\\" = \\\"value2\\\"\"",
			},
		},
		{
			name:     "one actual extra attribute",
			expected: sdk.Events{event("found", attr("key1", "value1"))},
			actual:   sdk.Events{event("found", attr("key1", "value1"), attr("key2", "value2"))},
			expOutput: []string{
				errorLead + "Not equal:",
				blankLead + "expected: []string{\"[0]found[0]: \\\"key1\\\" = \\\"value1\\\"\"}",
				blankLead + "actual  : []string{\"[0]found[0]: \\\"key1\\\" = \\\"value1\\\"\", \"[0]found[1]: \\\"key2\\\" = \\\"value2\\\"\"}",
				"+ (string) (len=30) \"[0]found[1]: \\\"key2\\\" = \\\"value2\\\"\"",
			},
		},
		{
			name:     "one diff attr order",
			expected: sdk.Events{event("found", attr("key1", "value1"), attr("key2", "value2"))},
			actual:   sdk.Events{event("found", attr("key2", "value2"), attr("key1", "value1"))},
			expOutput: []string{
				errorLead + "Not equal:",
				blankLead + "expected: []string{\"[0]found[0]: \\\"key1\\\" = \\\"value1\\\"\", \"[0]found[1]: \\\"key2\\\" = \\\"value2\\\"\"}",
				blankLead + "actual  : []string{\"[0]found[0]: \\\"key2\\\" = \\\"value2\\\"\", \"[0]found[1]: \\\"key1\\\" = \\\"value1\\\"\"}",
				"- (string) (len=30) \"[0]found[0]: \\\"key1\\\" = \\\"value1\\\"\"",
				"- (string) (len=30) \"[0]found[1]: \\\"key2\\\" = \\\"value2\\\"\"",
				"+ (string) (len=30) \"[0]found[0]: \\\"key2\\\" = \\\"value2\\\"\"",
				"+ (string) (len=30) \"[0]found[1]: \\\"key1\\\" = \\\"value1\\\"\"",
			},
		},
		{
			name: "extra expected",
			expected: sdk.Events{
				event("found", attr("key1", "value1")),
				event("missing", attr("key2", "value2")),
			},
			actual: sdk.Events{event("found", attr("key1", "value1"))},
			expOutput: []string{
				errorLead + "Not equal:",
				blankLead + "expected: []string{\"[0]found[0]: \\\"key1\\\" = \\\"value1\\\"\", \"[1]missing[0]: \\\"key2\\\" = \\\"value2\\\"\"}",
				blankLead + "actual  : []string{\"[0]found[0]: \\\"key1\\\" = \\\"value1\\\"\"}",
				"- (string) (len=32) \"[1]missing[0]: \\\"key2\\\" = \\\"value2\\\"\"",
			},
		},
		{
			name:     "extra actual",
			expected: sdk.Events{event("found", attr("key1", "value1"))},
			actual: sdk.Events{
				event("found", attr("key1", "value1")),
				event("missing", attr("key2", "value2")),
			},
			expOutput: []string{
				errorLead + "Not equal:",
				blankLead + "expected: []string{\"[0]found[0]: \\\"key1\\\" = \\\"value1\\\"\"}",
				blankLead + "actual  : []string{\"[0]found[0]: \\\"key1\\\" = \\\"value1\\\"\", \"[1]missing[0]: \\\"key2\\\" = \\\"value2\\\"\"}",
				"+ (string) (len=32) \"[1]missing[0]: \\\"key2\\\" = \\\"value2\\\"\"",
			},
		},
		{
			name: "two in same order",
			expected: sdk.Events{
				event("first", attr("key1", "value1")),
				event("second", attr("key2", "value2")),
			},
			actual: sdk.Events{
				event("first", attr("key1", "value1")),
				event("second", attr("key2", "value2")),
			},
			expOutput: nil,
		},
		{
			name: "two in different order",
			expected: sdk.Events{
				event("first", attr("key1", "value1")),
				event("second", attr("key2", "value2")),
			},
			actual: sdk.Events{
				event("second", attr("key2", "value2")),
				event("first", attr("key1", "value1")),
			},
			expOutput: []string{
				errorLead + "Not equal:",
				blankLead + "expected: []string{\"[0]first[0]: \\\"key1\\\" = \\\"value1\\\"\", \"[1]second[0]: \\\"key2\\\" = \\\"value2\\\"\"}",
				blankLead + "actual  : []string{\"[0]second[0]: \\\"key2\\\" = \\\"value2\\\"\", \"[1]first[0]: \\\"key1\\\" = \\\"value1\\\"\"}",
				"- (string) (len=30) \"[0]first[0]: \\\"key1\\\" = \\\"value1\\\"\"",
				"- (string) (len=31) \"[1]second[0]: \\\"key2\\\" = \\\"value2\\\"\"",
				"+ (string) (len=31) \"[0]second[0]: \\\"key2\\\" = \\\"value2\\\"\"",
				"+ (string) (len=30) \"[1]first[0]: \\\"key1\\\" = \\\"value1\\\"\"",
			},
		},
	}
}

func TestAssertEqualEvents(t *testing.T) {
	funcName := "AssertEqualEvents"
	for _, tc := range getEqualEventsTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			msg := "msg with %d args: %q %q"
			args := []interface{}{2, "msg arg 1", "msg arg 2"}
			msgAndArgs := append([]interface{}{msg}, args...)
			expMsgAndArgs := "Messages:   \t" + fmt.Sprintf(msg, args...)

			var success bool
			testFunc := func(testTB TB) {
				success = AssertEqualEvents(testTB, tc.expected, tc.actual, msgAndArgs...)
			}
			tb := mockRun(t, testFunc)

			assertMockRunAssertResult(t, funcName, tb, success, tc.expOutput, expMsgAndArgs)
		})
	}
}

func TestRequireEqualEvents(t *testing.T) {
	funcName := "RequireEqualEvents"
	for _, tc := range getEqualEventsTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			msg := "msg with %d args: %q %q"
			args := []interface{}{2, "msg arg 1", "msg arg 2"}
			msgAndArgs := append([]interface{}{msg}, args...)
			expMsgAndArgs := "Messages:   \t" + fmt.Sprintf(msg, args...)

			exited := true
			testFunc := func(testTB TB) {
				RequireEqualEvents(testTB, tc.expected, tc.actual, msgAndArgs...)
				exited = false
			}
			tb := mockRun(t, testFunc)

			assertMockRunRequireResult(t, funcName, tb, exited, tc.expOutput, expMsgAndArgs)
		})
	}
}

func TestAssertEqualEventsf(t *testing.T) {
	funcName := "AssertEqualEventsf"
	for _, tc := range getEqualEventsTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			msg := "msg with %d args: %q %q"
			args := []interface{}{2, "msg arg 1", "msg arg 2"}
			expMsgAndArgs := "Messages:   \t" + fmt.Sprintf(msg, args...)

			var success bool
			testFunc := func(testTB TB) {
				success = AssertEqualEventsf(testTB, tc.expected, tc.actual, msg, args...)
			}
			tb := mockRun(t, testFunc)

			assertMockRunAssertResult(t, funcName, tb, success, tc.expOutput, expMsgAndArgs)
		})
	}
}

func TestRequireEqualEventsf(t *testing.T) {
	funcName := "RequireEqualEventsf"
	for _, tc := range getEqualEventsTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			msg := "msg with %d args: %q %q"
			args := []interface{}{2, "msg arg 1", "msg arg 2"}
			expMsgAndArgs := "Messages:   \t" + fmt.Sprintf(msg, args...)

			exited := true
			testFunc := func(testTB TB) {
				RequireEqualEventsf(testTB, tc.expected, tc.actual, msg, args...)
				exited = false
			}
			tb := mockRun(t, testFunc)

			assertMockRunRequireResult(t, funcName, tb, exited, tc.expOutput, expMsgAndArgs)
		})
	}
}

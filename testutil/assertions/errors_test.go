package assertions

import (
	"errors"
	"fmt"
	"testing"
)

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
			theError:  errors.New("test error"),
			contains:  nil,
			expOutput: []string{"Received unexpected error:", "test error"},
		},
		{
			name:      "with error one contains in error",
			theError:  errors.New("test error"),
			contains:  []string{"test"},
			expOutput: nil,
		},
		{
			name:     "with error one contains not in error",
			theError: errors.New("test error"),
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
			theError:  errors.New("this is a test error"),
			contains:  []string{"test", "error", "this"},
			expOutput: nil,
		},
		{
			name:     "with error three contains first not in error",
			theError: errors.New("this is a test error"),
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
			theError: errors.New("this is a test error"),
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
			theError: errors.New("this is a test error"),
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
			theError: errors.New("test error"),
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
			theError: errors.New("test error"),
			expected: "",
			expOutput: []string{
				errorLead + "Received unexpected error:",
				blankLead + "test error",
			},
		},
		{
			name:      "with error and same expected",
			theError:  errors.New("test error"),
			expected:  "test error",
			expOutput: nil,
		},
		{
			name:     "with error and different expected",
			theError: errors.New("this error is bananas"),
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

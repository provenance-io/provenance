package assertions

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
		{
			name: "panic with anonymous struct",
			f: func() {
				called = true
				panic(struct {
					name  string
					value int
				}{name: "you don't know me", value: 43})
			},
			expPanicked: true,
			expMessage: struct {
				name  string
				value int
			}{name: "you don't know me", value: 43},
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
				assert.Contains(t, stack, "assertions.didPanic.func1()", "stack trace")
				assert.Contains(t, stack, "runtime/panic.go", "stack trace")
				assert.Contains(t, stack, "assertions.TestDidPanic", "stack trace")
				assert.Contains(t, stack, "assertions/panics_test.go", "stack trace")
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
				errorLead + "func (assertions.PanicTestFunc)",
				"should panic, but did not.",
				blankLead + "\tExpected to contain:\t\"bananas\"",
			},
		}),
		newPanicContentsTestCase(panicContentsTestCase{
			name:     "panics no contains",
			f:        panicFuncWithString("arthur dent"),
			contains: nil,
			expOutput: []string{
				errorLead + "func (assertions.PanicTestFunc)",
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
			name: "panics with struct as expected",
			f: func() {
				panic(struct {
					name  string
					value int
				}{name: "who am i", value: 43})
			},
			contains:  []string{"who am i", "43"},
			expOutput: nil,
		}),
		newPanicContentsTestCase(panicContentsTestCase{
			name: "panics with struct missing contains",
			f: func() {
				panic(struct {
					name  string
					value int
				}{name: "who am i", value: 43})
			},
			contains: []string{"who am i", "43", "bananas"},
			expOutput: []string{
				errorLead + "func (assertions.PanicTestFunc)",
				"panic message incorrect.",
				blankLead + "\t   Panic message:\t\"{who am i 43}\"",
				blankLead + "\tDoes not contain:\t\"bananas\"",
			},
		}),
		newPanicContentsTestCase(panicContentsTestCase{
			name:     "error panic missing first contains",
			f:        panicFuncWithError("there is a fly in my soup"),
			contains: []string{"bananas", "fly", "soup"},
			expOutput: []string{
				errorLead + "func (assertions.PanicTestFunc)",
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
				errorLead + "func (assertions.PanicTestFunc)",
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
				errorLead + "func (assertions.PanicTestFunc)",
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
				errorLead + "func (assertions.PanicTestFunc)",
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
				errorLead + "func (assertions.PanicTestFunc)",
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
				errorLead + "func (assertions.PanicTestFunc)",
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
				errorLead + "func (assertions.PanicTestFunc)",
				"did not panic, but should.",
				blankLead + "\tExpected message:\t\"bananas\"",
			},
		}),
		newPanicEqualsTestCase(panicEqualsTestCase{
			name:     "panics but not expected",
			f:        panicFuncWithError("this dip is tepid"),
			expected: "",
			expOutput: []string{
				errorLead + "func (assertions.PanicTestFunc)",
				"should not panic, but did.",
				blankLead + "\tPanic message:\t\"this dip is tepid\"",
				blankLead + "\t  Panic value:\t&errors.errorString{s:\"this dip is tepid\"}",
				blankLead + "\t  Panic stack:",
				"assertions.AssertPanicEquals",
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
			name: "panics with struct as expected",
			f: func() {
				panic(struct {
					msg   string
					value int
				}{msg: "no paint for outhouse", value: 986})
			},
			expected:  "{no paint for outhouse 986}",
			expOutput: nil,
		}),
		newPanicEqualsTestCase(panicEqualsTestCase{
			name:     "panics with error different from expected",
			f:        panicFuncWithError("this dip is tepid"),
			expected: "this dip is ice cold",
			expOutput: []string{
				errorLead + "func (assertions.PanicTestFunc)",
				"panic message incorrect.",
				blankLead + "\t   Panic message:\t\"this dip is tepid\"",
				blankLead + "\tExpected message:\t\"this dip is ice cold\"",
				blankLead + "\tPanic value:\t&errors.errorString{s:\"this dip is tepid\"}",
				blankLead + "\tPanic stack:",
				"assertions.AssertPanicEquals",
			},
		}),
		newPanicEqualsTestCase(panicEqualsTestCase{
			name:     "panics with string different from expected",
			f:        panicFuncWithString("this dip is tepid"),
			expected: "this dip is ice cold",
			expOutput: []string{
				errorLead + "func (assertions.PanicTestFunc)",
				"panic message incorrect.",
				blankLead + "\t   Panic message:\t\"this dip is tepid\"",
				blankLead + "\tExpected message:\t\"this dip is ice cold\"",
				blankLead + "\tPanic value:\t\"this dip is tepid\"",
				blankLead + "\tPanic stack:",
				"assertions.AssertPanicEquals",
			},
		}),
		newPanicEqualsTestCase(panicEqualsTestCase{
			name: "panics with struct different from expected",
			f: func() {
				panic(struct {
					msg   string
					value int
				}{msg: "no paint for outhouse", value: 986})
			},
			expected: "{no aint for outhouse 987}",
			expOutput: []string{
				errorLead + "func (assertions.PanicTestFunc)",
				"panic message incorrect.",
				blankLead + "\t   Panic message:\t\"{no paint for outhouse 986}\"",
				blankLead + "\tExpected message:\t\"{no aint for outhouse 987}\"",
				blankLead + "\tPanic value:\tstruct { msg string; value int }{msg:\"no paint for outhouse\", value:986}",
				blankLead + "\tPanic stack:",
				"assertions.AssertPanicEquals",
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

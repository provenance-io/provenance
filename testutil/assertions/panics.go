package assertions

import (
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/stretchr/testify/assert"
)

// PanicTestFunc is a type declaration for a function that will be tested for panic.
type PanicTestFunc func()

// didPanic safely executes the provided function and returns info about any panic it might have encountered.
func didPanic(f PanicTestFunc) (panicked bool, message interface{}, stack string) {
	panicked = true

	defer func() {
		message = recover()
		if panicked {
			stack = string(debug.Stack())
		}
	}()

	f()
	panicked = false

	return
}

// AssertPanicContents asserts that, if contains is empty, the provided func does not panic
// Otherwise, asserts that the func panics and that its panic message contains each of the provided strings.
//
// Returns true if it's all good, false if an assertion fails.
func AssertPanicContents(t TB, f PanicTestFunc, contains []string, msgAndArgs ...interface{}) bool {
	t.Helper()

	funcDidPanic, panicValue, panickedStack := didPanic(f)
	panicMsg := fmt.Sprintf("%v", panicValue)

	if len(contains) == 0 {
		if !funcDidPanic {
			return true
		}
		msg := fmt.Sprintf("func %#v should not panic, but did.", f)
		msg += fmt.Sprintf("\n\tPanic message:\t%q", panicMsg)
		msg += fmt.Sprintf("\n\t  Panic value:\t%#v", panicValue)
		msg += fmt.Sprintf("\n\t  Panic stack:\t%s", panickedStack)
		return assert.Fail(t, msg, msgAndArgs...)
	}

	if !funcDidPanic {
		msg := fmt.Sprintf("func %#v should panic, but did not.", f)
		for _, exp := range contains {
			msg += fmt.Sprintf("\n\tExpected to contain:\t%q", exp)
		}
		return assert.Fail(t, msg, msgAndArgs...)
	}

	var missing []string
	for _, exp := range contains {
		if !strings.Contains(panicMsg, exp) {
			missing = append(missing, exp)
		}
	}

	if len(missing) == 0 {
		return true
	}

	msg := fmt.Sprintf("func %#v panic message incorrect.", f)
	msg += fmt.Sprintf("\n\t   Panic message:\t%q", panicMsg)
	for _, exp := range missing {
		msg += fmt.Sprintf("\n\tDoes not contain:\t%q", exp)
	}
	msg += fmt.Sprintf("\n\tPanic value:\t%#v", panicValue)
	msg += fmt.Sprintf("\n\tPanic stack:\t%s", panickedStack)
	return assert.Fail(t, msg, msgAndArgs...)
}

// RequirePanicContents asserts that, if contains is empty, the provided func does not panic
// Otherwise, asserts that the func panics and that its panic message contains each of the provided strings.
//
// Returns if it's all good, halts the test if the assertion failed.
func RequirePanicContents(t TB, f PanicTestFunc, contains []string, msgAndArgs ...interface{}) {
	t.Helper()
	if !AssertPanicContents(t, f, contains, msgAndArgs...) {
		t.FailNow()
	}
}

// AssertPanicContentsf asserts that, if contains is empty, the provided func does not panic
// Otherwise, asserts that the func panics and that its panic message contains each of the provided strings.
//
// Returns true if it's all good, false if an assertion fails.
func AssertPanicContentsf(t TB, f PanicTestFunc, contains []string, msg string, args ...interface{}) bool {
	return AssertPanicContents(t, f, contains, append([]interface{}{msg}, args...)...)
}

// RequirePanicContentsf asserts that, if contains is empty, the provided func does not panic
// Otherwise, asserts that the func panics and that its panic message contains each of the provided strings.
//
// Returns if it's all good, halts the test if the assertion failed.
func RequirePanicContentsf(t TB, f PanicTestFunc, contains []string, msg string, args ...interface{}) {
	RequirePanicContents(t, f, contains, append([]interface{}{msg}, args...)...)
}

// AssertPanicEquals asserts that, if a panic is expected, the provided func panics with the expected output string.
// And if a panic is not expected, asserts that the provided func does not panic.
//
// Returns true if it's all good, false if an assertion fails.
func AssertPanicEquals(t TB, f PanicTestFunc, expected string, msgAndArgs ...interface{}) bool {
	t.Helper()

	funcDidPanic, panicValue, panickedStack := didPanic(f)
	var panicMsg string
	if panicValue != nil {
		panicMsg = fmt.Sprintf("%v", panicValue)
	}

	if len(expected) == 0 {
		if !funcDidPanic {
			return true
		}
		msg := fmt.Sprintf("func %#v should not panic, but did.", f)
		msg += fmt.Sprintf("\n\tPanic message:\t%q", panicMsg)
		msg += fmt.Sprintf("\n\t  Panic value:\t%#v", panicValue)
		msg += fmt.Sprintf("\n\t  Panic stack:\t%s", panickedStack)
		return assert.Fail(t, msg, msgAndArgs...)
	}

	if panicMsg == expected {
		return true
	}
	msg := fmt.Sprintf("func %#v ", f)
	if len(panicMsg) == 0 {
		msg += "did not panic, but should."
	} else {
		msg += "panic message incorrect."
		msg += fmt.Sprintf("\n\t   Panic message:\t%q", panicMsg)
	}
	msg += fmt.Sprintf("\n\tExpected message:\t%q", expected)
	if panicValue != nil {
		msg += fmt.Sprintf("\n\tPanic value:\t%#v", panicValue)
	}
	if len(panickedStack) > 0 {
		msg += fmt.Sprintf("\n\tPanic stack:\t%s", panickedStack)
	}
	return assert.Fail(t, msg, msgAndArgs...)
}

// RequirePanicEquals asserts that, if a panic is expected, the provided func panics with the expected output string.
// And if a panic is not expected, asserts that the provided func does not panic.
//
// Returns if it's all good, halts the test if the assertion failed.
func RequirePanicEquals(t TB, f PanicTestFunc, expected string, msgAndArgs ...interface{}) {
	t.Helper()
	if !AssertPanicEquals(t, f, expected, msgAndArgs...) {
		t.FailNow()
	}
}

// AssertPanicEqualsf asserts that, if a panic is expected, the provided func panics with the expected output string.
// And if a panic is not expected, asserts that the provided func does not panic.
//
// Returns true if it's all good, false if an assertion fails.
func AssertPanicEqualsf(t TB, f PanicTestFunc, expected string, msg string, args ...interface{}) bool {
	return AssertPanicEquals(t, f, expected, append([]interface{}{msg}, args...)...)
}

// RequirePanicEqualsf asserts that, if a panic is expected, the provided func panics with the expected output string.
// And if a panic is not expected, asserts that the provided func does not panic.
//
// Returns if it's all good, halts the test if the assertion failed.
func RequirePanicEqualsf(t TB, f PanicTestFunc, expected string, msg string, args ...interface{}) {
	RequirePanicEquals(t, f, expected, append([]interface{}{msg}, args...)...)
}

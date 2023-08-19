package testutil

import (
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/stretchr/testify/assert"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TB is a paired down version of the testing.TB interface with just the stuff needed in here.
//
// I'm kind of annoyed that I needed to do this, but it's the only way I could figure out to
// have unit tests that would check failure conditions, but not fail the parent test.
type TB interface {
	Helper()
	Errorf(format string, args ...any)
	Logf(format string, args ...any)
	FailNow()
}

// AssertErrorContents asserts that the provided error is as expected.
// If contains is empty, it asserts there is no error.
// Otherwise, it asserts that the error contains each of the entries in the contains slice.
//
// Returns true if it's all good, false if one or more assertions failed.
func AssertErrorContents(t TB, theError error, contains []string, msgAndArgs ...interface{}) bool {
	t.Helper()
	if len(contains) == 0 {
		return assert.NoError(t, theError, msgAndArgs...)
	}
	if !assert.Error(t, theError, msgAndArgs...) {
		// Also output what it was expected to have.
		if len(contains) == 1 {
			t.Logf("Error was expected to contain: %q", contains[0])
		} else {
			var sb strings.Builder
			for _, c := range contains {
				sb.WriteString(fmt.Sprintf("\n\t\t%q", c))
			}
			t.Logf("Error was expected to contain:%s", sb.String())
		}
		return false
	}

	errMsg := theError.Error()
	var missing []string
	for _, expInErr := range contains {
		if !strings.Contains(errMsg, expInErr) {
			missing = append(missing, expInErr)
		}
	}
	if len(missing) > 0 {
		var failureMsg strings.Builder
		failureMsg.WriteString(fmt.Sprintf("%d expected substring(s) missing from error.", len(missing)))
		failureMsg.WriteString(fmt.Sprintf("\nActual:\t%q", errMsg))
		failureMsg.WriteString("\nMissing:")
		for i, m := range missing {
			failureMsg.WriteString(fmt.Sprintf("\n\t%d: %q", i+1, m))
		}
		return assert.Fail(t, failureMsg.String(), msgAndArgs...)
	}
	return true
}

// RequireErrorContents asserts that the provided error is as expected.
// If contains is empty, it asserts there is no error.
// Otherwise, it asserts that the error contains each of the entries in the contains slice.
//
// Returns if it's all good, halts the test if one or more assertions failed.
func RequireErrorContents(t TB, theError error, contains []string, msgAndArgs ...interface{}) {
	t.Helper()
	if !AssertErrorContents(t, theError, contains, msgAndArgs...) {
		t.FailNow()
	}
}

// AssertErrorContentsf asserts that the provided error is as expected.
// If contains is empty, it asserts there is no error.
// Otherwise, it asserts that the error contains each of the entries in the contains slice.
//
// Returns true if it's all good, false if one or more assertions failed.
func AssertErrorContentsf(t TB, theError error, contains []string, msg string, args ...interface{}) bool {
	return AssertErrorContents(t, theError, contains, append([]interface{}{msg}, args...)...)
}

// RequireErrorContentsf asserts that the provided error is as expected.
// If contains is empty, it asserts there is no error.
// Otherwise, it asserts that the error contains each of the entries in the contains slice.
//
// Returns if it's all good, halts the test if one or more assertions failed.
func RequireErrorContentsf(t TB, theError error, contains []string, msg string, args ...interface{}) {
	RequireErrorContents(t, theError, contains, append([]interface{}{msg}, args...)...)
}

// AssertErrorValue asserts that, if an error is expected, theError equals the expected, otherwise asserts that
// there wasn't an error.
//
// Returns true if it's all good, false if the assertion failed.
func AssertErrorValue(t TB, theError error, expected string, msgAndArgs ...interface{}) bool {
	t.Helper()
	if len(expected) > 0 {
		return assert.EqualError(t, theError, expected, msgAndArgs...)
	}
	return assert.NoError(t, theError, msgAndArgs...)
}

// RequireErrorValue asserts that, if an error is expected, theError equals the expected, otherwise asserts that
// there wasn't an error.
//
// Returns if it's all good, halts the test if the assertion failed.
func RequireErrorValue(t TB, theError error, expected string, msgAndArgs ...interface{}) {
	t.Helper()
	if !AssertErrorValue(t, theError, expected, msgAndArgs...) {
		t.FailNow()
	}
}

// AssertErrorValuef asserts that, if an error is expected, theError equals the expected, otherwise asserts that
// there wasn't an error.
//
// Returns true if it's all good, false if the assertion failed.
func AssertErrorValuef(t TB, theError error, expected string, msg string, args ...interface{}) bool {
	return AssertErrorValue(t, theError, expected, append([]interface{}{msg}, args...)...)
}

// RequireErrorValuef asserts that, if an error is expected, theError equals the expected, otherwise asserts that
// there wasn't an error.
//
// Returns if it's all good, halts the test if the assertion failed.
func RequireErrorValuef(t TB, theError error, expected string, msg string, args ...interface{}) {
	RequireErrorValue(t, theError, expected, append([]interface{}{msg}, args...)...)
}

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

// ErrorTestFunc is a type declaration for a function that will be tested for an error.
type ErrorTestFunc func() error

// AssertNotPanicsNoError asserts that the code inside the provided function does not panic
// and that it does not return an error.
//
// Returns true if it neither panics nor errors, returns false otherwise.
func AssertNotPanicsNoError(t TB, f ErrorTestFunc, msgAndArgs ...interface{}) bool {
	t.Helper()
	var err error
	if !assert.NotPanics(t, func() { err = f() }, msgAndArgs...) {
		return false
	}
	return assert.NoError(t, err, msgAndArgs...)
}

// RequireNotPanicsNoError asserts that the code inside the provided function does not panic
// and that it does not return an error.
//
// Returns if it neither panics nor errors, otherwise it halts the test.
func RequireNotPanicsNoError(t TB, f ErrorTestFunc, msgAndArgs ...interface{}) {
	t.Helper()
	if !AssertNotPanicsNoError(t, f, msgAndArgs...) {
		t.FailNow()
	}
}

// AssertNotPanicsNoErrorf asserts that the code inside the provided function does not panic
// and that it does not return an error.
//
// Returns true if it neither panics nor errors, returns false otherwise.
func AssertNotPanicsNoErrorf(t TB, f ErrorTestFunc, msg string, args ...interface{}) bool {
	return AssertNotPanicsNoError(t, f, append([]interface{}{msg}, args...)...)
}

// RequireNotPanicsNoErrorf asserts that the code inside the provided function does not panic
// and that it does not return an error.
//
// Returns if it neither panics nor errors, otherwise it halts the test.
func RequireNotPanicsNoErrorf(t TB, f ErrorTestFunc, msg string, args ...interface{}) {
	RequireNotPanicsNoError(t, f, append([]interface{}{msg}, args...)...)
}

// prependToEach prepends the provided prefix to each of the provide lines.
func prependToEach(prefix string, lines []string) []string {
	for i, line := range lines {
		lines[i] = prefix + line
	}
	return lines
}

// eventsToStrings converts events to strings representing the events, one line per attribute.
func eventsToStrings(events sdk.Events) []string {
	var rv []string
	for i, event := range events {
		rv = append(rv, prependToEach(fmt.Sprintf("[%d]", i), eventToStrings(event))...)
	}
	return rv
}

// eventToStrings converts a single event to strings, one string per attribute.
func eventToStrings(event sdk.Event) []string {
	return prependToEach(event.Type, attrsToStrings(event.Attributes))
}

// attrsToStrings creates and returns a string for each attribute.
func attrsToStrings(attrs []abci.EventAttribute) []string {
	rv := make([]string, len(attrs))
	for i, attr := range attrs {
		rv[i] = fmt.Sprintf("[%d]: %q = %q", i, string(attr.Key), string(attr.Value))
		if attr.Index {
			rv[i] += " (indexed)"
		}
	}
	return rv
}

// AssertEqualEvents asserts that the expected events equal the actual events.
//
// Returns success (true = they're equal, false = they're different).
func AssertEqualEvents(t TB, expected, actual sdk.Events, msgAndArgs ...interface{}) bool {
	t.Helper()
	// This converts them to strings for the comparison so that the failure output is significantly easier to read and understand.
	expectedStrs := eventsToStrings(expected)
	actualStrs := eventsToStrings(actual)
	return assert.Equal(t, expectedStrs, actualStrs, msgAndArgs...)
}

// RequireEqualEvents asserts that the expected events equal the actual events.
//
// Returns if they're equal, halts tests if not.
func RequireEqualEvents(t TB, expected, actual sdk.Events, msgAndArgs ...interface{}) {
	if !AssertEqualEvents(t, expected, actual, msgAndArgs...) {
		t.FailNow()
	}
}

// AssertEqualEventsf asserts that the expected events equal the actual events.
//
// Returns success (true = they're equal, false = they're different).
func AssertEqualEventsf(t TB, expected, actual sdk.Events, msg string, args ...interface{}) bool {
	return AssertEqualEvents(t, expected, actual, append([]interface{}{msg}, args...)...)
}

// RequireEqualEventsf asserts that the expected events equal the actual events.
//
// Returns if they're equal, halts tests if not.
func RequireEqualEventsf(t TB, expected, actual sdk.Events, msg string, args ...interface{}) {
	RequireEqualEvents(t, expected, actual, append([]interface{}{msg}, args...)...)
}

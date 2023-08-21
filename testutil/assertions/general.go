package assertions

import "github.com/stretchr/testify/assert"

// TB is a paired down version of the testing.TB interface with just the stuff needed in here.
//
// I'm kind of annoyed that I needed to do this, but it's the only way I could figure out to
// have unit tests on these assertions that would check failure conditions, but not fail the parent test.
type TB interface {
	Helper()
	Errorf(format string, args ...any)
	Logf(format string, args ...any)
	FailNow()
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

package assertions

import (
	"fmt"
	"strings"

	"github.com/stretchr/testify/assert"
)

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

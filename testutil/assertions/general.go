package assertions

import (
	"fmt"
	"strings"

	"github.com/stretchr/testify/assert"

	"github.com/cosmos/cosmos-sdk/types/query"
)

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

// prefixMsgAndArgs will add the provided prefix to the front of the first element of msgAndArgs.
// If no msgAndArgs are provided, a slice with just the prefix is returned.
// If the first element of msgAndArgs isn't a string, msgAndArgs is returned unchanged.
func prefixMsgAndArgs(pre string, msgAndArgs []interface{}) []interface{} {
	if len(pre) == 0 {
		return msgAndArgs
	}
	if len(msgAndArgs) == 0 {
		return []interface{}{pre}
	}

	v0, ok := msgAndArgs[0].(string)
	if !ok {
		return msgAndArgs
	}

	rv := make([]interface{}, len(msgAndArgs))
	copy(rv, msgAndArgs)
	rv[0] = prefixMsg(pre, v0)
	return rv
}

// prefixMsg adds a prefix to a message. If either are empty, the other is returned, otherwise it's "<pre>: <msg>".
func prefixMsg(pre, msg string) string {
	if len(pre) == 0 {
		return msg
	}
	if len(msg) == 0 {
		return pre
	}
	return pre + ": " + msg
}

// AssertEqualPageResponses asserts that the two provided page responses are equal.
//
// Returns success (true = they're equal, false = they're different).
func AssertEqualPageResponses(t TB, exp, act *query.PageResponse, msgAndArgs ...interface{}) bool {
	t.Helper()
	if exp == nil {
		return assert.Nil(t, act, prefixMsgAndArgs("PageResponse", msgAndArgs)...)
	}
	if !assert.NotNil(t, act, prefixMsgAndArgs("PageResponse", msgAndArgs)...) {
		return false
	}

	ok := true

	expKey := strings.Trim(fmt.Sprintf("%q", exp.NextKey), `"`)
	actKey := strings.Trim(fmt.Sprintf("%q", act.NextKey), `"`)
	ok = assert.Equal(t, expKey, actKey, prefixMsgAndArgs("PageResponse.NextKey", msgAndArgs)...) && ok

	expTotal := fmt.Sprintf("%d", exp.Total)
	actTotal := fmt.Sprintf("%d", act.Total)
	ok = assert.Equal(t, expTotal, actTotal, prefixMsgAndArgs("PageResponse.Total", msgAndArgs)...) && ok

	return ok && assert.Equal(t, exp, act, prefixMsgAndArgs("PageResponse", msgAndArgs)...)
}

// RequireEqualPageResponses requires that the two provided page responses are equal.
//
// Returns if they're equal, halts tests if not.
func RequireEqualPageResponses(t TB, exp, act *query.PageResponse, msgAndArgs ...interface{}) {
	t.Helper()
	if !AssertEqualPageResponses(t, exp, act, msgAndArgs...) {
		t.FailNow()
	}
}

// AssertEqualPageResponsesf asserts that the two provided page responses are equal.
//
// Returns success (true = they're equal, false = they're different).
func AssertEqualPageResponsesf(t TB, exp, act *query.PageResponse, msg string, args ...interface{}) bool {
	t.Helper()
	return AssertEqualPageResponses(t, exp, act, append([]interface{}{msg}, args...)...)
}

// RequireEqualPageResponsesf requires that the two provided page responses are equal.
//
// Returns if they're equal, halts tests if not.
func RequireEqualPageResponsesf(t TB, exp, act *query.PageResponse, msg string, args ...interface{}) {
	t.Helper()
	RequireEqualPageResponses(t, exp, act, append([]interface{}{msg}, args...)...)
}

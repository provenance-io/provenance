package assertions

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
)

const (
	errorLead = "\n\tError:      \t"
	blankLead = "\n\t            \t"
)

// mockTB is a TB to use for testing since there's no way to create
// a divorced T to run tests on things that take in a T.
// It keeps track of output and failure status.
type mockTB struct {
	mu       sync.RWMutex
	panic    any
	isFailed bool
	output   string
}

var _ TB = (*mockTB)(nil)

func (t *mockTB) Helper() {}
func (t *mockTB) Errorf(format string, args ...any) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.isFailed = true
	t.output += fmt.Sprintf(format, args...)
}
func (t *mockTB) Logf(format string, args ...any) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.output += fmt.Sprintf(format, args...)
}
func (t *mockTB) FailNow() {
	t.mu.Lock()
	defer t.mu.Unlock()
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
			if r := recover(); r != nil {
				rv.mu.Lock()
				defer rv.mu.Unlock()
				rv.panic = r
			}
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
			assert.Contains(t, tb.output, exp, "%s output, looking for %q", funcName, exp)
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
			assert.Contains(t, tb.output, exp, "%s output, looking for %q", funcName, exp)
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

func TestPrefixMsgAndArgs(t *testing.T) {
	tests := []struct {
		name       string
		pre        string
		msgAndArgs []interface{}
		exp        []interface{}
	}{
		{
			name:       "empty pre: nil args",
			pre:        "",
			msgAndArgs: nil,
			exp:        nil,
		},
		{
			name:       "empty pre: empty args",
			pre:        "",
			msgAndArgs: []interface{}{},
			exp:        []interface{}{},
		},
		{
			name:       "empty pre: one arg",
			pre:        "",
			msgAndArgs: []interface{}{"one"},
			exp:        []interface{}{"one"},
		},
		{
			name:       "empty pre: three args",
			pre:        "",
			msgAndArgs: []interface{}{"one", "two", "three"},
			exp:        []interface{}{"one", "two", "three"},
		},
		{
			name:       "nil msgAndArgs",
			pre:        "addme",
			msgAndArgs: nil,
			exp:        []interface{}{"addme"},
		},
		{
			name:       "empty msgAndArgs",
			pre:        "addme",
			msgAndArgs: []interface{}{},
			exp:        []interface{}{"addme"},
		},
		{
			name:       "one msgAndArgs: string",
			pre:        "plusone",
			msgAndArgs: []interface{}{"arg0"},
			exp:        []interface{}{"plusone: arg0"},
		},
		{
			name:       "one msgAndArgs: not string",
			pre:        "plusone",
			msgAndArgs: []interface{}{77},
			exp:        []interface{}{77},
		},
		{
			name:       "three args: first is string",
			pre:        "label",
			msgAndArgs: []interface{}{"thingy with %d and %q", 99, "bananas"},
			exp:        []interface{}{"label: thingy with %d and %q", 99, "bananas"},
		},
		{
			name:       "three args: first is not string",
			pre:        "label",
			msgAndArgs: []interface{}{4, "thingy with %d and %q", 99, "bananas"},
			exp:        []interface{}{4, "thingy with %d and %q", 99, "bananas"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act []interface{}
			testFunc := func() {
				act = prefixMsgAndArgs(tc.pre, tc.msgAndArgs)
			}
			require.NotPanics(t, testFunc, "prefixMsgAndArgs(%q, %#v)", tc.pre, tc.msgAndArgs)
			assert.Equal(t, tc.exp, act, "prefixMsgAndArgs(%q, %#v) result", tc.pre, tc.msgAndArgs)
		})
	}
}

func TestPrefixMsg(t *testing.T) {
	tests := []struct {
		name string
		pre  string
		msg  string
		exp  string
	}{
		{name: "both are empty", pre: "", msg: "", exp: ""},
		{name: "only empty pre", pre: "", msg: "the message", exp: "the message"},
		{name: "only empty msg", pre: "daprefix", msg: "", exp: "daprefix"},
		{name: "both have value", pre: "label", msg: "content", exp: "label: content"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act string
			testFunc := func() {
				act = prefixMsg(tc.pre, tc.msg)
			}
			require.NotPanics(t, testFunc, "prefixMsg(%q, %q)", tc.pre, tc.msg)
			assert.Equal(t, tc.exp, act, "prefixMsg(%q, %q) result", tc.pre, tc.msg)
		})
	}
}

// equalPageResponsesTestCase is a test case for the [Assert|Require]EqualPageResponses[f]? functions.
type equalPageResponsesTestCase struct {
	name      string
	exp       *query.PageResponse
	act       *query.PageResponse
	errLabel  string
	expOutput []string
}

// getEqualPageResponsesTestCases returns all the tests cases for the [Assert|Require]EqualPageResponses[f]? functions.
func getEqualPageResponsesTestCases() []*equalPageResponsesTestCase {
	return []*equalPageResponsesTestCase{
		{
			name:      "both nil",
			exp:       nil,
			act:       nil,
			expOutput: nil,
		},
		{
			name:     "exp nil, act not",
			exp:      nil,
			act:      &query.PageResponse{},
			errLabel: "PageResponse",
			expOutput: []string{
				errorLead + "Expected nil, but got: &query.PageResponse{NextKey:[]uint8(nil), Total:0x0}",
			},
		},
		{
			name:     "act nil, exp not",
			exp:      &query.PageResponse{},
			act:      nil,
			errLabel: "PageResponse",
			expOutput: []string{
				errorLead + "Expected value not to be nil.",
			},
		},
		{
			name:     "different next keys, same totals",
			exp:      &query.PageResponse{NextKey: []byte("gofor1"), Total: 10},
			act:      &query.PageResponse{NextKey: []byte("gofor2"), Total: 10},
			errLabel: "PageResponse.NextKey",
			expOutput: []string{
				errorLead + "Not equal: ",
				blankLead + "expected: \"gofor1\"",
				blankLead + "actual  : \"gofor2\"",
			},
		},
		{
			name:     "same next keys, different totals",
			exp:      &query.PageResponse{NextKey: []byte("gofor1"), Total: 10},
			act:      &query.PageResponse{NextKey: []byte("gofor1"), Total: 11},
			errLabel: "PageResponse.Total",
			expOutput: []string{
				errorLead + "Not equal: ",
				blankLead + "expected: \"10\"",
				blankLead + "actual  : \"11\"",
			},
		},
		{
			name:     "different next keys and totals",
			exp:      &query.PageResponse{NextKey: []byte{0x01, 0x70, 0x72, 0x6F, 0x76, 0x07, 0x31}, Total: 44},
			act:      &query.PageResponse{NextKey: []byte{0x01, 0x50, 0x52, 0x4F, 0x56, 0x07, 0x31}, Total: 74},
			errLabel: "PageResponse.NextKey",
			expOutput: []string{
				errorLead + "Not equal: ",
				blankLead + `expected: "\\x01prov\\a1"`,
				blankLead + `actual  : "\\x01PROV\\a1"`,
				blankLead + "expected: \"44\"",
				blankLead + "actual  : \"74\"",
			},
		},
	}
}

func TestAssertEqualPageResponses(t *testing.T) {
	funcName := "AssertEqualPageResponses"
	for _, tc := range getEqualPageResponsesTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			msg := "a message with %d args: %v %v %v"
			args := []interface{}{3, "arg 1", "arg 2", "arg 3"}
			msgAndArgs := append([]interface{}{msg}, args...)
			expMsgAndArgs := "Messages:   \t" + tc.errLabel + ": " + fmt.Sprintf(msg, args...)

			var success bool
			testFunc := func(testTB TB) {
				success = AssertEqualPageResponses(testTB, tc.exp, tc.act, msgAndArgs...)
			}
			tb := mockRun(t, testFunc)

			assertMockRunAssertResult(t, funcName, tb, success, tc.expOutput, expMsgAndArgs)
		})
	}
}

func TestRequireEqualPageResponses(t *testing.T) {
	funcName := "RequireEqualPageResponses"
	for _, tc := range getEqualPageResponsesTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			msg := "a message with %d args: %v %v %v"
			args := []interface{}{3, "arg 1", "arg 2", "arg 3"}
			msgAndArgs := append([]interface{}{msg}, args...)
			expMsgAndArgs := "Messages:   \t" + tc.errLabel + ": " + fmt.Sprintf(msg, args...)

			exited := true
			testFunc := func(testTB TB) {
				RequireEqualPageResponses(testTB, tc.exp, tc.act, msgAndArgs...)
				exited = false
			}
			tb := mockRun(t, testFunc)

			assertMockRunRequireResult(t, funcName, tb, exited, tc.expOutput, expMsgAndArgs)
		})
	}
}

func TestAssertEqualPageResponsesf(t *testing.T) {
	funcName := "AssertEqualPageResponsesf"
	for _, tc := range getEqualPageResponsesTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			msg := "a message with %d args: %v %v %v"
			args := []interface{}{3, "arg 1", "arg 2", "arg 3"}
			expMsgAndArgs := "Messages:   \t" + tc.errLabel + ": " + fmt.Sprintf(msg, args...)

			var success bool
			testFunc := func(testTB TB) {
				success = AssertEqualPageResponsesf(testTB, tc.exp, tc.act, msg, args...)
			}
			tb := mockRun(t, testFunc)

			assertMockRunAssertResult(t, funcName, tb, success, tc.expOutput, expMsgAndArgs)
		})
	}
}

func TestRequireEqualPageResponsesf(t *testing.T) {
	funcName := "RequireEqualPageResponsesf"
	for _, tc := range getEqualPageResponsesTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			msg := "a message with %d args: %v %v %v"
			args := []interface{}{3, "arg 1", "arg 2", "arg 3"}
			expMsgAndArgs := "Messages:   \t" + tc.errLabel + ": " + fmt.Sprintf(msg, args...)

			exited := true
			testFunc := func(testTB TB) {
				RequireEqualPageResponsesf(testTB, tc.exp, tc.act, msg, args...)
				exited = false
			}
			tb := mockRun(t, testFunc)

			assertMockRunRequireResult(t, funcName, tb, exited, tc.expOutput, expMsgAndArgs)
		})
	}
}

package assertions

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"

	abci "github.com/tendermint/tendermint/abci/types"
)

func TestPrependToEach(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		lines  []string
		exp    []string
	}{
		{
			name:   "nil lines",
			prefix: "ignored",
			lines:  nil,
			exp:    nil,
		},
		{
			name:   "empty lines",
			prefix: "ignored",
			lines:  []string{},
			exp:    []string{},
		},
		{
			name:   "one line no prefix",
			prefix: "",
			lines:  []string{"line one"},
			exp:    []string{"line one"},
		},
		{
			name:   "one line with prefix",
			prefix: "new beginning",
			lines:  []string{"line one"},
			exp:    []string{"new beginningline one"},
		},
		{
			name:   "two lines no prefix",
			prefix: "",
			lines:  []string{"one", "two"},
			exp:    []string{"one", "two"},
		},
		{
			name:   "two lines with prefix",
			prefix: "indent: ",
			lines:  []string{"one", "two"},
			exp:    []string{"indent: one", "indent: two"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := PrependToEach(tc.prefix, tc.lines)
			assert.Equal(t, tc.exp, actual, "PrependToEach")
		})
	}
}

func TestEventsToStrings(t *testing.T) {
	attr := func(key, value string) abci.EventAttribute {
		return abci.EventAttribute{
			Key:   []byte(key),
			Value: []byte(fmt.Sprintf("%q", value)),
		}
	}
	addrAdd := sdk.AccAddress("address_add_event___").String()
	coinsAdd := "97acorn,12banana"
	reason := "just some test reason"
	eventAdd := sdk.Event{
		Type: "provenance.hold.v1.EventHoldAdded",
		Attributes: []abci.EventAttribute{
			attr("address", addrAdd),
			attr("amount", coinsAdd),
			attr("reason", reason),
		},
	}
	eventAdd.Attributes[0].Index = true

	addrRel := sdk.AccAddress("address_rel_event___").String()
	coinsRel := "13cucumber,81dill"
	eventRel := sdk.Event{
		Type: "provenance.hold.v1.EventHoldReleased",
		Attributes: []abci.EventAttribute{
			attr("address", addrRel),
			attr("amount", coinsRel),
		},
	}

	tests := []struct {
		name   string
		events sdk.Events
		exp    []string
	}{
		{
			name:   "nil events",
			events: nil,
			exp:    nil,
		},
		{
			name:   "empty events",
			events: sdk.Events{},
			exp:    nil,
		},
		{
			name:   "one event",
			events: sdk.Events{eventRel},
			exp: []string{
				fmt.Sprintf("[0]provenance.hold.v1.EventHoldReleased[0]: \"address\" = \"\\\"%s\\\"\"", addrRel),
				fmt.Sprintf("[0]provenance.hold.v1.EventHoldReleased[1]: \"amount\" = \"\\\"%s\\\"\"", coinsRel),
			},
		},
		{
			name:   "three events",
			events: sdk.Events{eventAdd, sdk.Event{Type: "weird.entry"}, eventRel},
			exp: []string{
				fmt.Sprintf("[0]provenance.hold.v1.EventHoldAdded[0]: \"address\" = \"\\\"%s\\\"\" (indexed)", addrAdd),
				fmt.Sprintf("[0]provenance.hold.v1.EventHoldAdded[1]: \"amount\" = \"\\\"%s\\\"\"", coinsAdd),
				fmt.Sprintf("[0]provenance.hold.v1.EventHoldAdded[2]: \"reason\" = \"\\\"%s\\\"\"", reason),
				"[1]weird.entry: (no attributes)",
				fmt.Sprintf("[2]provenance.hold.v1.EventHoldReleased[0]: \"address\" = \"\\\"%s\\\"\"", addrRel),
				fmt.Sprintf("[2]provenance.hold.v1.EventHoldReleased[1]: \"amount\" = \"\\\"%s\\\"\"", coinsRel),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := EventsToStrings(tc.events)
			assert.Equal(t, tc.exp, actual, "EventsToStrings")
		})
	}
}

func TestEventToStrings(t *testing.T) {
	attr := func(key, value string) abci.EventAttribute {
		return abci.EventAttribute{
			Key:   []byte(key),
			Value: []byte(fmt.Sprintf("%q", value)),
		}
	}

	addrAdd := sdk.AccAddress("address_add_event___").String()
	coinsAdd := "97acorn,12banana"
	reason := "just some test reason"
	eventAdd := sdk.Event{
		Type: "provenance.hold.v1.EventHoldAdded",
		Attributes: []abci.EventAttribute{
			attr("address", addrAdd),
			attr("amount", coinsAdd),
			attr("reason", reason),
		},
	}
	eventAdd.Attributes[0].Index = true

	tests := []struct {
		name  string
		event sdk.Event
		exp   []string
	}{
		{
			name:  "nil attributes",
			event: sdk.Event{Type: "some.type", Attributes: nil},
			exp:   []string{"some.type: (no attributes)"},
		},
		{
			name:  "empty attributes",
			event: sdk.Event{Type: "some.type", Attributes: []abci.EventAttribute{}},
			exp:   []string{"some.type: (no attributes)"},
		},
		{
			name:  "one attribute",
			event: sdk.Event{Type: "another.type", Attributes: []abci.EventAttribute{attr("key", "value")}},
			exp: []string{
				"another.type[0]: \"key\" = \"\\\"value\\\"\"",
			},
		},
		{
			name:  "three attributes",
			event: eventAdd,
			exp: []string{
				fmt.Sprintf("provenance.hold.v1.EventHoldAdded[0]: \"address\" = \"\\\"%s\\\"\" (indexed)", addrAdd),
				fmt.Sprintf("provenance.hold.v1.EventHoldAdded[1]: \"amount\" = \"\\\"%s\\\"\"", coinsAdd),
				fmt.Sprintf("provenance.hold.v1.EventHoldAdded[2]: \"reason\" = \"\\\"%s\\\"\"", reason),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := EventToStrings(tc.event)
			assert.Equal(t, tc.exp, actual, "EventToStrings")
		})
	}
}

func TestAttrsToStrings(t *testing.T) {
	attr := func(key, value string, index bool) abci.EventAttribute {
		return abci.EventAttribute{
			Key:   []byte(key),
			Value: []byte(fmt.Sprintf("%q", value)),
			Index: index,
		}
	}

	tests := []struct {
		name  string
		attrs []abci.EventAttribute
		exp   []string
	}{
		{
			name:  "nil attributes",
			attrs: nil,
			exp:   nil,
		},
		{
			name:  "empty attributes",
			attrs: []abci.EventAttribute{},
			exp:   nil,
		},
		{
			name:  "one unindexed attribute",
			attrs: []abci.EventAttribute{attr("somekey", "somevalue", false)},
			exp:   []string{"[0]: \"somekey\" = \"\\\"somevalue\\\"\""},
		},
		{
			name:  "one indexed attribute",
			attrs: []abci.EventAttribute{attr("anotherkey", "anothervalue", true)},
			exp:   []string{"[0]: \"anotherkey\" = \"\\\"anothervalue\\\"\" (indexed)"},
		},
		{
			name:  "value with a double quote char in it",
			attrs: []abci.EventAttribute{attr("weird", "this has a \" in it", false)},
			exp:   []string{"[0]: \"weird\" = \"\\\"this has a \\\\\\\" in it\\\"\""},
		},
		{
			name: "three attributes",
			attrs: []abci.EventAttribute{
				attr("attr0", "value0", false),
				attr("attr1", "value1", true),
				attr("attr2", "this is the third value", false),
			},
			exp: []string{
				"[0]: \"attr0\" = \"\\\"value0\\\"\"",
				"[1]: \"attr1\" = \"\\\"value1\\\"\" (indexed)",
				"[2]: \"attr2\" = \"\\\"this is the third value\\\"\"",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := AttrsToStrings(tc.attrs)
			assert.Equal(t, tc.exp, actual, "AttrsToStrings")
		})
	}
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

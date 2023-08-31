package assertions

import (
	"fmt"

	"github.com/stretchr/testify/assert"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// PrependToEach prepends the provided prefix to each of the provide lines.
func PrependToEach(prefix string, lines []string) []string {
	for i, line := range lines {
		lines[i] = prefix + line
	}
	return lines
}

// EventsToStrings converts events to strings representing the events, one line per attribute.
func EventsToStrings(events sdk.Events) []string {
	var rv []string
	for i, event := range events {
		rv = append(rv, PrependToEach(fmt.Sprintf("[%d]", i), EventToStrings(event))...)
	}
	return rv
}

// EventToStrings converts a single event to strings, one string per attribute.
func EventToStrings(event sdk.Event) []string {
	if len(event.Attributes) == 0 {
		return []string{fmt.Sprintf("%s: (no attributes)", event.Type)}
	}
	return PrependToEach(event.Type, AttrsToStrings(event.Attributes))
}

// AttrsToStrings creates and returns a string for each attribute.
func AttrsToStrings(attrs []abci.EventAttribute) []string {
	if len(attrs) == 0 {
		return nil
	}
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
	expectedStrs := EventsToStrings(expected)
	actualStrs := EventsToStrings(actual)
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

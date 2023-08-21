package assertions

import (
	"fmt"

	"github.com/stretchr/testify/assert"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

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

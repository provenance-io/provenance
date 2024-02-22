package assertions

import (
	"fmt"
	"strings"

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

// EventToString converts a single event to a one-line string.
func EventToString(event sdk.Event) string {
	attrs := "(no attributes)"
	if len(event.Attributes) > 0 {
		attrs = strings.Join(AttrsToStrings(event.Attributes), ", ")
	}
	return fmt.Sprintf("%s: %s", event.Type, attrs)
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

// AssertEventsContains asserts that each of the provided expected events is contained in the provided actual events.
//
// Returns success (true = they're all there, false = one or more is missing).
func AssertEventsContains(t TB, expected, actual sdk.Events, msgAndArgs ...interface{}) bool {
	t.Helper()
	if len(expected) == 0 {
		return true
	}

	actStrs := make([]string, len(actual))
	for i, event := range actual {
		actStrs[i] = EventToString(event)
	}

	var notFound sdk.Events
	for _, expEvent := range expected {
		exp := EventToString(expEvent)
		var found bool
		for _, act := range actStrs {
			if exp == act {
				found = true
				break
			}
		}
		if !found {
			notFound = append(notFound, expEvent)
		}
	}

	if len(notFound) == 0 {
		return true
	}

	var failureMsg strings.Builder
	failureMsg.WriteString(fmt.Sprintf("%d (of %d) expected events missing from %d actual events", len(notFound), len(expected), len(actual)))
	failureMsg.WriteString("\nActual:")
	switch {
	case actual == nil:
		failureMsg.WriteString(" <nil>")
	case len(actual) == 0:
		failureMsg.WriteString(" sdk.Events{}")
	default:
		failureMsg.WriteByte('\n')
		failureMsg.WriteString(strings.Join(PrependToEach("\n\t", EventsToStrings(actual)), ""))
	}
	failureMsg.WriteString("\nMissing:")
	for i, event := range notFound {
		for _, line := range EventToStrings(event) {
			failureMsg.WriteString(fmt.Sprintf("\n\t%d: %s", i, line))
		}
	}
	return assert.Fail(t, failureMsg.String(), msgAndArgs...)
}

// RequireEventsContains asserts that each of the provided expected events is contained in the provided actual events.
//
// Returns if they're all there, halts tests if not.
func RequireEventsContains(t TB, expected, actual sdk.Events, msgAndArgs ...interface{}) {
	if !AssertEventsContains(t, expected, actual, msgAndArgs...) {
		t.FailNow()
	}
}

// AssertEventsContainsf asserts that each of the provided expected events is contained in the provided actual events.
//
// Returns success (true = they're all there, false = one or more is missing).
func AssertEventsContainsf(t TB, expected, actual sdk.Events, msg string, args ...interface{}) bool {
	return AssertEventsContains(t, expected, actual, append([]interface{}{msg}, args...)...)
}

// RequireEventsContainsf asserts that each of the provided expected events is contained in the provided actual events.
//
// Returns if they're all there, halts tests if not.
func RequireEventsContainsf(t TB, expected, actual sdk.Events, msg string, args ...interface{}) {
	RequireEventsContains(t, expected, actual, append([]interface{}{msg}, args...)...)
}

package testlog

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

// WriteSlice writes the contents of the provided slice to the test logs.
//
// The first line will have a header with the name and length.
// Then, there'll be one line for eac entry, each with the format "<name>[<index>] = <value>" (with the = lined up).
func WriteSlice[S ~[]E, E any](t testing.TB, name string, vals S) {
	t.Log(createSliceLogString(name, vals))
}

// WriteVariables writes the provided named variables to the test logs under the given header.
//
// The namesAndValues args must be provided in pairs, the first is the name, the second is the value.
// Name args must be a string. Value args can be anything, and do not need to all be the same type.
//
// The test fails immediately if an odd number of namesAndValues are provided or if any name args are not a string.
//
// E.g. LogNamedValues(t, "addresses", "addr1", addr1, "addr2", addr2)
//
// The first line will contain the provided header and a count of variables being logged.
// Then, there'll be one line for each variable, each with the format "<name> = <value>" (with the = lined up).
//
// See also: WriteSlice.
func WriteVariables(t testing.TB, header string, namesAndValues ...interface{}) {
	t.Log(newNamedValues(t, namesAndValues).GetLogString(header))
}

// WriteVariable writes the provided named variable to the test logs in the format "<name> = <value>".
func WriteVariable(t testing.TB, name string, value interface{}) {
	t.Logf("%s = %s", name, valueString(value))
}

// createSliceLogString creates a multi-line string of the given slice for the purposes of logging.
func createSliceLogString[S ~[]E, E any](name string, vals S) string {
	return namedValuesForSlice(name, vals).GetLogString(name)
}

// namedValue associates a name with a value.
type namedValue struct {
	Name  string
	Value interface{}
}

// newNamedValue creates a new namedValue.
func newNamedValue(name string, value interface{}) *namedValue {
	return &namedValue{Name: name, Value: value}
}

// namedValues is a slice of namedValue entries.
type namedValues []*namedValue

// newNamedValues creates a namedValues from the provided namesAndValues.
//
// The namesAndValues args must be provided in pairs, the first is the name, the second is the value.
// Name args must be a string. Value args can be anything, and do not need to all be the same type.
//
// The test fails immediately if an odd number of namesAndValues are provided or if any name args are not a string.
//
// E.g. newNamedValues(t, "addr1", addr1, "addr2", addr2)
func newNamedValues(t testing.TB, namesAndValues []interface{}) namedValues {
	if len(namesAndValues) == 0 {
		return nil
	}
	if len(namesAndValues)%2 != 0 {
		t.Fatalf("Odd number of name/value args provided.\n%s",
			createSliceLogString("namesAndValues", namesAndValues))
	}
	rv := make(namedValues, 0, len(namesAndValues)/2)
	for i := 0; i < len(namesAndValues); i += 2 {
		nameArg := namesAndValues[i]
		valueArg := namesAndValues[i+1]
		name, nameOK := nameArg.(string)
		if !nameOK {
			t.Fatalf("Invalid name/value arg pair: name arg has type %T, expected string.\n"+
				" (name)=args[%d]=%#v\n"+
				"(value)=args[%d]=%#v\n"+
				"%s",
				nameArg, i, nameArg, i+1, valueArg, createSliceLogString("namesAndValues", namesAndValues))
		}
		rv = append(rv, newNamedValue(name, valueArg))
	}
	return rv
}

// namedValuesForSlice creates a namedValues representing the provided slice and its values.
func namedValuesForSlice[S ~[]E, E any](name string, vals S) namedValues {
	rv := make(namedValues, len(vals))
	for i, val := range vals {
		rv[i] = newNamedValue(fmt.Sprintf("%s[%d]", name, i), val)
	}
	return rv
}

// GetLogString creates a multi-line string with the provided header and each of the entries in this namedValues.
func (s namedValues) GetLogString(header string) string {
	nameWidth := 0
	for _, entry := range s {
		if len(entry.Name) > nameWidth {
			nameWidth = len(entry.Name)
		}
	}
	lineFmt := "%" + strconv.Itoa(nameWidth) + "s = %s"

	lines := make([]string, len(s))
	for i, entry := range s {
		if entry == nil {
			lines[i] = "<nil>"
		} else {
			lines[i] = fmt.Sprintf(lineFmt, entry.Name, valueString(entry.Value))
		}
	}
	return fmt.Sprintf("%s (%d):\n%s", header, len(s), strings.Join(lines, "\n"))
}

// valueString creates a string of the given value.
func valueString(value interface{}) string {
	if value == nil {
		return "<nil>"
	}
	switch val := value.(type) {
	case string:
		return val
	case fmt.Stringer:
		return val.String()
	}
	return fmt.Sprintf("%#v", value)
}

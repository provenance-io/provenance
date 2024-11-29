package provutils

import (
	"errors"
	"fmt"
	"strings"
)

// FindMissing returns all elements of the required list that are not found in the entries list.
// Duplicate entries in required do not require duplicate entries in toCheck.
// E.g. FindMissing([a, b, a], [a]) => [b], and FindMissing([a, b, a], [b]) => [a, a].
//
// See also: FindMissingFunc.
func FindMissing[S ~[]E, E comparable](required, toCheck S) S {
	return FindMissingFunc(required, toCheck, func(r, c E) bool { return r == c })
}

// FindMissingFunc returns all entries in required where the equals function returns false for all entries of toCheck.
// Duplicate entries in required do not require duplicate entries in toCheck.
// E.g. FindMissingFunc([a, b, a], [a]) => [b], and FindMissingFunc([a, b, a], [b]) => [a, a].
//
// See also: FindMissing.
func FindMissingFunc[SR ~[]R, SC ~[]C, R any, C any](required SR, toCheck SC, equals func(R, C) bool) SR {
	var rv []R
reqLoop:
	for _, req := range required {
		for _, entry := range toCheck {
			if equals(req, entry) {
				continue reqLoop
			}
		}
		rv = append(rv, req)
	}
	return rv
}

// SliceString converts a slice to a string in the foramt "[val1,val2,...]" or "<nil>".
func SliceString[S ~[]E, E fmt.Stringer](vals S) string {
	if vals == nil {
		return "<nil>"
	}
	if len(vals) == 0 {
		return "[]"
	}
	strs := make([]string, len(vals))
	for i, val := range vals {
		strs[i] = val.String()
	}
	return "[" + strings.Join(strs, ",") + "]"
}

// ValidateSlice runs the provided validator on each of the vals and returns an error that combines all errors.
// An empty slice will return nil.
func ValidateSlice[S ~[]E, E any](vals S, validator func(E) error) error {
	if len(vals) == 0 {
		return nil
	}
	var errs []error
	for i, val := range vals {
		if err := validator(val); err != nil {
			errs = append(errs, fmt.Errorf("%d: %w", i, err))
		}
	}
	return errors.Join(errs...)
}

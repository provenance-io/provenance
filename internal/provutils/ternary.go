package provutils

// Ternary returns ifTrue if test is true, or ifFalse if test is false.
// It's similar to Ternary assignments in other languages that often
// have the syntax like this: value = test ? ifTrue : ifFalse;
func Ternary[V any](test bool, ifTrue V, ifFalse V) V {
	if test {
		return ifTrue
	}
	return ifFalse
}

// Pluralize returns ifOne if the provided vals has length 1, otherwise returns ifOther.
//
// E.g. Pluralize(parties, "party", "parties")
//
// If the only difference between ifOne and ifOther is an additional "s", consider using PluralEnding.
func Pluralize[S ~[]any](vals S, ifOne, ifOther string) string {
	if len(vals) == 1 {
		return ifOne
	}
	return ifOther
}

// PluralEnding returns an empty string if vals has length 1, otherwise returns "s".
//
// E.g. name := "cat" + PluralEnding(cats)
//
// If you need more than the addition of an "s", use Pluralize.
func PluralEnding[S ~[]E, E any](vals S) string {
	if len(vals) == 1 {
		return ""
	}
	return "s"
}

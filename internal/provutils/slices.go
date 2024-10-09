package provutils

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

package helpers

// This file houses functions that are in "golang.org/x/exp/maps" but not "maps", but that we want to use.
// If any show up in "maps", delete them from here and switch uses to the official ones.

// Keys returns the unordered keys of the given map.
// This is the same as the experimental maps.Keys function.
// As of writing, that isn't in the standard library version yet. Once it is, remove this and switch to that.
func Keys[M ~map[K]V, K comparable, V any](m M) []K {
	rv := make([]K, 0, len(m))
	for k := range m {
		rv = append(rv, k)
	}
	return rv
}

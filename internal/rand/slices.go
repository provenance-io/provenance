package rand

import (
	"fmt"
	"math/rand"
)

// SelectEntries selects count entries from the ones provided.
// The entriesType string is used in the error message to describe the entries slice.
func SelectEntries[E any](r *rand.Rand, entries []E, count int, entriesType string) ([]E, error) {
	if count == 0 {
		return nil, nil
	}
	if len(entries) < count {
		return nil, fmt.Errorf("cannot choose %d %s because there are only %d", count, entriesType, len(entries))
	}
	if count == 1 {
		if len(entries) == 1 {
			return entries, nil
		}
		pivot := r.Intn(len(entries))
		return entries[pivot : pivot+1], nil
	}

	randomized := make([]E, 0, len(entries))
	randomized = append(randomized, entries...)
	r.Shuffle(len(randomized), func(i, j int) {
		randomized[i], randomized[j] = randomized[j], randomized[i]
	})
	return randomized[:count], nil
}

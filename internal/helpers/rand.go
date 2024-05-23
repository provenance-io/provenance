package helpers

import (
	"fmt"
	"math/rand"
)

// RandIntBetween generates a random number between min and max inclusive.
func RandIntBetween(r *rand.Rand, min, max int) int {
	return r.Intn(max-min+1) + min
}

func SelectRandomEntries[E any](r *rand.Rand, entries []E, count int) ([]E, error) {
	if count == 0 {
		return nil, nil
	}
	if len(entries) < count {
		return nil, fmt.Errorf("cannot choose %d entries because there are only %d", count, len(entries))
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

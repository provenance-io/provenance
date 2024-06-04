package rand

import "math/rand"

// IntBetween generates a random number between min and max inclusive.
func IntBetween(r *rand.Rand, min, max int) int {
	return r.Intn(max-min+1) + min
}

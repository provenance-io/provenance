package rand

import "math/rand"

// IntBetween generates a random number between min and max inclusive.
func IntBetween(r *rand.Rand, minVal, maxVal int) int {
	return r.Intn(maxVal-minVal+1) + minVal
}

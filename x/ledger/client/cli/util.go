package cli

func sliceToMap[T any, K comparable](list []T, keyFn func(T) K) map[K]T {
	result := make(map[K]T, len(list))
	for _, item := range list {
		result[keyFn(item)] = item
	}
	return result
}

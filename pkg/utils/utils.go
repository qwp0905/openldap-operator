package utils

func MergeMap[K comparable, V comparable](m1 map[K]V, m2 map[K]V) map[K]V {
	result := map[K]V{}

	for k, v := range m1 {
		result[k] = v
	}

	for k, v := range m2 {
		result[k] = v
	}

	return result
}

func Find[T comparable](s []T, cb func(e T) bool) *T {
	for _, e := range s {
		if cb(e) {
			return &e
		}
	}

	return nil
}

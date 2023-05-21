package utils

func MergeMap[K comparable, V comparable](m1 map[K]V, m2 ...map[K]V) map[K]V {
	result := map[K]V{}

	for _, m := range append([]map[K]V{m1}, m2...) {
		for k, v := range m {
			result[k] = v
		}
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

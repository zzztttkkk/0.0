package utils

func SliceFind[T comparable](s []T, ele T) int {
	for i, v := range s {
		if v == ele {
			return i
		}
	}
	return -1
}

func SliceFilter[T any](a []T, fn func(T) bool) []T {
	var r []T
	for _, ele := range a {
		if fn(ele) {
			r = append(r, ele)
		}
	}
	return r
}

func SliceMap[T any, V any](a []T, fn func(int, T) V) []V {
	var r = make([]V, 0, len(a))
	for i, ele := range a {
		r = append(r, fn(i, ele))
	}
	return r
}

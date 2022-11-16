package utils

func SliceFind[T comparable](s []T, ele T) int {
	for i, v := range s {
		if v == ele {
			return i
		}
	}
	return -1
}

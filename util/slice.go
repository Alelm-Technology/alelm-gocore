package util

func Map[T any, U any](input []T, fn func(T) U) []U {
	result := make([]U, len(input))
	for i, v := range input {
		result[i] = fn(v)
	}
	return result
}

func Filter[T any](input []T, fn func(T) bool) []T {
	result := make([]T, 0, len(input))
	for _, v := range input {
		if fn(v) {
			result = append(result, v)
		}
	}
	return result
}

func SliceContains[T comparable](slice []T, item T) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func Unique[T comparable](input []T) []T {
	seen := make(map[T]struct{}, len(input))
	result := make([]T, 0, len(input))
	for _, v := range input {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}

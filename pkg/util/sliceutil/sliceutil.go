package sliceutil

import "golang.org/x/exp/constraints"

// Repeat returns a slice with value repeated n times.
func Repeat[T any, N constraints.Integer](value T, n N) []T {
	arr := make([]T, n)
	for i := N(0); i < n; i++ {
		arr[i] = value
	}
	return arr
}
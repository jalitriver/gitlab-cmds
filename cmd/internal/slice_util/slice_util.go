package slice_util

// CountElements counts the number of times that element occurs in the slice.
func CountElements[T comparable](xs []T) map[T]int {
	result := map[T]int{}
	for _, x := range xs {
		result[x]++
	}
	return result
}

// SliceToSet returns the set of elements in the slice where the keys
// in the map that return for the elements in the set.
func SliceToSet[T comparable](xs []T) map[T]int {
	return CountElements(xs)
}

// SubtractSet returns the set of elements in xs that are not in ys.
// This is the same as the "relative complement" of ys with respect to
// xs.  The sets can be generated using [SliceToSet].
func SubtractSet[T comparable](xs, ys map[T]int) map[T]int {
	result := map[T]int{}

	// Iterate over xs.
	for x, count := range xs {

		// If the x is not in ys, add it to the result.
		_, ok := ys[x]
		if !ok {
			result[x] = count
		}

	}

	return result
}

// SubtractSlice returns the slice of elements in xs that are not in
// ys.  This is the same as the "relative complement" of ys with
// respect to xs.
func SubtractSlice[T comparable](xs, ys []T) []T {
	result := []T{}

	// Get the set of ys so we can perform quick lookups.
	ysSet := SliceToSet(ys)

	// Keep only the elements in xs that are not in ys.
	for _, x := range xs {
		_, ok := ysSet[x]
		if !ok {
			result = append(result, x)
		}
	}

	return result
}

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Guigui Authors

package basicwidget

import "slices"

func adjustSliceSize[T any](slice []T, size int) []T {
	if len(slice) == size {
		return slice
	}
	if len(slice) < size {
		return slices.Grow(slice, size-len(slice))[:size]
	}
	return slices.Delete(slice, size, len(slice))
}

func MoveItemsInSlice[T any](slice []T, from int, count int, to int) int {
	if count == 0 {
		return from
	}
	if from <= to && to <= from+count {
		return from
	}
	if from < to {
		to -= count
	}

	s := make([]T, count)
	copy(s, slice[from:from+count])
	slice = slices.Delete(slice, from, from+count)
	// Assume that the slice has enough capacity, then the underlying array should not change.
	_ = slices.Insert(slice, to, s...)

	return to
}

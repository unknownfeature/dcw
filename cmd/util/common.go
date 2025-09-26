package util

import (
	"golang.org/x/exp/constraints"
)

func AllEqual[T constraints.Ordered](slice []T, target T) bool {

	if len(slice) == 0 {
		return false
	}

	for _, item := range slice {
		if item != target {
			return false
		}
	}

	return true
}

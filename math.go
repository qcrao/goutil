package goutil

import (
	"math"
)

const defaultPrecision = 8

// Float64Equal compares two float64 numbers for equality to a certain number of decimal places.
// It takes two float64 numbers f1 and f2, and an optional uint8 precision.
// If precision is not provided, it defaults to defaultPrecision.
// The function returns true if the absolute difference between f1 and f2 is less than or equal to the threshold (10 to the power of -precision), indicating that they are equal to the specified number of decimal places.
// It returns false otherwise.
func Float64Equal(f1 float64, f2 float64, optionalPrecision ...uint8) bool {
	var actualPrecision uint8
	if len(optionalPrecision) == 0 {
		actualPrecision = defaultPrecision
	} else {
		actualPrecision = optionalPrecision[0]
	}

	threshold := math.Pow10(-int(actualPrecision))

	return math.Abs(f1-f2) <= threshold
}

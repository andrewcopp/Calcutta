package mathutil

import "math"

// Sigmoid computes the logistic sigmoid function in a numerically stable way.
func Sigmoid(x float64) float64 {
	if x >= 0 {
		z := math.Exp(-x)
		return 1.0 / (1.0 + z)
	}
	z := math.Exp(x)
	return z / (1.0 + z)
}

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

// WinProb computes the win probability for team1 given net ratings and a scale factor.
func WinProb(net1, net2, scale float64) float64 {
	if scale <= 0 {
		return 0.5
	}
	return Sigmoid((net1 - net2) / scale)
}

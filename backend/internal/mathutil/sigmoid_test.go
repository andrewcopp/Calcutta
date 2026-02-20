package mathutil

import (
	"math"
	"testing"
)

func TestThatSigmoidReturnsPointFiveForZero(t *testing.T) {
	// GIVEN x = 0
	// WHEN computing Sigmoid
	result := Sigmoid(0)

	// THEN result is 0.5
	if result != 0.5 {
		t.Errorf("expected 0.5, got %v", result)
	}
}

func TestThatSigmoidReturnsGreaterThanPointFiveForPositiveInput(t *testing.T) {
	// GIVEN a positive x
	// WHEN computing Sigmoid
	result := Sigmoid(1.0)

	// THEN result > 0.5
	if result <= 0.5 {
		t.Errorf("expected > 0.5, got %v", result)
	}
}

func TestThatSigmoidReturnsLessThanPointFiveForNegativeInput(t *testing.T) {
	// GIVEN a negative x
	// WHEN computing Sigmoid
	result := Sigmoid(-1.0)

	// THEN result < 0.5
	if result >= 0.5 {
		t.Errorf("expected < 0.5, got %v", result)
	}
}

func TestThatSigmoidApproachesOneForLargePositive(t *testing.T) {
	// GIVEN x = 500
	// WHEN computing Sigmoid
	result := Sigmoid(500)

	// THEN result is very close to 1.0
	if result < 0.999 {
		t.Errorf("expected close to 1.0, got %v", result)
	}
}

func TestThatSigmoidApproachesZeroForLargeNegative(t *testing.T) {
	// GIVEN x = -500
	// WHEN computing Sigmoid
	result := Sigmoid(-500)

	// THEN result is very close to 0.0
	if result > 0.001 {
		t.Errorf("expected close to 0.0, got %v", result)
	}
}

func TestThatSigmoidIsSymmetric(t *testing.T) {
	// GIVEN x = 2.5
	// WHEN computing Sigmoid(x) + Sigmoid(-x)
	sum := Sigmoid(2.5) + Sigmoid(-2.5)

	// THEN the sum equals 1.0
	if math.Abs(sum-1.0) > 1e-12 {
		t.Errorf("expected Sigmoid(x) + Sigmoid(-x) = 1.0, got %v", sum)
	}
}

func TestThatSigmoidIsStableForVeryLargePositive(t *testing.T) {
	// GIVEN x = 1000
	// WHEN computing Sigmoid
	result := Sigmoid(1000)

	// THEN result is finite (no NaN or Inf)
	if math.IsNaN(result) || math.IsInf(result, 0) {
		t.Errorf("expected finite result, got %v", result)
	}
}

func TestThatSigmoidIsStableForVeryLargeNegative(t *testing.T) {
	// GIVEN x = -1000
	// WHEN computing Sigmoid
	result := Sigmoid(-1000)

	// THEN result is finite (no NaN or Inf)
	if math.IsNaN(result) || math.IsInf(result, 0) {
		t.Errorf("expected finite result, got %v", result)
	}
}

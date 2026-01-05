package db

import (
	"math"
	"testing"
)

func TestThatComputeRationalInvestmentReturnsZeroWhenPoolSizeIsNonPositive(t *testing.T) {
	givenExpectedValue := 10.0
	givenTotalExpectedValue := 100.0
	givenPoolSize := 0.0

	whenGot := computeRationalInvestment(givenExpectedValue, givenTotalExpectedValue, givenPoolSize)

	thenWant := 0.0
	if whenGot != thenWant {
		t.Fatalf("expected rational=%.2f, got %.2f", thenWant, whenGot)
	}
}

func TestThatComputeRationalInvestmentAllocatesProportionallyToExpectedValue(t *testing.T) {
	givenExpectedValue := 10.0
	givenTotalExpectedValue := 100.0
	givenPoolSize := 4700.0

	whenGot := computeRationalInvestment(givenExpectedValue, givenTotalExpectedValue, givenPoolSize)

	thenWant := 470.0
	if math.Abs(whenGot-thenWant) > 1e-9 {
		t.Fatalf("expected rational=%.10f, got %.10f", thenWant, whenGot)
	}
}

func TestThatComputeDeltaPercentReturnsZeroWhenRationalIsNonPositive(t *testing.T) {
	givenPredicted := 123.0
	givenRational := 0.0

	whenGot := computeDeltaPercent(givenPredicted, givenRational)

	thenWant := 0.0
	if whenGot != thenWant {
		t.Fatalf("expected delta=%.2f, got %.2f", thenWant, whenGot)
	}
}

func TestThatComputeDeltaPercentComputesPercentDifference(t *testing.T) {
	givenPredicted := 120.0
	givenRational := 100.0

	whenGot := computeDeltaPercent(givenPredicted, givenRational)

	thenWant := 20.0
	if math.Abs(whenGot-thenWant) > 1e-9 {
		t.Fatalf("expected delta=%.10f, got %.10f", thenWant, whenGot)
	}
}

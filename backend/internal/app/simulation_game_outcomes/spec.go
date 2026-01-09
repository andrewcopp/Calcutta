package simulation_game_outcomes

import (
	"errors"
	"math"
	"strings"
)

type Spec struct {
	Kind  string  `json:"kind"`
	Sigma float64 `json:"sigma"`
}

func (s *Spec) Normalize() {
	if s == nil {
		return
	}
	k := strings.TrimSpace(strings.ToLower(s.Kind))
	if k == "" {
		k = "kenpom"
	}
	s.Kind = k
	if s.Sigma <= 0 {
		s.Sigma = 10.0
	}
}

func (s *Spec) Validate() error {
	if s == nil {
		return nil
	}
	s.Normalize()
	if s.Kind != "kenpom" {
		return errors.New("unsupported game outcome spec kind")
	}
	if s.Sigma <= 0 {
		return errors.New("sigma must be positive")
	}
	return nil
}

func (s *Spec) WinProb(net1 float64, net2 float64) float64 {
	if s == nil {
		return 0.5
	}
	s.Normalize()
	if s.Sigma <= 0 {
		return 0.5
	}
	return sigmoid((net1 - net2) / s.Sigma)
}

func sigmoid(x float64) float64 {
	if x >= 0 {
		z := math.Exp(-x)
		return 1.0 / (1.0 + z)
	}
	z := math.Exp(x)
	return z / (1.0 + z)
}

package winprob

import (
	"errors"
	"strings"

	"github.com/andrewcopp/Calcutta/backend/internal/mathutil"
)

// Model is a parameterized win-probability model used by both the prediction
// (deterministic) and simulation (Monte Carlo) pipelines.
type Model struct {
	Kind  string  `json:"kind"`
	Sigma float64 `json:"sigma"`
}

func (m *Model) Normalize() {
	if m == nil {
		return
	}
	k := strings.TrimSpace(strings.ToLower(m.Kind))
	if k == "" {
		k = "kenpom"
	}
	m.Kind = k
	if m.Sigma <= 0 {
		m.Sigma = 10.0
	}
}

func (m *Model) Validate() error {
	if m == nil {
		return errors.New("model must not be nil")
	}
	if m.Kind != "kenpom" {
		return errors.New("unsupported win probability model kind")
	}
	if m.Sigma <= 0 {
		return errors.New("sigma must be positive")
	}
	return nil
}

func (m *Model) WinProb(net1 float64, net2 float64) float64 {
	return mathutil.Sigmoid((net1 - net2) / m.Sigma)
}

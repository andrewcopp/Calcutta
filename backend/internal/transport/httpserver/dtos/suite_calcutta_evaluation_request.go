package dtos

import (
	"strings"

	"github.com/google/uuid"
)

type CreateSimulationRunRequest struct {
	CalcuttaID           string  `json:"calcuttaId"`
	SimulatedCalcuttaID  *string `json:"simulatedCalcuttaId"`
	CohortID             *string `json:"cohortId"`
	SimulationRunBatchID *string `json:"simulationRunBatchId"`
	OptimizerKey         *string `json:"optimizerKey"`
	GameOutcomeRunID     *string `json:"gameOutcomeRunId"`
	MarketShareRunID     *string `json:"marketShareRunId"`
	NSims                int     `json:"nSims"`
	Seed                 int     `json:"seed"`
	StartingStateKey     string  `json:"startingStateKey"`
	ExcludedEntryName    *string `json:"excludedEntryName"`
}

func (r *CreateSimulationRunRequest) Validate() error {
	if r.SimulatedCalcuttaID != nil {
		trimmed := strings.TrimSpace(*r.SimulatedCalcuttaID)
		if trimmed == "" {
			r.SimulatedCalcuttaID = nil
		} else {
			if _, err := uuid.Parse(trimmed); err != nil {
				return ErrFieldInvalid("simulatedCalcuttaId", "must be a valid UUID")
			}
			r.SimulatedCalcuttaID = &trimmed
		}
	}

	calcuttaID := strings.TrimSpace(r.CalcuttaID)
	if calcuttaID != "" {
		if _, err := uuid.Parse(calcuttaID); err != nil {
			return ErrFieldInvalid("calcuttaId", "must be a valid UUID")
		}
		r.CalcuttaID = calcuttaID
	}

	if calcuttaID == "" && r.SimulatedCalcuttaID == nil {
		return ErrFieldRequired("calcuttaId")
	}

	if r.SimulationRunBatchID != nil {
		trimmed := strings.TrimSpace(*r.SimulationRunBatchID)
		if trimmed == "" {
			r.SimulationRunBatchID = nil
		} else {
			if _, err := uuid.Parse(trimmed); err != nil {
				return ErrFieldInvalid("simulationRunBatchId", "must be a valid UUID")
			}
			r.SimulationRunBatchID = &trimmed
		}
	}

	if r.CohortID != nil {
		trimmed := strings.TrimSpace(*r.CohortID)
		if trimmed == "" {
			r.CohortID = nil
		} else {
			if _, err := uuid.Parse(trimmed); err != nil {
				return ErrFieldInvalid("cohortId", "must be a valid UUID")
			}
			r.CohortID = &trimmed
		}
	}

	if r.OptimizerKey != nil {
		trimmed := strings.TrimSpace(*r.OptimizerKey)
		if trimmed == "" {
			r.OptimizerKey = nil
		} else {
			r.OptimizerKey = &trimmed
		}
	}

	if r.GameOutcomeRunID != nil {
		trimmed := strings.TrimSpace(*r.GameOutcomeRunID)
		if trimmed == "" {
			r.GameOutcomeRunID = nil
		} else {
			if _, err := uuid.Parse(trimmed); err != nil {
				return ErrFieldInvalid("gameOutcomeRunId", "must be a valid UUID")
			}
			r.GameOutcomeRunID = &trimmed
		}
	}

	if r.MarketShareRunID != nil {
		trimmed := strings.TrimSpace(*r.MarketShareRunID)
		if trimmed == "" {
			r.MarketShareRunID = nil
		} else {
			if _, err := uuid.Parse(trimmed); err != nil {
				return ErrFieldInvalid("marketShareRunId", "must be a valid UUID")
			}
			r.MarketShareRunID = &trimmed
		}
	}

	if r.SimulationRunBatchID == nil {
		if r.GameOutcomeRunID == nil {
			return ErrFieldRequired("gameOutcomeRunId")
		}
	}

	if strings.TrimSpace(r.StartingStateKey) == "" {
		r.StartingStateKey = "post_first_four"
	}
	if r.StartingStateKey != "post_first_four" && r.StartingStateKey != "current" {
		return ErrFieldInvalid("startingStateKey", "must be 'current' or 'post_first_four'")
	}

	if r.SimulationRunBatchID == nil {
		if r.CohortID == nil {
			return ErrFieldRequired("cohortId")
		}
		if r.NSims <= 0 {
			return ErrFieldInvalid("nSims", "must be positive")
		}
		if r.Seed == 0 {
			r.Seed = 42
		}
	}

	if r.ExcludedEntryName != nil {
		trimmed := strings.TrimSpace(*r.ExcludedEntryName)
		if trimmed == "" {
			r.ExcludedEntryName = nil
		} else {
			r.ExcludedEntryName = &trimmed
		}
	}

	return nil
}

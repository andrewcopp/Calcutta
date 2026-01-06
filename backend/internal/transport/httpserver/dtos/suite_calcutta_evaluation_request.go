package dtos

import (
	"strings"

	"github.com/google/uuid"
)

type CreateSuiteCalcuttaEvaluationRequest struct {
	CalcuttaID           string  `json:"calcuttaId"`
	SuiteExecutionID     *string `json:"suiteExecutionId"`
	SuiteID              *string `json:"suiteId"`
	SuiteName            *string `json:"suiteName"`
	CohortID             *string `json:"cohortId"`
	CohortName           *string `json:"cohortName"`
	SimulationRunBatchID *string `json:"simulationRunBatchId"`
	OptimizerKey         *string `json:"optimizerKey"`
	GameOutcomeRunID     *string `json:"gameOutcomeRunId"`
	MarketShareRunID     *string `json:"marketShareRunId"`
	NSims                int     `json:"nSims"`
	Seed                 int     `json:"seed"`
	StartingStateKey     string  `json:"startingStateKey"`
	ExcludedEntryName    *string `json:"excludedEntryName"`
}

func (r *CreateSuiteCalcuttaEvaluationRequest) Validate() error {
	// Compatibility aliases: prefer explicit suite* fields, but accept cohort*/simulationRunBatch*.
	if r.SuiteID == nil && r.CohortID != nil {
		r.SuiteID = r.CohortID
	}
	if r.SuiteName == nil && r.CohortName != nil {
		r.SuiteName = r.CohortName
	}
	if r.SuiteExecutionID == nil && r.SimulationRunBatchID != nil {
		r.SuiteExecutionID = r.SimulationRunBatchID
	}

	if strings.TrimSpace(r.CalcuttaID) == "" {
		return ErrFieldRequired("calcuttaId")
	}
	if _, err := uuid.Parse(strings.TrimSpace(r.CalcuttaID)); err != nil {
		return ErrFieldInvalid("calcuttaId", "must be a valid UUID")
	}

	if r.SuiteExecutionID != nil {
		trimmed := strings.TrimSpace(*r.SuiteExecutionID)
		if trimmed == "" {
			r.SuiteExecutionID = nil
		} else {
			if _, err := uuid.Parse(trimmed); err != nil {
				return ErrFieldInvalid("suiteExecutionId", "must be a valid UUID")
			}
			r.SuiteExecutionID = &trimmed
		}
	}

	if r.SuiteID != nil {
		trimmed := strings.TrimSpace(*r.SuiteID)
		if trimmed == "" {
			r.SuiteID = nil
		} else {
			if _, err := uuid.Parse(trimmed); err != nil {
				return ErrFieldInvalid("suiteId", "must be a valid UUID")
			}
			r.SuiteID = &trimmed
		}
	}

	if r.SuiteName != nil {
		trimmed := strings.TrimSpace(*r.SuiteName)
		if trimmed == "" {
			r.SuiteName = nil
		} else {
			r.SuiteName = &trimmed
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

	if r.SuiteExecutionID == nil {
		if r.GameOutcomeRunID == nil {
			return ErrFieldRequired("gameOutcomeRunId")
		}
		if r.MarketShareRunID == nil {
			return ErrFieldRequired("marketShareRunId")
		}
	}

	if strings.TrimSpace(r.StartingStateKey) == "" {
		r.StartingStateKey = "post_first_four"
	}
	if r.StartingStateKey != "post_first_four" && r.StartingStateKey != "current" {
		return ErrFieldInvalid("startingStateKey", "must be 'current' or 'post_first_four'")
	}

	if r.SuiteExecutionID == nil {
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

	if r.SuiteID == nil && r.SuiteExecutionID == nil {
		// If not specifying SuiteID, require enough info to resolve/create.
		if r.SuiteName == nil {
			return ErrFieldRequired("suiteName")
		}
		if r.OptimizerKey == nil {
			return ErrFieldRequired("optimizerKey")
		}
	}

	return nil
}

package dtos

import (
	"strings"

	"github.com/google/uuid"
)

type CreateEntryEvaluationRequest struct {
	CalcuttaID        string  `json:"calcuttaId"`
	EntryCandidateID  string  `json:"entryCandidateId"`
	ExcludedEntryName *string `json:"excludedEntryName"`
	StartingStateKey  string  `json:"startingStateKey"`
	NSims             int     `json:"nSims"`
	Seed              int     `json:"seed"`
	ExperimentKey     *string `json:"experimentKey"`
	RequestSource     *string `json:"requestSource"`
}

func (r *CreateEntryEvaluationRequest) Validate() error {
	if strings.TrimSpace(r.CalcuttaID) == "" {
		return ErrFieldRequired("calcuttaId")
	}
	if _, err := uuid.Parse(strings.TrimSpace(r.CalcuttaID)); err != nil {
		return ErrFieldInvalid("calcuttaId", "must be a valid UUID")
	}

	if strings.TrimSpace(r.EntryCandidateID) == "" {
		return ErrFieldRequired("entryCandidateId")
	}
	if _, err := uuid.Parse(strings.TrimSpace(r.EntryCandidateID)); err != nil {
		return ErrFieldInvalid("entryCandidateId", "must be a valid UUID")
	}

	if strings.TrimSpace(r.StartingStateKey) == "" {
		return ErrFieldRequired("startingStateKey")
	}
	if r.StartingStateKey != "post_first_four" {
		return ErrFieldInvalid("startingStateKey", "must be 'post_first_four'")
	}

	if r.NSims <= 0 {
		return ErrFieldInvalid("nSims", "must be positive")
	}

	if r.ExcludedEntryName != nil {
		trimmed := strings.TrimSpace(*r.ExcludedEntryName)
		if trimmed == "" {
			r.ExcludedEntryName = nil
		} else {
			r.ExcludedEntryName = &trimmed
		}
	}

	if r.ExperimentKey != nil {
		trimmed := strings.TrimSpace(*r.ExperimentKey)
		if trimmed == "" {
			r.ExperimentKey = nil
		} else {
			r.ExperimentKey = &trimmed
		}
	}

	if r.RequestSource != nil {
		trimmed := strings.TrimSpace(*r.RequestSource)
		if trimmed == "" {
			r.RequestSource = nil
		} else {
			r.RequestSource = &trimmed
		}
	}

	return nil
}

package suite_scenarios

import "errors"

var (
	ErrCalcuttaNotFound              = errors.New("calcutta not found")
	ErrSimulatedCalcuttaNotFound     = errors.New("simulated calcutta not found")
	ErrSimulatedEntryNotFound        = errors.New("simulated entry not found")
	ErrCandidateNotFound             = errors.New("candidate not found")
	ErrCandidateHasNoBids            = errors.New("candidate has no bids")
	ErrCandidateInvalidTeamID        = errors.New("candidate has invalid team_id")
	ErrCandidateInvalidBidPoints     = errors.New("candidate has invalid bid_points")
	ErrDuplicatePayoutPosition       = errors.New("duplicate payout position")
	ErrDuplicateScoringRuleWinIndex  = errors.New("duplicate scoring rule winIndex")
	ErrHighlightedEntryDoesNotBelong = errors.New("highlighted simulated entry does not belong")
)

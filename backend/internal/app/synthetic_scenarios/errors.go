package synthetic_scenarios

import "errors"

var (
	ErrSyntheticCalcuttaNotFound      = errors.New("synthetic calcutta not found")
	ErrSyntheticCalcuttaHasNoSnapshot = errors.New("synthetic calcutta has no snapshot")
	ErrHighlightedEntryDoesNotBelong  = errors.New("highlighted entry does not belong to snapshot")

	ErrEntryArtifactNotFound       = errors.New("entry artifact not found")
	ErrEntryArtifactHasNoRunID     = errors.New("entry artifact has no run_id")
	ErrEntryArtifactNotMetrics     = errors.New("entry artifact must be metrics")
	ErrNoStrategyGenerationRunBids = errors.New("no strategy_generation_run_bids found")

	ErrSyntheticEntryNotFound          = errors.New("synthetic entry not found")
	ErrOnlyManualCandidatesCanBeEdited = errors.New("only manual candidates can be edited")
)

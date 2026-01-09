package suite_evaluations

import "errors"

var (
	ErrSimulationNotFound                     = errors.New("simulation not found")
	ErrEvaluationHasNoCalcuttaEvaluationRunID = errors.New("evaluation has no calcutta_evaluation_run_id")
	ErrSnapshotEntryNotFoundForEvaluation     = errors.New("snapshot entry not found for evaluation")
	ErrMissingMarketShareRunForBatch          = errors.New("missing market-share run for simulation batch")
)

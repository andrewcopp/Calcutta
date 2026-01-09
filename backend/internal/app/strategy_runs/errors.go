package strategy_runs

import "errors"

var (
	ErrMarketShareArtifactNotFound   = errors.New("market share artifact not found")
	ErrMarketShareRunMissingMetrics  = errors.New("market share run missing metrics artifact")
	ErrEntryRunNotFound              = errors.New("entry run not found")
	ErrStrategyGenerationRunNotFound = errors.New("strategy generation run not found")
	ErrEntryArtifactNotFound         = errors.New("entry artifact not found")
)

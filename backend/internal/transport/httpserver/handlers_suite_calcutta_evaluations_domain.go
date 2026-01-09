package httpserver

import (
	"context"
)

func (s *Server) computeHypotheticalFinishByEntryNameForStrategyRun(ctx context.Context, calcuttaID string, strategyGenerationRunID string) (map[string]*suiteCalcuttaEvalFinish, bool, error) {
	if calcuttaID == "" || strategyGenerationRunID == "" {
		return nil, false, nil
	}
	items, ok, err := s.app.SuiteEvaluations.ComputeHypotheticalFinishByEntryName(ctx, calcuttaID, strategyGenerationRunID)
	if err != nil {
		return nil, false, err
	}
	if !ok {
		return nil, false, nil
	}
	out := make(map[string]*suiteCalcuttaEvalFinish)
	for name, f := range items {
		if f == nil {
			continue
		}
		out[name] = &suiteCalcuttaEvalFinish{
			FinishPosition: f.FinishPosition,
			IsTied:         f.IsTied,
			InTheMoney:     f.InTheMoney,
			PayoutCents:    f.PayoutCents,
			TotalPoints:    f.TotalPoints,
		}
	}
	return out, true, nil
}

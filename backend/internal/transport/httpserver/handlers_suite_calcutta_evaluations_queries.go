package httpserver

import (
	"context"
	"errors"
	"strings"

	"github.com/andrewcopp/Calcutta/backend/internal/app/suite_evaluations"
	"github.com/jackc/pgx/v5"
)

func (s *Server) loadSuiteCalcuttaEvaluations(
	ctx context.Context,
	calcuttaID string,
	suiteID string,
	suiteExecutionID string,
	limit int,
	offset int,
) ([]suiteCalcuttaEvaluationListItem, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	var calcuttaIDPtr *string
	if strings.TrimSpace(calcuttaID) != "" {
		v := strings.TrimSpace(calcuttaID)
		calcuttaIDPtr = &v
	}
	var cohortIDPtr *string
	if strings.TrimSpace(suiteID) != "" {
		v := strings.TrimSpace(suiteID)
		cohortIDPtr = &v
	}
	var batchIDPtr *string
	if strings.TrimSpace(suiteExecutionID) != "" {
		v := strings.TrimSpace(suiteExecutionID)
		batchIDPtr = &v
	}

	rows, err := s.app.SuiteEvaluations.ListEvaluations(ctx, calcuttaIDPtr, cohortIDPtr, batchIDPtr, limit, offset)
	if err != nil {
		return nil, err
	}

	items := make([]suiteCalcuttaEvaluationListItem, 0, len(rows))
	for _, it := range rows {
		items = append(items, suiteCalcuttaEvaluationListItem{
			ID:                        it.ID,
			SimulationBatchID:         it.SimulationBatchID,
			CohortID:                  it.CohortID,
			CohortName:                it.CohortName,
			OptimizerKey:              it.OptimizerKey,
			NSims:                     it.NSims,
			Seed:                      it.Seed,
			OurRank:                   it.OurRank,
			OurMeanNormalizedPayout:   it.OurMeanNormalizedPayout,
			OurMedianNormalizedPayout: it.OurMedianNormalizedPayout,
			OurPTop1:                  it.OurPTop1,
			OurPInMoney:               it.OurPInMoney,
			TotalSimulations:          it.TotalSimulations,
			CalcuttaID:                it.CalcuttaID,
			GameOutcomeRunID:          it.GameOutcomeRunID,
			MarketShareRunID:          it.MarketShareRunID,
			StrategyGenerationRunID:   it.StrategyGenerationRunID,
			CalcuttaEvaluationRunID:   it.CalcuttaEvaluationRunID,
			RealizedFinishPosition:    it.RealizedFinishPosition,
			RealizedIsTied:            it.RealizedIsTied,
			RealizedInTheMoney:        it.RealizedInTheMoney,
			RealizedPayoutCents:       it.RealizedPayoutCents,
			RealizedTotalPoints:       it.RealizedTotalPoints,
			StartingStateKey:          it.StartingStateKey,
			ExcludedEntryName:         it.ExcludedEntryName,
			Status:                    it.Status,
			ClaimedAt:                 it.ClaimedAt,
			ClaimedBy:                 it.ClaimedBy,
			ErrorMessage:              it.ErrorMessage,
			CreatedAt:                 it.CreatedAt,
			UpdatedAt:                 it.UpdatedAt,
		})
	}

	return items, nil
}

func (s *Server) loadSuiteCalcuttaEvaluationByID(ctx context.Context, id string) (*suiteCalcuttaEvaluationListItem, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	it, err := s.app.SuiteEvaluations.GetEvaluation(ctx, id)
	if err != nil {
		if errors.Is(err, suite_evaluations.ErrSimulationNotFound) {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}
	out := &suiteCalcuttaEvaluationListItem{
		ID:                        it.ID,
		SimulationBatchID:         it.SimulationBatchID,
		CohortID:                  it.CohortID,
		CohortName:                it.CohortName,
		OptimizerKey:              it.OptimizerKey,
		NSims:                     it.NSims,
		Seed:                      it.Seed,
		OurRank:                   it.OurRank,
		OurMeanNormalizedPayout:   it.OurMeanNormalizedPayout,
		OurMedianNormalizedPayout: it.OurMedianNormalizedPayout,
		OurPTop1:                  it.OurPTop1,
		OurPInMoney:               it.OurPInMoney,
		TotalSimulations:          it.TotalSimulations,
		CalcuttaID:                it.CalcuttaID,
		GameOutcomeRunID:          it.GameOutcomeRunID,
		MarketShareRunID:          it.MarketShareRunID,
		StrategyGenerationRunID:   it.StrategyGenerationRunID,
		CalcuttaEvaluationRunID:   it.CalcuttaEvaluationRunID,
		RealizedFinishPosition:    it.RealizedFinishPosition,
		RealizedIsTied:            it.RealizedIsTied,
		RealizedInTheMoney:        it.RealizedInTheMoney,
		RealizedPayoutCents:       it.RealizedPayoutCents,
		RealizedTotalPoints:       it.RealizedTotalPoints,
		StartingStateKey:          it.StartingStateKey,
		ExcludedEntryName:         it.ExcludedEntryName,
		Status:                    it.Status,
		ClaimedAt:                 it.ClaimedAt,
		ClaimedBy:                 it.ClaimedBy,
		ErrorMessage:              it.ErrorMessage,
		CreatedAt:                 it.CreatedAt,
		UpdatedAt:                 it.UpdatedAt,
	}
	return out, nil
}

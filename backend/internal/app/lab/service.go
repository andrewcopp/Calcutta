package lab

import "context"

// Service provides lab-related business logic.
type Service struct {
	repo Repository
}

// New creates a new lab service.
func New(repo Repository) *Service {
	return &Service{repo: repo}
}

// ListInvestmentModels returns investment models matching the filter.
func (s *Service) ListInvestmentModels(ctx context.Context, filter ListModelsFilter, page Pagination) ([]InvestmentModel, error) {
	if page.Limit <= 0 {
		page.Limit = 50
	}
	if page.Limit > 200 {
		page.Limit = 200
	}
	if page.Offset < 0 {
		page.Offset = 0
	}
	return s.repo.ListInvestmentModels(filter, page)
}

// GetInvestmentModel returns a single investment model by ID.
func (s *Service) GetInvestmentModel(ctx context.Context, id string) (*InvestmentModel, error) {
	return s.repo.GetInvestmentModel(id)
}

// GetModelLeaderboard returns the model leaderboard sorted by avg mean payout.
func (s *Service) GetModelLeaderboard(ctx context.Context) ([]LeaderboardEntry, error) {
	return s.repo.GetModelLeaderboard()
}

// ListEntries returns entries matching the filter.
func (s *Service) ListEntries(ctx context.Context, filter ListEntriesFilter, page Pagination) ([]EntryDetail, error) {
	if page.Limit <= 0 {
		page.Limit = 50
	}
	if page.Limit > 200 {
		page.Limit = 200
	}
	if page.Offset < 0 {
		page.Offset = 0
	}
	return s.repo.ListEntries(filter, page)
}

// GetEntry returns a single entry by ID with full details.
func (s *Service) GetEntry(ctx context.Context, id string) (*EntryDetail, error) {
	return s.repo.GetEntry(id)
}

// GetEntryEnriched returns a single entry with enriched bids (team names, seeds, naive allocation).
func (s *Service) GetEntryEnriched(ctx context.Context, id string) (*EntryDetailEnriched, error) {
	return s.repo.GetEntryEnriched(id)
}

// GetEntryEnrichedByModelAndCalcutta returns an enriched entry for a model/calcutta pair.
func (s *Service) GetEntryEnrichedByModelAndCalcutta(ctx context.Context, modelName, calcuttaID, startingStateKey string) (*EntryDetailEnriched, error) {
	return s.repo.GetEntryEnrichedByModelAndCalcutta(modelName, calcuttaID, startingStateKey)
}

// ListEvaluations returns evaluations matching the filter.
func (s *Service) ListEvaluations(ctx context.Context, filter ListEvaluationsFilter, page Pagination) ([]EvaluationDetail, error) {
	if page.Limit <= 0 {
		page.Limit = 50
	}
	if page.Limit > 200 {
		page.Limit = 200
	}
	if page.Offset < 0 {
		page.Offset = 0
	}
	return s.repo.ListEvaluations(filter, page)
}

// GetEvaluation returns a single evaluation by ID with full details.
func (s *Service) GetEvaluation(ctx context.Context, id string) (*EvaluationDetail, error) {
	return s.repo.GetEvaluation(id)
}

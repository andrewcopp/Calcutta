package lab_candidates

import (
	"context"

	"github.com/andrewcopp/Calcutta/backend/internal/app/apperrors"
	"github.com/google/uuid"
)

type GenerateCandidatesFromAlgorithmsRequest struct {
	GameOutcomesAlgorithmID string
	MarketShareAlgorithmID  string
	OptimizerKey            string
	StartingStateKey        string
	ExcludedEntryName       *string
	DisplayName             string
}

type GenerateCandidatesFromAlgorithmsResult struct {
	TotalCalcuttas         int
	EligibleCalcuttas      int
	CreatedCandidates      int
	SkippedExisting        int
	SkippedMissingUpstream int
}

type ListCandidatesFilter struct {
	CalcuttaID              *string
	TournamentID            *string
	StrategyGenerationRunID *string
	MarketShareArtifactID   *string
	AdvancementRunID        *string
	GameOutcomesAlgorithmID *string
	MarketShareAlgorithmID  *string
	OptimizerKey            *string
	StartingStateKey        *string
	ExcludedEntryName       *string
	SourceKind              *string
}

type CandidateComboCoverage struct {
	GameOutcomesAlgorithmID string
	MarketShareAlgorithmID  string
	OptimizerKey            string
	ExistingCandidates      int
}

type CandidateComboCoverageSummary struct {
	TotalCalcuttas int
	Items          []CandidateComboCoverage
}

type ListCandidatesPagination struct {
	Limit  int
	Offset int
}

type CreateCandidateRequest struct {
	CalcuttaID            string
	AdvancementRunID      string
	MarketShareArtifactID string
	OptimizerKey          string
	StartingStateKey      string
	ExcludedEntryName     *string
	DisplayName           *string
}

type CreateCandidateResult struct {
	CandidateID             string
	StrategyGenerationRunID string
}

type CandidateDetailTeam struct {
	TeamID    string
	BidPoints int
}

type CandidateDetail struct {
	CandidateID             string
	DisplayName             string
	SourceKind              string
	SourceEntryArtifactID   *string
	CalcuttaID              string
	CalcuttaName            string
	TournamentID            string
	StrategyGenerationRunID string
	MarketShareRunID        string
	MarketShareArtifactID   string
	AdvancementRunID        string
	OptimizerKey            string
	StartingStateKey        string
	SeedPreview             string
	ExcludedEntryName       *string
	GitSHA                  *string
	Teams                   []CandidateDetailTeam
}

type Repository interface {
	GetCandidateDetail(ctx context.Context, candidateID string) (*CandidateDetail, error)
	ListCandidates(ctx context.Context, filter ListCandidatesFilter, page ListCandidatesPagination) ([]CandidateDetail, error)
	ListCandidateComboCoverage(ctx context.Context, startingStateKey string, excludedEntryName *string) (*CandidateComboCoverageSummary, error)
	CreateCandidatesBulk(ctx context.Context, items []CreateCandidateRequest) ([]CreateCandidateResult, error)
	GenerateCandidatesFromAlgorithms(ctx context.Context, req GenerateCandidatesFromAlgorithmsRequest) (*GenerateCandidatesFromAlgorithmsResult, error)
	DeleteCandidate(ctx context.Context, candidateID string) error
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetCandidateDetail(ctx context.Context, candidateID string) (*CandidateDetail, error) {
	if candidateID == "" {
		return nil, &apperrors.InvalidArgumentError{Field: "candidateId", Message: "candidateId is required"}
	}
	if _, err := uuid.Parse(candidateID); err != nil {
		return nil, &apperrors.InvalidArgumentError{Field: "candidateId", Message: "candidateId must be a valid UUID"}
	}
	return s.repo.GetCandidateDetail(ctx, candidateID)
}

func (s *Service) ListCandidates(ctx context.Context, filter ListCandidatesFilter, page ListCandidatesPagination) ([]CandidateDetail, error) {
	// Basic normalization/validation (keep transport rules stable but allow reuse outside HTTP).
	if filter.CalcuttaID != nil && *filter.CalcuttaID != "" {
		if _, err := uuid.Parse(*filter.CalcuttaID); err != nil {
			return nil, apperrors.FieldInvalid("calcutta_id", "calcutta_id must be a valid UUID")
		}
	}
	if filter.TournamentID != nil && *filter.TournamentID != "" {
		if _, err := uuid.Parse(*filter.TournamentID); err != nil {
			return nil, apperrors.FieldInvalid("tournament_id", "tournament_id must be a valid UUID")
		}
	}
	if filter.StrategyGenerationRunID != nil && *filter.StrategyGenerationRunID != "" {
		if _, err := uuid.Parse(*filter.StrategyGenerationRunID); err != nil {
			return nil, apperrors.FieldInvalid("strategy_generation_run_id", "strategy_generation_run_id must be a valid UUID")
		}
	}
	if filter.MarketShareArtifactID != nil && *filter.MarketShareArtifactID != "" {
		if _, err := uuid.Parse(*filter.MarketShareArtifactID); err != nil {
			return nil, apperrors.FieldInvalid("market_share_artifact_id", "market_share_artifact_id must be a valid UUID")
		}
	}
	if filter.AdvancementRunID != nil && *filter.AdvancementRunID != "" {
		if _, err := uuid.Parse(*filter.AdvancementRunID); err != nil {
			return nil, apperrors.FieldInvalid("advancement_run_id", "advancement_run_id must be a valid UUID")
		}
	}
	if filter.GameOutcomesAlgorithmID != nil && *filter.GameOutcomesAlgorithmID != "" {
		if _, err := uuid.Parse(*filter.GameOutcomesAlgorithmID); err != nil {
			return nil, apperrors.FieldInvalid("game_outcomes_algorithm_id", "game_outcomes_algorithm_id must be a valid UUID")
		}
	}
	if filter.MarketShareAlgorithmID != nil && *filter.MarketShareAlgorithmID != "" {
		if _, err := uuid.Parse(*filter.MarketShareAlgorithmID); err != nil {
			return nil, apperrors.FieldInvalid("market_share_algorithm_id", "market_share_algorithm_id must be a valid UUID")
		}
	}

	limit := page.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	offset := page.Offset
	if offset < 0 {
		offset = 0
	}

	return s.repo.ListCandidates(ctx, filter, ListCandidatesPagination{Limit: limit, Offset: offset})
}

func (s *Service) ListCandidateComboCoverage(ctx context.Context, startingStateKey string, excludedEntryName *string) (*CandidateComboCoverageSummary, error) {
	if startingStateKey == "" {
		startingStateKey = "post_first_four"
	}
	return s.repo.ListCandidateComboCoverage(ctx, startingStateKey, excludedEntryName)
}

func (s *Service) CreateCandidatesBulk(ctx context.Context, items []CreateCandidateRequest) ([]CreateCandidateResult, error) {
	if len(items) == 0 {
		return nil, apperrors.FieldRequired("items")
	}

	for i := range items {
		req := &items[i]
		if req.CalcuttaID == "" {
			return nil, apperrors.FieldRequired("calcuttaId")
		}
		if _, err := uuid.Parse(req.CalcuttaID); err != nil {
			return nil, apperrors.FieldInvalid("calcuttaId", "must be a valid UUID")
		}
		if req.AdvancementRunID == "" {
			return nil, apperrors.FieldRequired("advancementRunId")
		}
		if _, err := uuid.Parse(req.AdvancementRunID); err != nil {
			return nil, apperrors.FieldInvalid("advancementRunId", "must be a valid UUID")
		}
		if req.MarketShareArtifactID == "" {
			return nil, apperrors.FieldRequired("marketShareArtifactId")
		}
		if _, err := uuid.Parse(req.MarketShareArtifactID); err != nil {
			return nil, apperrors.FieldInvalid("marketShareArtifactId", "must be a valid UUID")
		}
		if req.OptimizerKey == "" {
			return nil, apperrors.FieldRequired("optimizerKey")
		}
		if req.StartingStateKey == "" {
			return nil, apperrors.FieldRequired("startingStateKey")
		}
	}

	return s.repo.CreateCandidatesBulk(ctx, items)
}

func (s *Service) GenerateCandidatesFromAlgorithms(ctx context.Context, req GenerateCandidatesFromAlgorithmsRequest) (*GenerateCandidatesFromAlgorithmsResult, error) {
	if req.GameOutcomesAlgorithmID == "" {
		return nil, apperrors.FieldRequired("gameOutcomesAlgorithmId")
	}
	if _, err := uuid.Parse(req.GameOutcomesAlgorithmID); err != nil {
		return nil, apperrors.FieldInvalid("gameOutcomesAlgorithmId", "must be a valid UUID")
	}
	if req.MarketShareAlgorithmID == "" {
		return nil, apperrors.FieldRequired("marketShareAlgorithmId")
	}
	if _, err := uuid.Parse(req.MarketShareAlgorithmID); err != nil {
		return nil, apperrors.FieldInvalid("marketShareAlgorithmId", "must be a valid UUID")
	}
	if req.OptimizerKey == "" {
		return nil, apperrors.FieldRequired("optimizerKey")
	}
	if req.StartingStateKey == "" {
		return nil, apperrors.FieldRequired("startingStateKey")
	}
	if req.DisplayName == "" {
		req.DisplayName = "Lab Candidate"
	}
	return s.repo.GenerateCandidatesFromAlgorithms(ctx, req)
}

func (s *Service) DeleteCandidate(ctx context.Context, candidateID string) error {
	if candidateID == "" {
		return &apperrors.InvalidArgumentError{Field: "candidateId", Message: "candidateId is required"}
	}
	if _, err := uuid.Parse(candidateID); err != nil {
		return &apperrors.InvalidArgumentError{Field: "candidateId", Message: "candidateId must be a valid UUID"}
	}
	return s.repo.DeleteCandidate(ctx, candidateID)
}

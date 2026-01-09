package lab_candidates

import (
	"context"

	"github.com/andrewcopp/Calcutta/backend/internal/app/apperrors"
	"github.com/google/uuid"
)

type ListCandidatesFilter struct {
	CalcuttaID              *string
	TournamentID            *string
	StrategyGenerationRunID *string
	MarketShareArtifactID   *string
	AdvancementRunID        *string
	OptimizerKey            *string
	StartingStateKey        *string
	ExcludedEntryName       *string
	SourceKind              *string
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
	TournamentID            string
	StrategyGenerationRunID string
	MarketShareRunID        string
	MarketShareArtifactID   string
	AdvancementRunID        string
	OptimizerKey            string
	StartingStateKey        string
	ExcludedEntryName       *string
	GitSHA                  *string
	Teams                   []CandidateDetailTeam
}

type Repository interface {
	GetCandidateDetail(ctx context.Context, candidateID string) (*CandidateDetail, error)
	ListCandidates(ctx context.Context, filter ListCandidatesFilter, page ListCandidatesPagination) ([]CandidateDetail, error)
	CreateCandidatesBulk(ctx context.Context, items []CreateCandidateRequest) ([]CreateCandidateResult, error)
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

func (s *Service) DeleteCandidate(ctx context.Context, candidateID string) error {
	if candidateID == "" {
		return &apperrors.InvalidArgumentError{Field: "candidateId", Message: "candidateId is required"}
	}
	if _, err := uuid.Parse(candidateID); err != nil {
		return &apperrors.InvalidArgumentError{Field: "candidateId", Message: "candidateId must be a valid UUID"}
	}
	return s.repo.DeleteCandidate(ctx, candidateID)
}

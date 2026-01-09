package strategy_runs

import (
	"context"
	"encoding/json"
)

type CreateRunParams struct {
	CalcuttaID            string
	Name                  string
	OptimizerKey          string
	RunKey                string
	RunKeyUUID            string
	MarketShareArtifactID *string
	MarketShareRunID      *string
	ParamsJSON            []byte
	GitSHA                *string
	Source                string
}

type CreateRunResult struct {
	RunID  string
	RunKey string
}

type RunListItem struct {
	ID                    string
	RunKey                *string
	Name                  *string
	CalcuttaID            *string
	SimulatedTournamentID *string
	Purpose               string
	ReturnsModelKey       string
	InvestmentModelKey    string
	OptimizerKey          string
	CreatedAt             string
}

type RunArtifactListItem struct {
	ID                         string
	RunID                      string
	RunKey                     *string
	ArtifactKind               string
	SchemaVersion              string
	StorageURI                 *string
	SummaryJSON                json.RawMessage
	InputMarketShareArtifactID *string
	InputAdvancementArtifactID *string
	CreatedAt                  string
	UpdatedAt                  string
}

type Repository interface {
	CreateRun(ctx context.Context, p CreateRunParams) (*CreateRunResult, error)
	ListRuns(ctx context.Context, calcuttaID *string, limit, offset int) ([]RunListItem, error)
	GetRun(ctx context.Context, id string) (*RunListItem, error)

	GetEntryArtifact(ctx context.Context, artifactID string) (*RunArtifactListItem, error)
	ListRunArtifacts(ctx context.Context, runID string) ([]RunArtifactListItem, error)
	GetRunArtifact(ctx context.Context, runID, artifactKind string) (*RunArtifactListItem, error)
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateEntryRun(ctx context.Context, p CreateRunParams) (*CreateRunResult, error) {
	return s.repo.CreateRun(ctx, p)
}

func (s *Service) CreateStrategyGenerationRun(ctx context.Context, p CreateRunParams) (*CreateRunResult, error) {
	return s.repo.CreateRun(ctx, p)
}

func (s *Service) ListRuns(ctx context.Context, calcuttaID *string, limit, offset int) ([]RunListItem, error) {
	return s.repo.ListRuns(ctx, calcuttaID, limit, offset)
}

func (s *Service) GetRun(ctx context.Context, id string) (*RunListItem, error) {
	return s.repo.GetRun(ctx, id)
}

func (s *Service) GetEntryArtifact(ctx context.Context, artifactID string) (*RunArtifactListItem, error) {
	return s.repo.GetEntryArtifact(ctx, artifactID)
}

func (s *Service) ListRunArtifacts(ctx context.Context, runID string) ([]RunArtifactListItem, error) {
	return s.repo.ListRunArtifacts(ctx, runID)
}

func (s *Service) GetRunArtifact(ctx context.Context, runID, artifactKind string) (*RunArtifactListItem, error) {
	return s.repo.GetRunArtifact(ctx, runID, artifactKind)
}

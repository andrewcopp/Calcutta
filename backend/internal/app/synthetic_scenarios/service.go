package synthetic_scenarios

import (
	"context"
	"encoding/json"
	"time"
)

type SyntheticCalcuttaListItem struct {
	ID                        string
	CohortID                  string
	CalcuttaID                string
	CalcuttaSnapshotID        *string
	HighlightedEntryID        *string
	FocusStrategyGenerationID *string
	FocusEntryName            *string
	LatestSimulationStatus    *string
	OurRank                   *int
	OurMeanNormalizedPayout   *float64
	OurPTop1                  *float64
	OurPInMoney               *float64
	TotalSimulations          *int
	StartingStateKey          *string
	ExcludedEntryName         *string
	Notes                     *string
	Metadata                  json.RawMessage
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
}

type CreateSyntheticCalcuttaParams struct {
	CohortID                  string
	CalcuttaID                string
	CalcuttaSnapshotID        *string
	FocusStrategyGenerationID *string
	FocusEntryName            *string
	StartingStateKey          *string
	ExcludedEntryName         *string
}

type PatchSyntheticCalcuttaParams struct {
	HighlightedEntryID *string
	Notes              *string
	Metadata           *json.RawMessage
}

type SyntheticEntryTeam struct {
	TeamID    string
	BidPoints int
}

type SyntheticEntryListItem struct {
	ID            string
	CandidateID   string
	SnapshotEntry string
	EntryID       *string
	DisplayName   string
	IsSynthetic   bool
	Rank          *int
	Mean          *float64
	PTop1         *float64
	PInMoney      *float64
	Teams         []SyntheticEntryTeam
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type CreateSyntheticEntryParams struct {
	SyntheticCalcuttaID string
	DisplayName         string
	Teams               []SyntheticEntryTeam
}

type ImportSyntheticEntryParams struct {
	SyntheticCalcuttaID string
	EntryArtifactID     string
	DisplayName         *string
}

type PatchSyntheticEntryParams struct {
	AttachmentID string
	DisplayName  *string
	Teams        *[]SyntheticEntryTeam
}

type DeleteSyntheticEntryParams struct {
	AttachmentID string
}

type Repository interface {
	ListSyntheticCalcuttas(ctx context.Context, cohortID, calcuttaID *string, limit, offset int) ([]SyntheticCalcuttaListItem, error)
	GetSyntheticCalcutta(ctx context.Context, id string) (*SyntheticCalcuttaListItem, error)
	CreateSyntheticCalcutta(ctx context.Context, p CreateSyntheticCalcuttaParams) (string, error)
	PatchSyntheticCalcutta(ctx context.Context, id string, p PatchSyntheticCalcuttaParams) error

	ListSyntheticEntries(ctx context.Context, syntheticCalcuttaID string) ([]SyntheticEntryListItem, error)
	CreateSyntheticEntry(ctx context.Context, p CreateSyntheticEntryParams) (string, error)
	ImportSyntheticEntry(ctx context.Context, p ImportSyntheticEntryParams) (string, int, error)
	PatchSyntheticEntry(ctx context.Context, p PatchSyntheticEntryParams) error
	DeleteSyntheticEntry(ctx context.Context, p DeleteSyntheticEntryParams) error
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListSyntheticCalcuttas(ctx context.Context, cohortID, calcuttaID *string, limit, offset int) ([]SyntheticCalcuttaListItem, error) {
	return s.repo.ListSyntheticCalcuttas(ctx, cohortID, calcuttaID, limit, offset)
}

func (s *Service) GetSyntheticCalcutta(ctx context.Context, id string) (*SyntheticCalcuttaListItem, error) {
	return s.repo.GetSyntheticCalcutta(ctx, id)
}

func (s *Service) CreateSyntheticCalcutta(ctx context.Context, p CreateSyntheticCalcuttaParams) (string, error) {
	return s.repo.CreateSyntheticCalcutta(ctx, p)
}

func (s *Service) PatchSyntheticCalcutta(ctx context.Context, id string, p PatchSyntheticCalcuttaParams) error {
	return s.repo.PatchSyntheticCalcutta(ctx, id, p)
}

func (s *Service) ListSyntheticEntries(ctx context.Context, syntheticCalcuttaID string) ([]SyntheticEntryListItem, error) {
	return s.repo.ListSyntheticEntries(ctx, syntheticCalcuttaID)
}

func (s *Service) CreateSyntheticEntry(ctx context.Context, p CreateSyntheticEntryParams) (string, error) {
	return s.repo.CreateSyntheticEntry(ctx, p)
}

func (s *Service) ImportSyntheticEntry(ctx context.Context, p ImportSyntheticEntryParams) (string, int, error) {
	return s.repo.ImportSyntheticEntry(ctx, p)
}

func (s *Service) PatchSyntheticEntry(ctx context.Context, p PatchSyntheticEntryParams) error {
	return s.repo.PatchSyntheticEntry(ctx, p)
}

func (s *Service) DeleteSyntheticEntry(ctx context.Context, p DeleteSyntheticEntryParams) error {
	return s.repo.DeleteSyntheticEntry(ctx, p)
}

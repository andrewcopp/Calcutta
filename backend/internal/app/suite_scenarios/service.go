package suite_scenarios

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"time"
)

type SimulatedCalcutta struct {
	ID                        string
	Name                      string
	Description               *string
	TournamentID              string
	BaseCalcuttaID            *string
	StartingStateKey          string
	ExcludedEntryName         *string
	HighlightedSimulatedEntry *string
	Metadata                  json.RawMessage
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
}

type SimulatedCalcuttaPayout struct {
	Position    int
	AmountCents int
}

type SimulatedCalcuttaScoringRule struct {
	WinIndex      int
	PointsAwarded int
}

type CreateSimulatedCalcuttaParams struct {
	Name              string
	Description       *string
	TournamentID      string
	StartingStateKey  string
	ExcludedEntryName *string
	Metadata          json.RawMessage
	Payouts           []SimulatedCalcuttaPayout
	ScoringRules      []SimulatedCalcuttaScoringRule
}

type CreateSimulatedCalcuttaFromCalcuttaParams struct {
	CalcuttaID        string
	Name              *string
	Description       *string
	StartingStateKey  string
	ExcludedEntryName *string
	Metadata          json.RawMessage
}

type PatchSimulatedCalcuttaParams struct {
	Name                      *string
	Description               *string
	HighlightedSimulatedEntry *string
	Metadata                  *json.RawMessage
}

type ReplaceSimulatedCalcuttaRulesParams struct {
	Payouts      []SimulatedCalcuttaPayout
	ScoringRules []SimulatedCalcuttaScoringRule
}

type SimulatedEntryTeam struct {
	TeamID    string
	BidPoints int
}

type SimulatedEntry struct {
	ID                  string
	SimulatedCalcuttaID string
	DisplayName         string
	SourceKind          string
	SourceEntryID       *string
	SourceCandidateID   *string
	Teams               []SimulatedEntryTeam
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type CreateSimulatedEntryParams struct {
	SimulatedCalcuttaID string
	DisplayName         string
	Teams               []SimulatedEntryTeam
}

type PatchSimulatedEntryParams struct {
	SimulatedCalcuttaID string
	EntryID             string
	DisplayName         *string
	Teams               *[]SimulatedEntryTeam
}

type DeleteSimulatedEntryParams struct {
	SimulatedCalcuttaID string
	EntryID             string
}

type ImportCandidateAsSimulatedEntryParams struct {
	SimulatedCalcuttaID string
	CandidateID         string
	DisplayName         *string
}

type Repository interface {
	ListSimulatedCalcuttas(ctx context.Context, tournamentID *string, baseCalcuttaID *string, cohortID *string, limit, offset int) ([]SimulatedCalcutta, error)
	GetSimulatedCalcutta(ctx context.Context, id string) (*SimulatedCalcutta, []SimulatedCalcuttaPayout, []SimulatedCalcuttaScoringRule, error)
	CreateSimulatedCalcutta(ctx context.Context, p CreateSimulatedCalcuttaParams) (string, error)
	CreateSimulatedCalcuttaFromCalcutta(ctx context.Context, p CreateSimulatedCalcuttaFromCalcuttaParams) (string, int, error)
	PatchSimulatedCalcutta(ctx context.Context, id string, p PatchSimulatedCalcuttaParams) error
	ReplaceSimulatedCalcuttaRules(ctx context.Context, id string, p ReplaceSimulatedCalcuttaRulesParams) error

	ListSimulatedEntries(ctx context.Context, simulatedCalcuttaID string) (bool, []SimulatedEntry, error)
	CreateSimulatedEntry(ctx context.Context, p CreateSimulatedEntryParams) (string, error)
	PatchSimulatedEntry(ctx context.Context, p PatchSimulatedEntryParams) error
	DeleteSimulatedEntry(ctx context.Context, p DeleteSimulatedEntryParams) error
	ImportCandidateAsSimulatedEntry(ctx context.Context, p ImportCandidateAsSimulatedEntryParams) (string, int, error)
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListSimulatedCalcuttas(ctx context.Context, tournamentID *string, limit, offset int) ([]SimulatedCalcutta, error) {
	return s.repo.ListSimulatedCalcuttas(ctx, tournamentID, nil, nil, limit, offset)
}

func (s *Service) ListSimulatedCalcuttasWithFilters(ctx context.Context, tournamentID *string, baseCalcuttaID *string, cohortID *string, limit, offset int) ([]SimulatedCalcutta, error) {
	return s.repo.ListSimulatedCalcuttas(ctx, tournamentID, baseCalcuttaID, cohortID, limit, offset)
}

func (s *Service) GetSimulatedCalcutta(ctx context.Context, id string) (*SimulatedCalcutta, []SimulatedCalcuttaPayout, []SimulatedCalcuttaScoringRule, error) {
	return s.repo.GetSimulatedCalcutta(ctx, id)
}

func (s *Service) CreateSimulatedCalcutta(ctx context.Context, p CreateSimulatedCalcuttaParams) (string, error) {
	if len(p.Metadata) == 0 {
		p.Metadata = json.RawMessage([]byte("{}"))
	}
	return s.repo.CreateSimulatedCalcutta(ctx, p)
}

func (s *Service) CreateSimulatedCalcuttaFromCalcutta(ctx context.Context, p CreateSimulatedCalcuttaFromCalcuttaParams) (string, int, error) {
	if len(p.Metadata) == 0 {
		p.Metadata = json.RawMessage([]byte("{}"))
	}
	return s.repo.CreateSimulatedCalcuttaFromCalcutta(ctx, p)
}

func (s *Service) PatchSimulatedCalcutta(ctx context.Context, id string, p PatchSimulatedCalcuttaParams) error {
	return s.repo.PatchSimulatedCalcutta(ctx, id, p)
}

func (s *Service) ReplaceSimulatedCalcuttaRules(ctx context.Context, id string, p ReplaceSimulatedCalcuttaRulesParams) error {
	return s.repo.ReplaceSimulatedCalcuttaRules(ctx, id, p)
}

func (s *Service) ListSimulatedEntries(ctx context.Context, simulatedCalcuttaID string) (bool, []SimulatedEntry, error) {
	ok, items, err := s.repo.ListSimulatedEntries(ctx, simulatedCalcuttaID)
	if err != nil {
		return false, nil, err
	}
	for i := range items {
		sort.Slice(items[i].Teams, func(a, b int) bool { return items[i].Teams[a].BidPoints > items[i].Teams[b].BidPoints })
	}
	return ok, items, nil
}

func (s *Service) CreateSimulatedEntry(ctx context.Context, p CreateSimulatedEntryParams) (string, error) {
	for i := range p.Teams {
		p.Teams[i].TeamID = strings.TrimSpace(p.Teams[i].TeamID)
	}
	sort.Slice(p.Teams, func(i, j int) bool { return p.Teams[i].BidPoints > p.Teams[j].BidPoints })
	return s.repo.CreateSimulatedEntry(ctx, p)
}

func (s *Service) PatchSimulatedEntry(ctx context.Context, p PatchSimulatedEntryParams) error {
	if p.Teams != nil {
		teams := *p.Teams
		sort.Slice(teams, func(i, j int) bool { return teams[i].BidPoints > teams[j].BidPoints })
		p.Teams = &teams
	}
	return s.repo.PatchSimulatedEntry(ctx, p)
}

func (s *Service) DeleteSimulatedEntry(ctx context.Context, p DeleteSimulatedEntryParams) error {
	return s.repo.DeleteSimulatedEntry(ctx, p)
}

func (s *Service) ImportCandidateAsSimulatedEntry(ctx context.Context, p ImportCandidateAsSimulatedEntryParams) (string, int, error) {
	id, n, err := s.repo.ImportCandidateAsSimulatedEntry(ctx, p)
	if err != nil {
		return "", 0, err
	}
	return id, n, nil
}

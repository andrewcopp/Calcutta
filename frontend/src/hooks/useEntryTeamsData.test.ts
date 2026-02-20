import { describe, it, expect } from 'vitest';
import type {
  CalcuttaDashboard,
  CalcuttaEntry,
  CalcuttaEntryTeam,
  CalcuttaPortfolio,
  CalcuttaPortfolioTeam,
  TournamentTeam,
  Calcutta,
} from '../types/calcutta';
import type { EntryTeamsData } from './useEntryTeamsData';

// ---------------------------------------------------------------------------
// The hook is a single useMemo that transforms CalcuttaDashboard into
// EntryTeamsData for a given entryId. We extract the logic into a pure
// function to test without React rendering (the vitest env is node, not jsdom).
// ---------------------------------------------------------------------------

function computeEntryTeamsData(dashboardData: CalcuttaDashboard | undefined, entryId: string | undefined): EntryTeamsData {
  if (!dashboardData || !entryId) {
    return {
      calcuttaName: '',
      entryName: '',
      teams: [] as CalcuttaEntryTeam[],
      schools: [] as { id: string; name: string }[],
      portfolios: [] as CalcuttaPortfolio[],
      portfolioTeams: [] as CalcuttaPortfolioTeam[],
      tournamentTeams: dashboardData?.tournamentTeams ?? [],
      allEntryTeams: [] as CalcuttaEntryTeam[],
      allCalcuttaPortfolios: [] as (CalcuttaPortfolio & { entryName?: string })[],
      allCalcuttaPortfolioTeams: [] as CalcuttaPortfolioTeam[],
    };
  }

  const { calcutta, entries, entryTeams, portfolios, portfolioTeams, schools, tournamentTeams } = dashboardData;
  const schoolMap = new Map(schools.map((s) => [s.id, s]));
  const entryNameMap = new Map(entries.map((e) => [e.id, e.name]));

  const currentEntry = entries.find((e) => e.id === entryId);
  const entryName = currentEntry?.name || '';

  const thisEntryTeams = entryTeams.filter((et) => et.entryId === entryId);

  const teamsWithSchools: CalcuttaEntryTeam[] = thisEntryTeams.map((team) => ({
    ...team,
    team: team.team
      ? {
          ...team.team,
          school: schoolMap.get(team.team.schoolId),
        }
      : undefined,
  }));

  const thisEntryPortfolios = portfolios.filter((p) => p.entryId === entryId);

  const thisPortfolioIds = new Set(thisEntryPortfolios.map((p) => p.id));
  const thisPortfolioTeams: CalcuttaPortfolioTeam[] = portfolioTeams
    .filter((pt) => thisPortfolioIds.has(pt.portfolioId))
    .map((pt) => ({
      ...pt,
      team: pt.team
        ? {
            ...pt.team,
            school: schoolMap.get(pt.team.schoolId),
          }
        : undefined,
    }));

  const allPortfoliosWithNames: (CalcuttaPortfolio & { entryName?: string })[] = portfolios.map((p) => ({
    ...p,
    entryName: entryNameMap.get(p.entryId),
  }));

  const allPortfolioTeamsWithSchools: CalcuttaPortfolioTeam[] = portfolioTeams.map((pt) => ({
    ...pt,
    team: pt.team
      ? {
          ...pt.team,
          school: schoolMap.get(pt.team.schoolId),
        }
      : undefined,
  }));

  return {
    calcuttaName: calcutta.name,
    entryName,
    teams: teamsWithSchools,
    schools,
    portfolios: thisEntryPortfolios,
    portfolioTeams: thisPortfolioTeams,
    tournamentTeams,
    allEntryTeams: entryTeams,
    allCalcuttaPortfolios: allPortfoliosWithNames,
    allCalcuttaPortfolioTeams: allPortfolioTeamsWithSchools,
  };
}

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

function makeCalcutta(overrides: Partial<Calcutta> = {}): Calcutta {
  return {
    id: 'calc-1',
    name: 'March Madness 2026',
    tournamentId: 'tourn-1',
    ownerId: 'owner-1',
    minTeams: 3,
    maxTeams: 10,
    maxBid: 50,
    budgetPoints: 100,
    created: '2026-01-01T00:00:00Z',
    updated: '2026-01-01T00:00:00Z',
    ...overrides,
  };
}

function makeEntry(overrides: Partial<CalcuttaEntry> & { id: string }): CalcuttaEntry {
  return {
    name: 'Entry',
    calcuttaId: 'calc-1',
    created: '2026-01-01T00:00:00Z',
    updated: '2026-01-01T00:00:00Z',
    ...overrides,
  };
}

function makeEntryTeam(overrides: Partial<CalcuttaEntryTeam> & { id: string; entryId: string; teamId: string }): CalcuttaEntryTeam {
  return {
    bid: 10,
    created: '2026-01-01T00:00:00Z',
    updated: '2026-01-01T00:00:00Z',
    ...overrides,
  };
}

function makePortfolio(overrides: Partial<CalcuttaPortfolio> & { id: string; entryId: string }): CalcuttaPortfolio {
  return {
    maximumPoints: 100,
    created: '2026-01-01T00:00:00Z',
    updated: '2026-01-01T00:00:00Z',
    ...overrides,
  };
}

function makePortfolioTeam(
  overrides: Partial<CalcuttaPortfolioTeam> & { id: string; portfolioId: string; teamId: string },
): CalcuttaPortfolioTeam {
  return {
    ownershipPercentage: 1,
    actualPoints: 0,
    expectedPoints: 0,
    created: '2026-01-01T00:00:00Z',
    updated: '2026-01-01T00:00:00Z',
    ...overrides,
  };
}

function makeTournamentTeam(overrides: Partial<TournamentTeam> & { id: string; schoolId: string }): TournamentTeam {
  return {
    tournamentId: 'tourn-1',
    seed: 1,
    region: 'East',
    byes: 0,
    wins: 0,
    eliminated: false,
    created: '2026-01-01T00:00:00Z',
    updated: '2026-01-01T00:00:00Z',
    ...overrides,
  };
}

function makeDashboard(overrides: Partial<CalcuttaDashboard> = {}): CalcuttaDashboard {
  return {
    calcutta: makeCalcutta(),
    entries: [],
    entryTeams: [],
    portfolios: [],
    portfolioTeams: [],
    schools: [],
    tournamentTeams: [],
    ...overrides,
  };
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe('useEntryTeamsData (pure transformation)', () => {
  describe('when dashboard is undefined', () => {
    it('returns empty calcuttaName', () => {
      // GIVEN no dashboard data
      // WHEN computing entry teams data
      const result = computeEntryTeamsData(undefined, 'e1');

      // THEN calcuttaName is empty
      expect(result.calcuttaName).toBe('');
    });

    it('returns empty entryName', () => {
      // GIVEN no dashboard data
      // WHEN computing entry teams data
      const result = computeEntryTeamsData(undefined, 'e1');

      // THEN entryName is empty
      expect(result.entryName).toBe('');
    });

    it('returns empty teams array', () => {
      // GIVEN no dashboard data
      // WHEN computing entry teams data
      const result = computeEntryTeamsData(undefined, 'e1');

      // THEN teams is empty
      expect(result.teams).toEqual([]);
    });

    it('returns empty portfolios array', () => {
      // GIVEN no dashboard data
      // WHEN computing entry teams data
      const result = computeEntryTeamsData(undefined, 'e1');

      // THEN portfolios is empty
      expect(result.portfolios).toEqual([]);
    });

    it('returns empty allEntryTeams array', () => {
      // GIVEN no dashboard data
      // WHEN computing entry teams data
      const result = computeEntryTeamsData(undefined, 'e1');

      // THEN allEntryTeams is empty
      expect(result.allEntryTeams).toEqual([]);
    });
  });

  describe('when entryId is undefined', () => {
    it('returns empty entryName', () => {
      // GIVEN a valid dashboard but no entryId
      const dashboard = makeDashboard();

      // WHEN computing entry teams data with undefined entryId
      const result = computeEntryTeamsData(dashboard, undefined);

      // THEN entryName is empty
      expect(result.entryName).toBe('');
    });

    it('still returns tournamentTeams from dashboard', () => {
      // GIVEN a dashboard with tournament teams but no entryId
      const teams = [makeTournamentTeam({ id: 't1', schoolId: 's1' })];
      const dashboard = makeDashboard({ tournamentTeams: teams });

      // WHEN computing entry teams data with undefined entryId
      const result = computeEntryTeamsData(dashboard, undefined);

      // THEN tournamentTeams from the dashboard are still returned
      expect(result.tournamentTeams).toHaveLength(1);
    });
  });

  describe('when entryId does not match any entry', () => {
    it('returns empty entryName for non-existent entry', () => {
      // GIVEN a dashboard with entry "e1" but we query for "e999"
      const dashboard = makeDashboard({
        entries: [makeEntry({ id: 'e1', name: 'Alice' })],
      });

      // WHEN computing entry teams data for a non-existent entry
      const result = computeEntryTeamsData(dashboard, 'e999');

      // THEN entryName is empty
      expect(result.entryName).toBe('');
    });

    it('returns empty teams for non-existent entry', () => {
      // GIVEN a dashboard with entry teams for "e1" but we query for "e999"
      const dashboard = makeDashboard({
        entries: [makeEntry({ id: 'e1', name: 'Alice' })],
        entryTeams: [makeEntryTeam({ id: 'et1', entryId: 'e1', teamId: 't1' })],
      });

      // WHEN computing entry teams data for a non-existent entry
      const result = computeEntryTeamsData(dashboard, 'e999');

      // THEN filtered teams is empty
      expect(result.teams).toEqual([]);
    });

    it('returns empty portfolios for non-existent entry', () => {
      // GIVEN a dashboard with portfolios for "e1" but we query for "e999"
      const dashboard = makeDashboard({
        entries: [makeEntry({ id: 'e1', name: 'Alice' })],
        portfolios: [makePortfolio({ id: 'p1', entryId: 'e1' })],
      });

      // WHEN computing entry teams data for a non-existent entry
      const result = computeEntryTeamsData(dashboard, 'e999');

      // THEN filtered portfolios is empty
      expect(result.portfolios).toEqual([]);
    });
  });

  describe('filtering entry teams for the given entry', () => {
    it('returns only teams belonging to the specified entry', () => {
      // GIVEN entry teams for two different entries
      const dashboard = makeDashboard({
        entries: [
          makeEntry({ id: 'e1', name: 'Alice' }),
          makeEntry({ id: 'e2', name: 'Bob' }),
        ],
        entryTeams: [
          makeEntryTeam({ id: 'et1', entryId: 'e1', teamId: 't1' }),
          makeEntryTeam({ id: 'et2', entryId: 'e1', teamId: 't2' }),
          makeEntryTeam({ id: 'et3', entryId: 'e2', teamId: 't3' }),
        ],
      });

      // WHEN computing entry teams data for "e1"
      const result = computeEntryTeamsData(dashboard, 'e1');

      // THEN only entry e1's teams are returned
      expect(result.teams).toHaveLength(2);
    });

    it('includes all entry teams in allEntryTeams regardless of entryId filter', () => {
      // GIVEN entry teams for two entries
      const dashboard = makeDashboard({
        entries: [makeEntry({ id: 'e1', name: 'Alice' })],
        entryTeams: [
          makeEntryTeam({ id: 'et1', entryId: 'e1', teamId: 't1' }),
          makeEntryTeam({ id: 'et2', entryId: 'e2', teamId: 't2' }),
        ],
      });

      // WHEN computing entry teams data for "e1"
      const result = computeEntryTeamsData(dashboard, 'e1');

      // THEN allEntryTeams contains all entry teams (unfiltered)
      expect(result.allEntryTeams).toHaveLength(2);
    });
  });

  describe('enriching entry teams with school info', () => {
    it('attaches school data to entry teams via schoolMap', () => {
      // GIVEN an entry team referencing a team with schoolId "s1"
      const dashboard = makeDashboard({
        entries: [makeEntry({ id: 'e1', name: 'Alice' })],
        schools: [{ id: 's1', name: 'Duke' }],
        entryTeams: [
          makeEntryTeam({
            id: 'et1',
            entryId: 'e1',
            teamId: 't1',
            team: { id: 't1', schoolId: 's1', seed: 1 },
          }),
        ],
      });

      // WHEN computing entry teams data
      const result = computeEntryTeamsData(dashboard, 'e1');

      // THEN the team has school info attached
      expect(result.teams[0].team?.school?.name).toBe('Duke');
    });

    it('handles entry team with no nested team gracefully', () => {
      // GIVEN an entry team with no nested team object
      const dashboard = makeDashboard({
        entries: [makeEntry({ id: 'e1', name: 'Alice' })],
        entryTeams: [
          makeEntryTeam({ id: 'et1', entryId: 'e1', teamId: 't1', team: undefined }),
        ],
      });

      // WHEN computing entry teams data
      const result = computeEntryTeamsData(dashboard, 'e1');

      // THEN the team field is undefined (no crash)
      expect(result.teams[0].team).toBeUndefined();
    });
  });

  describe('filtering portfolios for the given entry', () => {
    it('returns only portfolios belonging to the specified entry', () => {
      // GIVEN portfolios for two entries
      const dashboard = makeDashboard({
        entries: [makeEntry({ id: 'e1', name: 'Alice' })],
        portfolios: [
          makePortfolio({ id: 'p1', entryId: 'e1' }),
          makePortfolio({ id: 'p2', entryId: 'e2' }),
        ],
      });

      // WHEN computing entry teams data for "e1"
      const result = computeEntryTeamsData(dashboard, 'e1');

      // THEN only e1's portfolio is returned
      expect(result.portfolios).toHaveLength(1);
    });

    it('filters portfolio teams to those belonging to the entry portfolios', () => {
      // GIVEN portfolio teams across multiple entries
      const dashboard = makeDashboard({
        entries: [makeEntry({ id: 'e1', name: 'Alice' })],
        portfolios: [
          makePortfolio({ id: 'p1', entryId: 'e1' }),
          makePortfolio({ id: 'p2', entryId: 'e2' }),
        ],
        portfolioTeams: [
          makePortfolioTeam({ id: 'pt1', portfolioId: 'p1', teamId: 't1' }),
          makePortfolioTeam({ id: 'pt2', portfolioId: 'p2', teamId: 't2' }),
        ],
      });

      // WHEN computing entry teams data for "e1"
      const result = computeEntryTeamsData(dashboard, 'e1');

      // THEN only portfolio teams from e1's portfolios are returned
      expect(result.portfolioTeams).toHaveLength(1);
    });
  });

  describe('calcutta metadata', () => {
    it('returns calcuttaName from the dashboard calcutta', () => {
      // GIVEN a dashboard with a named calcutta
      const dashboard = makeDashboard({
        calcutta: makeCalcutta({ name: 'Championship Pool 2026' }),
        entries: [makeEntry({ id: 'e1', name: 'Alice' })],
      });

      // WHEN computing entry teams data
      const result = computeEntryTeamsData(dashboard, 'e1');

      // THEN calcuttaName matches the calcutta name
      expect(result.calcuttaName).toBe('Championship Pool 2026');
    });

    it('returns the matching entry name', () => {
      // GIVEN a dashboard with an entry named "Alice"
      const dashboard = makeDashboard({
        entries: [makeEntry({ id: 'e1', name: 'Alice' })],
      });

      // WHEN computing entry teams data for "e1"
      const result = computeEntryTeamsData(dashboard, 'e1');

      // THEN entryName matches
      expect(result.entryName).toBe('Alice');
    });
  });

  describe('all-portfolios enrichment', () => {
    it('enriches all portfolios with entry names', () => {
      // GIVEN portfolios belonging to different entries
      const dashboard = makeDashboard({
        entries: [
          makeEntry({ id: 'e1', name: 'Alice' }),
          makeEntry({ id: 'e2', name: 'Bob' }),
        ],
        portfolios: [
          makePortfolio({ id: 'p1', entryId: 'e1' }),
          makePortfolio({ id: 'p2', entryId: 'e2' }),
        ],
      });

      // WHEN computing entry teams data for "e1"
      const result = computeEntryTeamsData(dashboard, 'e1');

      // THEN allCalcuttaPortfolios has entry names for both portfolios
      expect(result.allCalcuttaPortfolios.map((p) => p.entryName)).toEqual(['Alice', 'Bob']);
    });

    it('enriches all portfolio teams with school info', () => {
      // GIVEN portfolio teams referencing teams with schools
      const dashboard = makeDashboard({
        entries: [makeEntry({ id: 'e1', name: 'Alice' })],
        schools: [{ id: 's1', name: 'Duke' }],
        portfolioTeams: [
          makePortfolioTeam({
            id: 'pt1',
            portfolioId: 'p1',
            teamId: 't1',
            team: { id: 't1', schoolId: 's1' },
          }),
        ],
      });

      // WHEN computing entry teams data
      const result = computeEntryTeamsData(dashboard, 'e1');

      // THEN allCalcuttaPortfolioTeams have school info attached
      expect(result.allCalcuttaPortfolioTeams[0].team?.school?.name).toBe('Duke');
    });
  });
});

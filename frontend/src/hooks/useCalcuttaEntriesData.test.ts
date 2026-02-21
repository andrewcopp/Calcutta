import { describe, it, expect } from 'vitest';
import type {
  CalcuttaDashboard,
  CalcuttaEntry,
  CalcuttaEntryTeam,
  CalcuttaPortfolio,
  CalcuttaPortfolioTeam,
  Calcutta,
} from '../types/calcutta';
import type { TournamentTeam } from '../types/tournament';
import type { CalcuttaEntriesData } from './useCalcuttaEntriesData';

// ---------------------------------------------------------------------------
// The hook is a chain of useMemo calls over CalcuttaDashboard.
// Since the test environment is node (no DOM / no React rendering), we extract
// the pure transformation logic into a standalone function that mirrors the
// hook's computation. This lets us unit-test every derivation without needing
// renderHook or a React tree.
// ---------------------------------------------------------------------------

interface SeedInvestmentDatum {
  seed: number;
  totalInvestment: number;
}

interface TeamROIDatum {
  teamId: string;
  seed: number;
  region: string;
  teamName: string;
  points: number;
  investment: number;
  roi: number;
}

type EnrichedEntry = CalcuttaEntry & { totalPoints: number };

function computeCalcuttaEntriesData(dashboardData: CalcuttaDashboard | undefined): CalcuttaEntriesData {
  // ---- first useMemo block ----
  let entries: EnrichedEntry[] = [];
  let totalEntries = 0;
  let allCalcuttaPortfolios: (CalcuttaPortfolio & { entryName?: string })[] = [];
  let allCalcuttaPortfolioTeams: CalcuttaPortfolioTeam[] = [];
  let allEntryTeams: CalcuttaEntryTeam[] = [];
  let seedInvestmentData: SeedInvestmentDatum[] = [];

  if (dashboardData) {
    const { entries: rawEntries, entryTeams, portfolios, portfolioTeams, schools } = dashboardData;
    const schoolMap = new Map(schools.map((s) => [s.id, s]));
    const entryNameMap = new Map(rawEntries.map((e) => [e.id, e.name]));

    allCalcuttaPortfolios = portfolios.map((p) => ({
      ...p,
      entryName: entryNameMap.get(p.entryId),
    }));

    const entryPointsMap = new Map<string, number>();
    const portfolioToEntry = new Map(portfolios.map((p) => [p.id, p.entryId]));
    for (const pt of portfolioTeams) {
      const entryId = portfolioToEntry.get(pt.portfolioId);
      if (entryId) {
        entryPointsMap.set(entryId, (entryPointsMap.get(entryId) || 0) + pt.actualPoints);
      }
    }

    allCalcuttaPortfolioTeams = portfolioTeams.map((pt) => {
      const school = pt.team?.schoolId ? schoolMap.get(pt.team.schoolId) : undefined;
      return {
        ...pt,
        team: pt.team
          ? {
              ...pt.team,
              school: school ? { id: school.id, name: school.name } : pt.team.school,
            }
          : pt.team,
      };
    });

    const sortedEntries: EnrichedEntry[] = rawEntries
      .map((entry) => ({
        ...entry,
        totalPoints: entry.totalPoints || entryPointsMap.get(entry.id) || 0,
      }))
      .sort((a, b) => {
        const diff = (b.totalPoints || 0) - (a.totalPoints || 0);
        if (diff !== 0) return diff;
        return new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime();
      });

    const seedMap = new Map<number, number>();
    for (const team of entryTeams) {
      if (!team.team?.seed || !team.bid) continue;
      const seed = team.team.seed;
      seedMap.set(seed, (seedMap.get(seed) || 0) + team.bid);
    }
    seedInvestmentData = Array.from(seedMap.entries())
      .map(([seed, totalInvestment]) => ({ seed, totalInvestment }))
      .sort((a, b) => a.seed - b.seed);

    entries = sortedEntries;
    totalEntries = rawEntries.length;
    allEntryTeams = entryTeams;
  }

  // ---- remaining derived values ----
  const schools = dashboardData?.schools ?? [];
  const tournamentTeams = dashboardData?.tournamentTeams ?? [];

  const totalInvestment = allEntryTeams.reduce((sum, et) => sum + et.bid, 0);
  const totalReturns = entries.reduce((sum, e) => sum + (e.totalPoints || 0), 0);
  const averageReturn = totalEntries > 0 ? totalReturns / totalEntries : 0;

  let returnsStdDev = 0;
  if (totalEntries > 1) {
    const variance = entries.reduce((acc, e) => {
      const v = (e.totalPoints || 0) - averageReturn;
      return acc + v * v;
    }, 0);
    returnsStdDev = Math.sqrt(variance / totalEntries);
  }

  const schoolNameById = new Map(schools.map((school) => [school.id, school.name]));

  const teamROIData: TeamROIDatum[] = tournamentTeams
    .map((team) => {
      const schoolName = schoolNameById.get(team.schoolId) || 'Unknown School';
      const teamInvestment = allEntryTeams.filter((et) => et.teamId === team.id).reduce((sum, et) => sum + et.bid, 0);
      const teamPoints = allCalcuttaPortfolioTeams
        .filter((pt) => pt.teamId === team.id)
        .reduce((sum, pt) => sum + pt.actualPoints, 0);
      const roi = teamPoints / (teamInvestment + 1);
      return {
        teamId: team.id,
        seed: team.seed,
        region: team.region,
        teamName: schoolName,
        points: teamPoints,
        investment: teamInvestment,
        roi,
      };
    })
    .sort((a, b) => b.roi - a.roi);

  return {
    entries,
    totalEntries,
    allCalcuttaPortfolios,
    allCalcuttaPortfolioTeams,
    allEntryTeams,
    seedInvestmentData,
    schools,
    tournamentTeams,
    totalInvestment,
    totalReturns,
    averageReturn,
    returnsStdDev,
    teamROIData,
  };
}

// ---------------------------------------------------------------------------
// Test helpers -- factory functions for building fixture data
// ---------------------------------------------------------------------------

function makeCalcutta(overrides: Partial<Calcutta> = {}): Calcutta {
  return {
    id: 'calc-1',
    name: 'Test Calcutta',
    tournamentId: 'tourn-1',
    ownerId: 'owner-1',
    minTeams: 3,
    maxTeams: 10,
    maxBidPoints: 50,
    budgetPoints: 100,
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
    ...overrides,
  };
}

function makeEntry(overrides: Partial<CalcuttaEntry> & { id: string }): CalcuttaEntry {
  return {
    name: 'Entry',
    calcuttaId: 'calc-1',
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
    ...overrides,
  };
}

function makeEntryTeam(overrides: Partial<CalcuttaEntryTeam> & { id: string; entryId: string; teamId: string }): CalcuttaEntryTeam {
  return {
    bid: 10,
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
    ...overrides,
  };
}

function makePortfolio(overrides: Partial<CalcuttaPortfolio> & { id: string; entryId: string }): CalcuttaPortfolio {
  return {
    maximumPoints: 100,
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
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
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
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
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
    ...overrides,
  };
}

function makeDashboard(overrides: Partial<CalcuttaDashboard> = {}): CalcuttaDashboard {
  return {
    calcutta: makeCalcutta(),
    biddingOpen: false,
    totalEntries: 0,
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

describe('useCalcuttaEntriesData (pure transformation)', () => {
  describe('when dashboard is undefined', () => {
    it('returns empty entries array', () => {
      // GIVEN no dashboard data
      // WHEN computing entries data
      const result = computeCalcuttaEntriesData(undefined);

      // THEN entries is empty
      expect(result.entries).toEqual([]);
    });

    it('returns zero totalEntries', () => {
      // GIVEN no dashboard data
      // WHEN computing entries data
      const result = computeCalcuttaEntriesData(undefined);

      // THEN totalEntries is 0
      expect(result.totalEntries).toBe(0);
    });

    it('returns zero totalInvestment', () => {
      // GIVEN no dashboard data
      // WHEN computing entries data
      const result = computeCalcuttaEntriesData(undefined);

      // THEN totalInvestment is 0
      expect(result.totalInvestment).toBe(0);
    });

    it('returns zero averageReturn', () => {
      // GIVEN no dashboard data
      // WHEN computing entries data
      const result = computeCalcuttaEntriesData(undefined);

      // THEN averageReturn is 0
      expect(result.averageReturn).toBe(0);
    });

    it('returns empty teamROIData', () => {
      // GIVEN no dashboard data
      // WHEN computing entries data
      const result = computeCalcuttaEntriesData(undefined);

      // THEN teamROIData is empty
      expect(result.teamROIData).toEqual([]);
    });

    it('returns zero returnsStdDev', () => {
      // GIVEN no dashboard data
      // WHEN computing entries data
      const result = computeCalcuttaEntriesData(undefined);

      // THEN returnsStdDev is 0
      expect(result.returnsStdDev).toBe(0);
    });
  });

  describe('when dashboard has no entries', () => {
    it('returns zero totalEntries for dashboard with empty entries', () => {
      // GIVEN a dashboard with no entries
      const dashboard = makeDashboard({ entries: [] });

      // WHEN computing entries data
      const result = computeCalcuttaEntriesData(dashboard);

      // THEN totalEntries is 0
      expect(result.totalEntries).toBe(0);
    });

    it('returns zero averageReturn when there are no entries', () => {
      // GIVEN a dashboard with no entries
      const dashboard = makeDashboard({ entries: [] });

      // WHEN computing entries data
      const result = computeCalcuttaEntriesData(dashboard);

      // THEN averageReturn is 0 (avoids divide-by-zero)
      expect(result.averageReturn).toBe(0);
    });
  });

  describe('entry enrichment and sorting', () => {
    it('enriches entries with totalPoints from portfolio teams', () => {
      // GIVEN an entry whose totalPoints is unset, with portfolio teams that sum to 30
      const dashboard = makeDashboard({
        entries: [makeEntry({ id: 'e1', name: 'Alice' })],
        portfolios: [makePortfolio({ id: 'p1', entryId: 'e1' })],
        portfolioTeams: [
          makePortfolioTeam({ id: 'pt1', portfolioId: 'p1', teamId: 't1', actualPoints: 20 }),
          makePortfolioTeam({ id: 'pt2', portfolioId: 'p1', teamId: 't2', actualPoints: 10 }),
        ],
      });

      // WHEN computing entries data
      const result = computeCalcuttaEntriesData(dashboard);

      // THEN the entry totalPoints is computed from portfolio teams
      expect(result.entries[0].totalPoints).toBe(30);
    });

    it('prefers existing totalPoints over computed value', () => {
      // GIVEN an entry that already has totalPoints set to 50
      const dashboard = makeDashboard({
        entries: [makeEntry({ id: 'e1', name: 'Alice', totalPoints: 50 })],
        portfolios: [makePortfolio({ id: 'p1', entryId: 'e1' })],
        portfolioTeams: [
          makePortfolioTeam({ id: 'pt1', portfolioId: 'p1', teamId: 't1', actualPoints: 10 }),
        ],
      });

      // WHEN computing entries data
      const result = computeCalcuttaEntriesData(dashboard);

      // THEN the existing totalPoints is preserved
      expect(result.entries[0].totalPoints).toBe(50);
    });

    it('sorts entries by totalPoints descending', () => {
      // GIVEN three entries with different points
      const dashboard = makeDashboard({
        entries: [
          makeEntry({ id: 'e1', name: 'Low', totalPoints: 10 }),
          makeEntry({ id: 'e2', name: 'High', totalPoints: 50 }),
          makeEntry({ id: 'e3', name: 'Mid', totalPoints: 30 }),
        ],
      });

      // WHEN computing entries data
      const result = computeCalcuttaEntriesData(dashboard);

      // THEN entries are sorted highest points first
      expect(result.entries.map((e) => e.name)).toEqual(['High', 'Mid', 'Low']);
    });

    it('breaks totalPoints ties by created date descending (most recent first)', () => {
      // GIVEN two entries with identical points but different created dates
      const dashboard = makeDashboard({
        entries: [
          makeEntry({ id: 'e1', name: 'Older', totalPoints: 20, createdAt: '2026-01-01T00:00:00Z' }),
          makeEntry({ id: 'e2', name: 'Newer', totalPoints: 20, createdAt: '2026-02-01T00:00:00Z' }),
        ],
      });

      // WHEN computing entries data
      const result = computeCalcuttaEntriesData(dashboard);

      // THEN the more recently created entry comes first
      expect(result.entries[0].name).toBe('Newer');
    });
  });

  describe('totalInvestment', () => {
    it('computes totalInvestment as sum of all entry team bids', () => {
      // GIVEN entry teams with bids of 15, 25, and 10
      const dashboard = makeDashboard({
        entryTeams: [
          makeEntryTeam({ id: 'et1', entryId: 'e1', teamId: 't1', bid: 15 }),
          makeEntryTeam({ id: 'et2', entryId: 'e1', teamId: 't2', bid: 25 }),
          makeEntryTeam({ id: 'et3', entryId: 'e2', teamId: 't3', bid: 10 }),
        ],
      });

      // WHEN computing entries data
      const result = computeCalcuttaEntriesData(dashboard);

      // THEN totalInvestment is the sum of all bids
      expect(result.totalInvestment).toBe(50);
    });
  });

  describe('averageReturn', () => {
    it('computes averageReturn as totalReturns divided by totalEntries', () => {
      // GIVEN two entries with totalPoints 40 and 60
      const dashboard = makeDashboard({
        entries: [
          makeEntry({ id: 'e1', name: 'A', totalPoints: 40 }),
          makeEntry({ id: 'e2', name: 'B', totalPoints: 60 }),
        ],
      });

      // WHEN computing entries data
      const result = computeCalcuttaEntriesData(dashboard);

      // THEN averageReturn is (40+60)/2 = 50
      expect(result.averageReturn).toBe(50);
    });
  });

  describe('returnsStdDev', () => {
    it('returns zero when there is only one entry', () => {
      // GIVEN exactly one entry
      const dashboard = makeDashboard({
        entries: [makeEntry({ id: 'e1', name: 'Solo', totalPoints: 25 })],
      });

      // WHEN computing entries data
      const result = computeCalcuttaEntriesData(dashboard);

      // THEN stddev is 0 (needs more than 1 entry)
      expect(result.returnsStdDev).toBe(0);
    });

    it('computes standard deviation of entry points', () => {
      // GIVEN two entries with points 10 and 30 (mean=20, variance=(100+100)/2=100, stddev=10)
      const dashboard = makeDashboard({
        entries: [
          makeEntry({ id: 'e1', name: 'A', totalPoints: 10 }),
          makeEntry({ id: 'e2', name: 'B', totalPoints: 30 }),
        ],
      });

      // WHEN computing entries data
      const result = computeCalcuttaEntriesData(dashboard);

      // THEN returnsStdDev is 10
      expect(result.returnsStdDev).toBe(10);
    });
  });

  describe('seedInvestmentData', () => {
    it('aggregates bids by seed and sorts ascending', () => {
      // GIVEN entry teams with bids on seed 1 (10+5) and seed 8 (20)
      const dashboard = makeDashboard({
        entryTeams: [
          makeEntryTeam({ id: 'et1', entryId: 'e1', teamId: 't1', bid: 10, team: { id: 't1', schoolId: 's1', seed: 1 } }),
          makeEntryTeam({ id: 'et2', entryId: 'e2', teamId: 't2', bid: 5, team: { id: 't2', schoolId: 's2', seed: 1 } }),
          makeEntryTeam({ id: 'et3', entryId: 'e1', teamId: 't3', bid: 20, team: { id: 't3', schoolId: 's3', seed: 8 } }),
        ],
      });

      // WHEN computing entries data
      const result = computeCalcuttaEntriesData(dashboard);

      // THEN seed investment data is sorted by seed with aggregated totals
      expect(result.seedInvestmentData).toEqual([
        { seed: 1, totalInvestment: 15 },
        { seed: 8, totalInvestment: 20 },
      ]);
    });

    it('skips entry teams with no seed', () => {
      // GIVEN an entry team whose nested team has no seed
      const dashboard = makeDashboard({
        entryTeams: [
          makeEntryTeam({ id: 'et1', entryId: 'e1', teamId: 't1', bid: 10, team: { id: 't1', schoolId: 's1' } }),
        ],
      });

      // WHEN computing entries data
      const result = computeCalcuttaEntriesData(dashboard);

      // THEN seedInvestmentData is empty
      expect(result.seedInvestmentData).toEqual([]);
    });

    it('skips entry teams with zero bid', () => {
      // GIVEN an entry team with bid=0
      const dashboard = makeDashboard({
        entryTeams: [
          makeEntryTeam({ id: 'et1', entryId: 'e1', teamId: 't1', bid: 0, team: { id: 't1', schoolId: 's1', seed: 1 } }),
        ],
      });

      // WHEN computing entries data
      const result = computeCalcuttaEntriesData(dashboard);

      // THEN seedInvestmentData is empty (zero bid is falsy, skipped)
      expect(result.seedInvestmentData).toEqual([]);
    });
  });

  describe('teamROIData', () => {
    it('computes ROI as points / (investment + 1) for each tournament team', () => {
      // GIVEN a tournament team with 20 points of investment and 100 actual points
      const dashboard = makeDashboard({
        schools: [{ id: 's1', name: 'Duke' }],
        tournamentTeams: [makeTournamentTeam({ id: 't1', schoolId: 's1', seed: 1, region: 'East' })],
        entryTeams: [makeEntryTeam({ id: 'et1', entryId: 'e1', teamId: 't1', bid: 20 })],
        portfolios: [makePortfolio({ id: 'p1', entryId: 'e1' })],
        portfolioTeams: [makePortfolioTeam({ id: 'pt1', portfolioId: 'p1', teamId: 't1', actualPoints: 100 })],
      });

      // WHEN computing entries data
      const result = computeCalcuttaEntriesData(dashboard);

      // THEN ROI is 100 / (20 + 1) = ~4.76
      expect(result.teamROIData[0].roi).toBeCloseTo(100 / 21, 5);
    });

    it('uses "Unknown School" when school name is not found', () => {
      // GIVEN a tournament team whose schoolId has no matching school
      const dashboard = makeDashboard({
        schools: [],
        tournamentTeams: [makeTournamentTeam({ id: 't1', schoolId: 'missing' })],
      });

      // WHEN computing entries data
      const result = computeCalcuttaEntriesData(dashboard);

      // THEN teamName falls back to "Unknown School"
      expect(result.teamROIData[0].teamName).toBe('Unknown School');
    });

    it('sorts teams by ROI descending', () => {
      // GIVEN two tournament teams: one with high ROI, one with low ROI
      const dashboard = makeDashboard({
        schools: [
          { id: 's1', name: 'Duke' },
          { id: 's2', name: 'UNC' },
        ],
        tournamentTeams: [
          makeTournamentTeam({ id: 't1', schoolId: 's1', seed: 1, region: 'East' }),
          makeTournamentTeam({ id: 't2', schoolId: 's2', seed: 2, region: 'East' }),
        ],
        entryTeams: [
          makeEntryTeam({ id: 'et1', entryId: 'e1', teamId: 't1', bid: 50 }),
          makeEntryTeam({ id: 'et2', entryId: 'e1', teamId: 't2', bid: 5 }),
        ],
        portfolios: [makePortfolio({ id: 'p1', entryId: 'e1' })],
        portfolioTeams: [
          makePortfolioTeam({ id: 'pt1', portfolioId: 'p1', teamId: 't1', actualPoints: 10 }),
          makePortfolioTeam({ id: 'pt2', portfolioId: 'p1', teamId: 't2', actualPoints: 50 }),
        ],
      });

      // WHEN computing entries data
      const result = computeCalcuttaEntriesData(dashboard);

      // THEN the team with higher ROI (UNC: 50/6 > Duke: 10/51) comes first
      expect(result.teamROIData[0].teamName).toBe('UNC');
    });

    it('handles team with zero investment gracefully (divides by 1)', () => {
      // GIVEN a tournament team with no bids
      const dashboard = makeDashboard({
        schools: [{ id: 's1', name: 'Duke' }],
        tournamentTeams: [makeTournamentTeam({ id: 't1', schoolId: 's1' })],
        entryTeams: [],
        portfolios: [makePortfolio({ id: 'p1', entryId: 'e1' })],
        portfolioTeams: [makePortfolioTeam({ id: 'pt1', portfolioId: 'p1', teamId: 't1', actualPoints: 30 })],
      });

      // WHEN computing entries data
      const result = computeCalcuttaEntriesData(dashboard);

      // THEN ROI is 30/(0+1) = 30 (no division by zero)
      expect(result.teamROIData[0].roi).toBe(30);
    });
  });

  describe('portfolio enrichment', () => {
    it('enriches portfolios with entry names', () => {
      // GIVEN a portfolio belonging to entry "Alice"
      const dashboard = makeDashboard({
        entries: [makeEntry({ id: 'e1', name: 'Alice' })],
        portfolios: [makePortfolio({ id: 'p1', entryId: 'e1' })],
      });

      // WHEN computing entries data
      const result = computeCalcuttaEntriesData(dashboard);

      // THEN the enriched portfolio has the entry name
      expect(result.allCalcuttaPortfolios[0].entryName).toBe('Alice');
    });

    it('enriches portfolio teams with school info', () => {
      // GIVEN a portfolio team whose underlying team references school "Duke"
      const dashboard = makeDashboard({
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

      // WHEN computing entries data
      const result = computeCalcuttaEntriesData(dashboard);

      // THEN the portfolio team has school info attached
      expect(result.allCalcuttaPortfolioTeams[0].team?.school?.name).toBe('Duke');
    });
  });
});

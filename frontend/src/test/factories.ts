import type {
  Calcutta,
  CalcuttaDashboard,
  CalcuttaEntry,
  CalcuttaEntryTeam,
  CalcuttaPortfolio,
  CalcuttaPortfolioTeam,
} from '../types/calcutta';
import type { TournamentTeam } from '../types/tournament';

export function makeCalcutta(overrides: Partial<Calcutta> = {}): Calcutta {
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

export function makeEntry(overrides: Partial<CalcuttaEntry> & { id: string }): CalcuttaEntry {
  return {
    name: 'Entry',
    calcuttaId: 'calc-1',
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
    ...overrides,
  };
}

export function makeEntryTeam(
  overrides: Partial<CalcuttaEntryTeam> & { id: string; entryId: string; teamId: string },
): CalcuttaEntryTeam {
  return {
    bid: 10,
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
    ...overrides,
  };
}

export function makePortfolio(
  overrides: Partial<CalcuttaPortfolio> & { id: string; entryId: string },
): CalcuttaPortfolio {
  return {
    maximumPoints: 100,
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
    ...overrides,
  };
}

export function makePortfolioTeam(
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

export function makeTournamentTeam(
  overrides: Partial<TournamentTeam> & { id: string; schoolId: string },
): TournamentTeam {
  return {
    tournamentId: 'tourn-1',
    seed: 1,
    region: 'East',
    byes: 0,
    wins: 0,
    isEliminated: false,
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
    ...overrides,
  };
}

export function makeDashboard(overrides: Partial<CalcuttaDashboard> = {}): CalcuttaDashboard {
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

import type {
  Pool,
  PoolDashboard,
  Portfolio,
  Investment,
  OwnershipSummary,
  OwnershipDetail,
} from '../schemas/pool';
import type { TournamentTeam } from '../schemas/tournament';

export function makePool(overrides: Partial<Pool> = {}): Pool {
  return {
    id: 'pool-1',
    name: 'Test Pool',
    tournamentId: 'tourn-1',
    ownerId: 'owner-1',
    minTeams: 3,
    maxTeams: 10,
    maxInvestmentCredits: 50,
    budgetCredits: 100,
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
    ...overrides,
  };
}

export function makePortfolio(overrides: Partial<Portfolio> & { id: string }): Portfolio {
  return {
    name: 'Portfolio',
    poolId: 'pool-1',
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
    ...overrides,
  };
}

export function makeInvestment(
  overrides: Partial<Investment> & { id: string; portfolioId: string; teamId: string },
): Investment {
  return {
    credits: 10,
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
    ...overrides,
  };
}

export function makeOwnershipSummary(
  overrides: Partial<OwnershipSummary> & { id: string; portfolioId: string },
): OwnershipSummary {
  return {
    maximumReturns: 100,
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
    ...overrides,
  };
}

export function makeOwnershipDetail(
  overrides: Partial<OwnershipDetail> & { id: string; portfolioId: string; teamId: string },
): OwnershipDetail {
  return {
    ownershipPercentage: 1,
    actualReturns: 0,
    expectedReturns: 0,
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

export function makeDashboard(overrides: Partial<PoolDashboard> = {}): PoolDashboard {
  return {
    pool: makePool(),
    investingOpen: false,
    totalPortfolios: 0,
    portfolios: [],
    investments: [],
    ownershipSummaries: [],
    ownershipDetails: [],
    schools: [],
    tournamentTeams: [],
    roundStandings: [],
    ...overrides,
  };
}

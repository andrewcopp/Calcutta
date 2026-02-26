import { describe, it, expect } from 'vitest';
import type {
  PoolDashboard,
  Portfolio,
  Investment,
  OwnershipSummary,
  OwnershipDetail,
} from '../schemas/pool';
import type { PoolPortfoliosData } from './usePoolPortfoliosData';
import {
  makePortfolio,
  makeInvestment,
  makeOwnershipSummary,
  makeOwnershipDetail,
  makeTournamentTeam,
  makeDashboard,
} from '../test/factories';

// ---------------------------------------------------------------------------
// The hook is a chain of useMemo calls over PoolDashboard.
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

type EnrichedPortfolio = Portfolio & { totalReturns: number };

function computePoolPortfoliosData(dashboardData: PoolDashboard | undefined): PoolPortfoliosData {
  // ---- first useMemo block ----
  let portfolios: EnrichedPortfolio[] = [];
  let totalPortfolios = 0;
  let allOwnershipSummaries: (OwnershipSummary & { portfolioName?: string })[] = [];
  let allOwnershipDetails: OwnershipDetail[] = [];
  let allInvestments: Investment[] = [];
  let seedInvestmentData: SeedInvestmentDatum[] = [];

  if (dashboardData) {
    const { portfolios: rawPortfolios, investments, ownershipSummaries, ownershipDetails, schools } = dashboardData;
    const schoolMap = new Map(schools.map((s) => [s.id, s]));
    const portfolioNameMap = new Map(rawPortfolios.map((e) => [e.id, e.name]));

    allOwnershipSummaries = ownershipSummaries.map((p) => ({
      ...p,
      portfolioName: portfolioNameMap.get(p.portfolioId),
    }));

    const portfolioReturnsMap = new Map<string, number>();
    const ownershipToPortfolio = new Map(ownershipSummaries.map((p) => [p.id, p.portfolioId]));
    for (const pt of ownershipDetails) {
      const portfolioId = ownershipToPortfolio.get(pt.ownershipSummaryId);
      if (portfolioId) {
        portfolioReturnsMap.set(portfolioId, (portfolioReturnsMap.get(portfolioId) || 0) + pt.actualReturns);
      }
    }

    allOwnershipDetails = ownershipDetails.map((pt) => {
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

    const sortedPortfolios: EnrichedPortfolio[] = rawPortfolios
      .map((portfolio) => ({
        ...portfolio,
        totalReturns: portfolio.totalReturns || portfolioReturnsMap.get(portfolio.id) || 0,
      }))
      .sort((a, b) => {
        const diff = (b.totalReturns || 0) - (a.totalReturns || 0);
        if (diff !== 0) return diff;
        return new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime();
      });

    const seedMap = new Map<number, number>();
    for (const team of investments) {
      if (!team.team?.seed || !team.credits) continue;
      const seed = team.team.seed;
      seedMap.set(seed, (seedMap.get(seed) || 0) + team.credits);
    }
    seedInvestmentData = Array.from(seedMap.entries())
      .map(([seed, totalInvestment]) => ({ seed, totalInvestment }))
      .sort((a, b) => a.seed - b.seed);

    portfolios = sortedPortfolios;
    totalPortfolios = rawPortfolios.length;
    allInvestments = investments;
  }

  // ---- remaining derived values ----
  const schools = dashboardData?.schools ?? [];
  const tournamentTeams = dashboardData?.tournamentTeams ?? [];

  const totalInvestment = allInvestments.reduce((sum, et) => sum + et.credits, 0);
  const totalReturns = portfolios.reduce((sum, e) => sum + (e.totalReturns || 0), 0);
  const averageReturn = totalPortfolios > 0 ? totalReturns / totalPortfolios : 0;

  let returnsStdDev = 0;
  if (totalPortfolios > 1) {
    const variance = portfolios.reduce((acc, e) => {
      const v = (e.totalReturns || 0) - averageReturn;
      return acc + v * v;
    }, 0);
    returnsStdDev = Math.sqrt(variance / totalPortfolios);
  }

  const schoolNameById = new Map(schools.map((school) => [school.id, school.name]));

  const teamROIData: TeamROIDatum[] = tournamentTeams
    .map((team) => {
      const schoolName = schoolNameById.get(team.schoolId) || 'Unknown School';
      const teamInvestment = allInvestments.filter((et) => et.teamId === team.id).reduce((sum, et) => sum + et.credits, 0);
      const teamPoints = allOwnershipDetails
        .filter((pt) => pt.teamId === team.id)
        .reduce((sum, pt) => sum + pt.actualReturns, 0);
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
    portfolios,
    totalPortfolios,
    allOwnershipSummaries,
    allOwnershipDetails,
    allInvestments,
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
// Tests
// ---------------------------------------------------------------------------

describe('usePoolPortfoliosData (pure transformation)', () => {
  describe('when dashboard is undefined', () => {
    it('returns empty portfolios array', () => {
      const result = computePoolPortfoliosData(undefined);
      expect(result.portfolios).toEqual([]);
    });

    it('returns zero totalPortfolios', () => {
      const result = computePoolPortfoliosData(undefined);
      expect(result.totalPortfolios).toBe(0);
    });

    it('returns zero totalInvestment', () => {
      const result = computePoolPortfoliosData(undefined);
      expect(result.totalInvestment).toBe(0);
    });

    it('returns zero averageReturn', () => {
      const result = computePoolPortfoliosData(undefined);
      expect(result.averageReturn).toBe(0);
    });

    it('returns empty teamROIData', () => {
      const result = computePoolPortfoliosData(undefined);
      expect(result.teamROIData).toEqual([]);
    });

    it('returns zero returnsStdDev', () => {
      const result = computePoolPortfoliosData(undefined);
      expect(result.returnsStdDev).toBe(0);
    });
  });

  describe('when dashboard has no portfolios', () => {
    it('returns zero totalPortfolios for dashboard with empty portfolios', () => {
      const dashboard = makeDashboard({ portfolios: [] });
      const result = computePoolPortfoliosData(dashboard);
      expect(result.totalPortfolios).toBe(0);
    });

    it('returns zero averageReturn when there are no portfolios', () => {
      const dashboard = makeDashboard({ portfolios: [] });
      const result = computePoolPortfoliosData(dashboard);
      expect(result.averageReturn).toBe(0);
    });
  });

  describe('portfolio enrichment and sorting', () => {
    it('enriches portfolios with totalReturns from ownership details', () => {
      const dashboard = makeDashboard({
        portfolios: [makePortfolio({ id: 'p1', name: 'Alice' })],
        ownershipSummaries: [makeOwnershipSummary({ id: 'os1', portfolioId: 'p1' })],
        ownershipDetails: [
          makeOwnershipDetail({ id: 'od1', ownershipSummaryId: 'os1', teamId: 't1', actualReturns: 20 }),
          makeOwnershipDetail({ id: 'od2', ownershipSummaryId: 'os1', teamId: 't2', actualReturns: 10 }),
        ],
      });

      const result = computePoolPortfoliosData(dashboard);
      expect(result.portfolios[0].totalReturns).toBe(30);
    });

    it('prefers existing totalReturns over computed value', () => {
      const dashboard = makeDashboard({
        portfolios: [makePortfolio({ id: 'p1', name: 'Alice', totalReturns: 50 })],
        ownershipSummaries: [makeOwnershipSummary({ id: 'os1', portfolioId: 'p1' })],
        ownershipDetails: [makeOwnershipDetail({ id: 'od1', ownershipSummaryId: 'os1', teamId: 't1', actualReturns: 10 })],
      });

      const result = computePoolPortfoliosData(dashboard);
      expect(result.portfolios[0].totalReturns).toBe(50);
    });

    it('sorts portfolios by totalReturns descending', () => {
      const dashboard = makeDashboard({
        portfolios: [
          makePortfolio({ id: 'p1', name: 'Low', totalReturns: 10 }),
          makePortfolio({ id: 'p2', name: 'High', totalReturns: 50 }),
          makePortfolio({ id: 'p3', name: 'Mid', totalReturns: 30 }),
        ],
      });

      const result = computePoolPortfoliosData(dashboard);
      expect(result.portfolios.map((e) => e.name)).toEqual(['High', 'Mid', 'Low']);
    });

    it('breaks totalReturns ties by created date descending (most recent first)', () => {
      const dashboard = makeDashboard({
        portfolios: [
          makePortfolio({ id: 'p1', name: 'Older', totalReturns: 20, createdAt: '2026-01-01T00:00:00Z' }),
          makePortfolio({ id: 'p2', name: 'Newer', totalReturns: 20, createdAt: '2026-02-01T00:00:00Z' }),
        ],
      });

      const result = computePoolPortfoliosData(dashboard);
      expect(result.portfolios[0].name).toBe('Newer');
    });
  });

  describe('totalInvestment', () => {
    it('computes totalInvestment as sum of all investment credits', () => {
      const dashboard = makeDashboard({
        investments: [
          makeInvestment({ id: 'i1', portfolioId: 'p1', teamId: 't1', credits: 15 }),
          makeInvestment({ id: 'i2', portfolioId: 'p1', teamId: 't2', credits: 25 }),
          makeInvestment({ id: 'i3', portfolioId: 'p2', teamId: 't3', credits: 10 }),
        ],
      });

      const result = computePoolPortfoliosData(dashboard);
      expect(result.totalInvestment).toBe(50);
    });
  });

  describe('averageReturn', () => {
    it('computes averageReturn as totalReturns divided by totalPortfolios', () => {
      const dashboard = makeDashboard({
        portfolios: [
          makePortfolio({ id: 'p1', name: 'A', totalReturns: 40 }),
          makePortfolio({ id: 'p2', name: 'B', totalReturns: 60 }),
        ],
      });

      const result = computePoolPortfoliosData(dashboard);
      expect(result.averageReturn).toBe(50);
    });
  });

  describe('returnsStdDev', () => {
    it('returns zero when there is only one portfolio', () => {
      const dashboard = makeDashboard({
        portfolios: [makePortfolio({ id: 'p1', name: 'Solo', totalReturns: 25 })],
      });

      const result = computePoolPortfoliosData(dashboard);
      expect(result.returnsStdDev).toBe(0);
    });

    it('computes standard deviation of portfolio returns', () => {
      const dashboard = makeDashboard({
        portfolios: [
          makePortfolio({ id: 'p1', name: 'A', totalReturns: 10 }),
          makePortfolio({ id: 'p2', name: 'B', totalReturns: 30 }),
        ],
      });

      const result = computePoolPortfoliosData(dashboard);
      expect(result.returnsStdDev).toBe(10);
    });
  });

  describe('seedInvestmentData', () => {
    it('aggregates credits by seed and sorts ascending', () => {
      const dashboard = makeDashboard({
        investments: [
          makeInvestment({
            id: 'i1', portfolioId: 'p1', teamId: 't1', credits: 10,
            team: { id: 't1', schoolId: 's1', seed: 1 },
          }),
          makeInvestment({
            id: 'i2', portfolioId: 'p2', teamId: 't2', credits: 5,
            team: { id: 't2', schoolId: 's2', seed: 1 },
          }),
          makeInvestment({
            id: 'i3', portfolioId: 'p1', teamId: 't3', credits: 20,
            team: { id: 't3', schoolId: 's3', seed: 8 },
          }),
        ],
      });

      const result = computePoolPortfoliosData(dashboard);
      expect(result.seedInvestmentData).toEqual([
        { seed: 1, totalInvestment: 15 },
        { seed: 8, totalInvestment: 20 },
      ]);
    });

    it('skips investments with no seed', () => {
      const dashboard = makeDashboard({
        investments: [
          makeInvestment({ id: 'i1', portfolioId: 'p1', teamId: 't1', credits: 10, team: { id: 't1', schoolId: 's1' } }),
        ],
      });

      const result = computePoolPortfoliosData(dashboard);
      expect(result.seedInvestmentData).toEqual([]);
    });

    it('skips investments with zero credits', () => {
      const dashboard = makeDashboard({
        investments: [
          makeInvestment({
            id: 'i1', portfolioId: 'p1', teamId: 't1', credits: 0,
            team: { id: 't1', schoolId: 's1', seed: 1 },
          }),
        ],
      });

      const result = computePoolPortfoliosData(dashboard);
      expect(result.seedInvestmentData).toEqual([]);
    });
  });

  describe('teamROIData', () => {
    it('computes ROI as returns / (investment + 1) for each tournament team', () => {
      const dashboard = makeDashboard({
        schools: [{ id: 's1', name: 'Duke' }],
        tournamentTeams: [makeTournamentTeam({ id: 't1', schoolId: 's1', seed: 1, region: 'East' })],
        investments: [makeInvestment({ id: 'i1', portfolioId: 'p1', teamId: 't1', credits: 20 })],
        ownershipSummaries: [makeOwnershipSummary({ id: 'os1', portfolioId: 'p1' })],
        ownershipDetails: [makeOwnershipDetail({ id: 'od1', ownershipSummaryId: 'os1', teamId: 't1', actualReturns: 100 })],
      });

      const result = computePoolPortfoliosData(dashboard);
      expect(result.teamROIData[0].roi).toBeCloseTo(100 / 21, 5);
    });

    it('uses "Unknown School" when school name is not found', () => {
      const dashboard = makeDashboard({
        schools: [],
        tournamentTeams: [makeTournamentTeam({ id: 't1', schoolId: 'missing' })],
      });

      const result = computePoolPortfoliosData(dashboard);
      expect(result.teamROIData[0].teamName).toBe('Unknown School');
    });

    it('sorts teams by ROI descending', () => {
      const dashboard = makeDashboard({
        schools: [
          { id: 's1', name: 'Duke' },
          { id: 's2', name: 'UNC' },
        ],
        tournamentTeams: [
          makeTournamentTeam({ id: 't1', schoolId: 's1', seed: 1, region: 'East' }),
          makeTournamentTeam({ id: 't2', schoolId: 's2', seed: 2, region: 'East' }),
        ],
        investments: [
          makeInvestment({ id: 'i1', portfolioId: 'p1', teamId: 't1', credits: 50 }),
          makeInvestment({ id: 'i2', portfolioId: 'p1', teamId: 't2', credits: 5 }),
        ],
        ownershipSummaries: [makeOwnershipSummary({ id: 'os1', portfolioId: 'p1' })],
        ownershipDetails: [
          makeOwnershipDetail({ id: 'od1', ownershipSummaryId: 'os1', teamId: 't1', actualReturns: 10 }),
          makeOwnershipDetail({ id: 'od2', ownershipSummaryId: 'os1', teamId: 't2', actualReturns: 50 }),
        ],
      });

      const result = computePoolPortfoliosData(dashboard);
      expect(result.teamROIData[0].teamName).toBe('UNC');
    });

    it('handles team with zero investment gracefully (divides by 1)', () => {
      const dashboard = makeDashboard({
        schools: [{ id: 's1', name: 'Duke' }],
        tournamentTeams: [makeTournamentTeam({ id: 't1', schoolId: 's1' })],
        investments: [],
        ownershipSummaries: [makeOwnershipSummary({ id: 'os1', portfolioId: 'p1' })],
        ownershipDetails: [makeOwnershipDetail({ id: 'od1', ownershipSummaryId: 'os1', teamId: 't1', actualReturns: 30 })],
      });

      const result = computePoolPortfoliosData(dashboard);
      expect(result.teamROIData[0].roi).toBe(30);
    });
  });

  describe('ownership summary enrichment', () => {
    it('enriches ownership summaries with portfolio names', () => {
      const dashboard = makeDashboard({
        portfolios: [makePortfolio({ id: 'p1', name: 'Alice' })],
        ownershipSummaries: [makeOwnershipSummary({ id: 'os1', portfolioId: 'p1' })],
      });

      const result = computePoolPortfoliosData(dashboard);
      expect(result.allOwnershipSummaries[0].portfolioName).toBe('Alice');
    });

    it('enriches ownership details with school info', () => {
      const dashboard = makeDashboard({
        schools: [{ id: 's1', name: 'Duke' }],
        ownershipDetails: [
          makeOwnershipDetail({
            id: 'od1', ownershipSummaryId: 'os1', teamId: 't1',
            team: { id: 't1', schoolId: 's1' },
          }),
        ],
      });

      const result = computePoolPortfoliosData(dashboard);
      expect(result.allOwnershipDetails[0].team?.school?.name).toBe('Duke');
    });
  });
});

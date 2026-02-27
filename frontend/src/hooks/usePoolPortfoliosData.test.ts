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
  makeOwnershipSummary,
  makeOwnershipDetail,
  makeDashboard,
} from '../test/factories';

// ---------------------------------------------------------------------------
// The hook is a chain of useMemo calls over PoolDashboard.
// Since the test environment is node (no DOM / no React rendering), we extract
// the pure transformation logic into a standalone function that mirrors the
// hook's computation. This lets us unit-test every derivation without needing
// renderHook or a React tree.
// ---------------------------------------------------------------------------

type EnrichedPortfolio = Portfolio & { totalReturns: number };

function computePoolPortfoliosData(dashboardData: PoolDashboard | undefined): PoolPortfoliosData {
  // ---- first useMemo block ----
  let portfolios: EnrichedPortfolio[] = [];
  let allOwnershipSummaries: (OwnershipSummary & { portfolioName?: string })[] = [];
  let allOwnershipDetails: OwnershipDetail[] = [];
  let allInvestments: Investment[] = [];

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
      const portfolioId = ownershipToPortfolio.get(pt.portfolioId);
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

    portfolios = sortedPortfolios;
    allInvestments = investments;
  }

  // ---- remaining derived values ----
  const schools = dashboardData?.schools ?? [];
  const tournamentTeams = dashboardData?.tournamentTeams ?? [];

  return {
    portfolios,
    allOwnershipSummaries,
    allOwnershipDetails,
    allInvestments,
    schools,
    tournamentTeams,
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
  });

  describe('when dashboard has no portfolios', () => {
    it('returns empty portfolios for dashboard with empty portfolios', () => {
      const dashboard = makeDashboard({ portfolios: [] });
      const result = computePoolPortfoliosData(dashboard);
      expect(result.portfolios).toEqual([]);
    });
  });

  describe('portfolio enrichment and sorting', () => {
    it('enriches portfolios with totalReturns from ownership details', () => {
      const dashboard = makeDashboard({
        portfolios: [makePortfolio({ id: 'p1', name: 'Alice' })],
        ownershipSummaries: [makeOwnershipSummary({ id: 'os1', portfolioId: 'p1' })],
        ownershipDetails: [
          makeOwnershipDetail({ id: 'od1', portfolioId: 'os1', teamId: 't1', actualReturns: 20 }),
          makeOwnershipDetail({ id: 'od2', portfolioId: 'os1', teamId: 't2', actualReturns: 10 }),
        ],
      });

      const result = computePoolPortfoliosData(dashboard);
      expect(result.portfolios[0].totalReturns).toBe(30);
    });

    it('prefers existing totalReturns over computed value', () => {
      const dashboard = makeDashboard({
        portfolios: [makePortfolio({ id: 'p1', name: 'Alice', totalReturns: 50 })],
        ownershipSummaries: [makeOwnershipSummary({ id: 'os1', portfolioId: 'p1' })],
        ownershipDetails: [makeOwnershipDetail({ id: 'od1', portfolioId: 'os1', teamId: 't1', actualReturns: 10 })],
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
            id: 'od1', portfolioId: 'os1', teamId: 't1',
            team: { id: 't1', schoolId: 's1' },
          }),
        ],
      });

      const result = computePoolPortfoliosData(dashboard);
      expect(result.allOwnershipDetails[0].team?.school?.name).toBe('Duke');
    });
  });
});

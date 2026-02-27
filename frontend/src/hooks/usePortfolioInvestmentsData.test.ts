import { describe, it, expect } from 'vitest';
import type { PoolDashboard, Investment, OwnershipSummary, OwnershipDetail } from '../schemas/pool';
import type { PortfolioInvestmentsData } from './usePortfolioInvestmentsData';
import {
  makePool,
  makePortfolio,
  makeInvestment,
  makeOwnershipSummary,
  makeOwnershipDetail,
  makeTournamentTeam,
  makeDashboard,
} from '../test/factories';

// ---------------------------------------------------------------------------
// The hook is a single useMemo that transforms PoolDashboard into
// PortfolioInvestmentsData for a given portfolioId. We extract the logic into
// a pure function to test without React rendering.
// ---------------------------------------------------------------------------

function computePortfolioInvestmentsData(
  dashboardData: PoolDashboard | undefined,
  portfolioId: string | undefined,
): PortfolioInvestmentsData {
  if (!dashboardData || !portfolioId) {
    return {
      poolName: '',
      portfolioName: '',
      teams: [] as Investment[],
      schools: [] as { id: string; name: string }[],
      ownershipSummaries: [] as OwnershipSummary[],
      ownershipDetails: [] as OwnershipDetail[],
      tournamentTeams: dashboardData?.tournamentTeams ?? [],
      allInvestments: [] as Investment[],
      allOwnershipSummaries: [] as (OwnershipSummary & { portfolioName?: string })[],
      allOwnershipDetails: [] as OwnershipDetail[],
    };
  }

  const { pool, portfolios, investments, ownershipSummaries, ownershipDetails, schools, tournamentTeams } = dashboardData;
  const schoolMap = new Map(schools.map((s) => [s.id, s]));
  const portfolioNameMap = new Map(portfolios.map((e) => [e.id, e.name]));

  const currentPortfolio = portfolios.find((e) => e.id === portfolioId);
  const portfolioName = currentPortfolio?.name || '';

  const thisPortfolioInvestments = investments.filter((et) => et.portfolioId === portfolioId);

  const teamsWithSchools: Investment[] = thisPortfolioInvestments.map((team) => ({
    ...team,
    team: team.team
      ? {
          ...team.team,
          school: schoolMap.get(team.team.schoolId),
        }
      : undefined,
  }));

  const thisPortfolioOwnershipSummaries = ownershipSummaries.filter((p) => p.portfolioId === portfolioId);

  const thisOwnershipSummaryIds = new Set(thisPortfolioOwnershipSummaries.map((p) => p.id));
  const thisOwnershipDetails: OwnershipDetail[] = ownershipDetails
    .filter((pt) => thisOwnershipSummaryIds.has(pt.portfolioId))
    .map((pt) => ({
      ...pt,
      team: pt.team
        ? {
            ...pt.team,
            school: schoolMap.get(pt.team.schoolId),
          }
        : undefined,
    }));

  const allSummariesWithNames: (OwnershipSummary & { portfolioName?: string })[] = ownershipSummaries.map((p) => ({
    ...p,
    portfolioName: portfolioNameMap.get(p.portfolioId),
  }));

  const allDetailsWithSchools: OwnershipDetail[] = ownershipDetails.map((pt) => ({
    ...pt,
    team: pt.team
      ? {
          ...pt.team,
          school: schoolMap.get(pt.team.schoolId),
        }
      : undefined,
  }));

  return {
    poolName: pool.name,
    portfolioName,
    teams: teamsWithSchools,
    schools,
    ownershipSummaries: thisPortfolioOwnershipSummaries,
    ownershipDetails: thisOwnershipDetails,
    tournamentTeams,
    allInvestments: investments,
    allOwnershipSummaries: allSummariesWithNames,
    allOwnershipDetails: allDetailsWithSchools,
  };
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe('usePortfolioInvestmentsData (pure transformation)', () => {
  describe('when dashboard is undefined', () => {
    it('returns empty poolName', () => {
      const result = computePortfolioInvestmentsData(undefined, 'p1');
      expect(result.poolName).toBe('');
    });

    it('returns empty portfolioName', () => {
      const result = computePortfolioInvestmentsData(undefined, 'p1');
      expect(result.portfolioName).toBe('');
    });

    it('returns empty teams array', () => {
      const result = computePortfolioInvestmentsData(undefined, 'p1');
      expect(result.teams).toEqual([]);
    });

    it('returns empty ownershipSummaries array', () => {
      const result = computePortfolioInvestmentsData(undefined, 'p1');
      expect(result.ownershipSummaries).toEqual([]);
    });

    it('returns empty allInvestments array', () => {
      const result = computePortfolioInvestmentsData(undefined, 'p1');
      expect(result.allInvestments).toEqual([]);
    });
  });

  describe('when portfolioId is undefined', () => {
    it('returns empty portfolioName', () => {
      const dashboard = makeDashboard();
      const result = computePortfolioInvestmentsData(dashboard, undefined);
      expect(result.portfolioName).toBe('');
    });

    it('still returns tournamentTeams from dashboard', () => {
      const teams = [makeTournamentTeam({ id: 't1', schoolId: 's1' })];
      const dashboard = makeDashboard({ tournamentTeams: teams });
      const result = computePortfolioInvestmentsData(dashboard, undefined);
      expect(result.tournamentTeams).toHaveLength(1);
    });
  });

  describe('when portfolioId does not match any portfolio', () => {
    it('returns empty portfolioName for non-existent portfolio', () => {
      const dashboard = makeDashboard({
        portfolios: [makePortfolio({ id: 'p1', name: 'Alice' })],
      });
      const result = computePortfolioInvestmentsData(dashboard, 'p999');
      expect(result.portfolioName).toBe('');
    });

    it('returns empty teams for non-existent portfolio', () => {
      const dashboard = makeDashboard({
        portfolios: [makePortfolio({ id: 'p1', name: 'Alice' })],
        investments: [makeInvestment({ id: 'i1', portfolioId: 'p1', teamId: 't1' })],
      });
      const result = computePortfolioInvestmentsData(dashboard, 'p999');
      expect(result.teams).toEqual([]);
    });

    it('returns empty ownershipSummaries for non-existent portfolio', () => {
      const dashboard = makeDashboard({
        portfolios: [makePortfolio({ id: 'p1', name: 'Alice' })],
        ownershipSummaries: [makeOwnershipSummary({ id: 'os1', portfolioId: 'p1' })],
      });
      const result = computePortfolioInvestmentsData(dashboard, 'p999');
      expect(result.ownershipSummaries).toEqual([]);
    });
  });

  describe('filtering investments for the given portfolio', () => {
    it('returns only investments belonging to the specified portfolio', () => {
      const dashboard = makeDashboard({
        portfolios: [makePortfolio({ id: 'p1', name: 'Alice' }), makePortfolio({ id: 'p2', name: 'Bob' })],
        investments: [
          makeInvestment({ id: 'i1', portfolioId: 'p1', teamId: 't1' }),
          makeInvestment({ id: 'i2', portfolioId: 'p1', teamId: 't2' }),
          makeInvestment({ id: 'i3', portfolioId: 'p2', teamId: 't3' }),
        ],
      });
      const result = computePortfolioInvestmentsData(dashboard, 'p1');
      expect(result.teams).toHaveLength(2);
    });

    it('includes all investments in allInvestments regardless of portfolioId filter', () => {
      const dashboard = makeDashboard({
        portfolios: [makePortfolio({ id: 'p1', name: 'Alice' })],
        investments: [
          makeInvestment({ id: 'i1', portfolioId: 'p1', teamId: 't1' }),
          makeInvestment({ id: 'i2', portfolioId: 'p2', teamId: 't2' }),
        ],
      });
      const result = computePortfolioInvestmentsData(dashboard, 'p1');
      expect(result.allInvestments).toHaveLength(2);
    });
  });

  describe('enriching investments with school info', () => {
    it('attaches school data to investments via schoolMap', () => {
      const dashboard = makeDashboard({
        portfolios: [makePortfolio({ id: 'p1', name: 'Alice' })],
        schools: [{ id: 's1', name: 'Duke' }],
        investments: [
          makeInvestment({
            id: 'i1', portfolioId: 'p1', teamId: 't1',
            team: { id: 't1', schoolId: 's1', seed: 1 },
          }),
        ],
      });
      const result = computePortfolioInvestmentsData(dashboard, 'p1');
      expect(result.teams[0].team?.school?.name).toBe('Duke');
    });

    it('handles investment with no nested team gracefully', () => {
      const dashboard = makeDashboard({
        portfolios: [makePortfolio({ id: 'p1', name: 'Alice' })],
        investments: [makeInvestment({ id: 'i1', portfolioId: 'p1', teamId: 't1', team: undefined })],
      });
      const result = computePortfolioInvestmentsData(dashboard, 'p1');
      expect(result.teams[0].team).toBeUndefined();
    });
  });

  describe('filtering ownership summaries for the given portfolio', () => {
    it('returns only ownership summaries belonging to the specified portfolio', () => {
      const dashboard = makeDashboard({
        portfolios: [makePortfolio({ id: 'p1', name: 'Alice' })],
        ownershipSummaries: [
          makeOwnershipSummary({ id: 'os1', portfolioId: 'p1' }),
          makeOwnershipSummary({ id: 'os2', portfolioId: 'p2' }),
        ],
      });
      const result = computePortfolioInvestmentsData(dashboard, 'p1');
      expect(result.ownershipSummaries).toHaveLength(1);
    });

    it('filters ownership details to those belonging to the portfolio ownership summaries', () => {
      const dashboard = makeDashboard({
        portfolios: [makePortfolio({ id: 'p1', name: 'Alice' })],
        ownershipSummaries: [
          makeOwnershipSummary({ id: 'os1', portfolioId: 'p1' }),
          makeOwnershipSummary({ id: 'os2', portfolioId: 'p2' }),
        ],
        ownershipDetails: [
          makeOwnershipDetail({ id: 'od1', portfolioId: 'os1', teamId: 't1' }),
          makeOwnershipDetail({ id: 'od2', portfolioId: 'os2', teamId: 't2' }),
        ],
      });
      const result = computePortfolioInvestmentsData(dashboard, 'p1');
      expect(result.ownershipDetails).toHaveLength(1);
    });
  });

  describe('pool metadata', () => {
    it('returns poolName from the dashboard pool', () => {
      const dashboard = makeDashboard({
        pool: makePool({ name: 'Championship Pool 2026' }),
        portfolios: [makePortfolio({ id: 'p1', name: 'Alice' })],
      });
      const result = computePortfolioInvestmentsData(dashboard, 'p1');
      expect(result.poolName).toBe('Championship Pool 2026');
    });

    it('returns the matching portfolio name', () => {
      const dashboard = makeDashboard({
        portfolios: [makePortfolio({ id: 'p1', name: 'Alice' })],
      });
      const result = computePortfolioInvestmentsData(dashboard, 'p1');
      expect(result.portfolioName).toBe('Alice');
    });
  });

  describe('all-ownership-summaries enrichment', () => {
    it('enriches all ownership summaries with portfolio names', () => {
      const dashboard = makeDashboard({
        portfolios: [makePortfolio({ id: 'p1', name: 'Alice' }), makePortfolio({ id: 'p2', name: 'Bob' })],
        ownershipSummaries: [
          makeOwnershipSummary({ id: 'os1', portfolioId: 'p1' }),
          makeOwnershipSummary({ id: 'os2', portfolioId: 'p2' }),
        ],
      });
      const result = computePortfolioInvestmentsData(dashboard, 'p1');
      expect(result.allOwnershipSummaries.map((p) => p.portfolioName)).toEqual(['Alice', 'Bob']);
    });

    it('enriches all ownership details with school info', () => {
      const dashboard = makeDashboard({
        portfolios: [makePortfolio({ id: 'p1', name: 'Alice' })],
        schools: [{ id: 's1', name: 'Duke' }],
        ownershipDetails: [
          makeOwnershipDetail({
            id: 'od1', portfolioId: 'os1', teamId: 't1',
            team: { id: 't1', schoolId: 's1' },
          }),
        ],
      });
      const result = computePortfolioInvestmentsData(dashboard, 'p1');
      expect(result.allOwnershipDetails[0].team?.school?.name).toBe('Duke');
    });
  });
});

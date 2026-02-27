import { describe, it, expect } from 'vitest';
import { http, HttpResponse } from 'msw';
import { server } from '../test/msw/server';
import { poolService } from './poolService';

const BASE = 'http://localhost:8080/api/v1';

describe('poolService', () => {
  describe('getAllPools', () => {
    it('returns parsed pools', async () => {
      const pools = await poolService.getAllPools();

      expect(pools).toHaveLength(1);
      expect(pools[0].name).toBe('Test Pool');
    });

    it('throws when pool missing required field', async () => {
      server.use(
        http.get(`${BASE}/pools`, () => {
          return HttpResponse.json({ items: [{ id: 'pool-1' }] });
        }),
      );

      await expect(poolService.getAllPools()).rejects.toThrow();
    });
  });

  describe('getPool', () => {
    it('returns parsed pool', async () => {
      const pool = await poolService.getPool('pool-1');

      expect(pool.id).toBe('pool-1');
      expect(pool.budgetCredits).toBe(100);
    });
  });

  describe('createPool', () => {
    it('returns created pool', async () => {
      const pool = await poolService.createPool(
        'New Pool',
        'tourn-1',
        [{ winIndex: 0, pointsAwarded: 10 }],
        3,
        10,
        50,
      );

      expect(pool.name).toBe('Test Pool');
    });
  });

  describe('updatePool', () => {
    it('returns updated pool', async () => {
      const pool = await poolService.updatePool('pool-1', { name: 'Updated' });

      expect(pool.id).toBe('pool-1');
    });
  });

  describe('createPortfolio', () => {
    it('returns created portfolio', async () => {
      const portfolio = await poolService.createPortfolio('pool-1', 'My Portfolio', [
        { teamId: 'team-1', credits: 10 },
      ]);

      expect(portfolio.id).toBe('portfolio-1');
    });
  });

  describe('deletePortfolio', () => {
    it('completes without error', async () => {
      await expect(poolService.deletePortfolio('pool-1', 'portfolio-1')).resolves.toBeUndefined();
    });
  });

  describe('getInvestments', () => {
    it('returns parsed investments', async () => {
      const investments = await poolService.getInvestments('portfolio-1', 'pool-1');

      expect(investments).toHaveLength(1);
      expect(investments[0].credits).toBe(10);
    });
  });

  describe('getPoolDashboard', () => {
    it('returns parsed dashboard with nested objects', async () => {
      const dashboard = await poolService.getPoolDashboard('pool-1');

      expect(dashboard.pool.id).toBe('pool-1');
      expect(dashboard.portfolios).toHaveLength(1);
      expect(dashboard.schools).toHaveLength(1);
      expect(dashboard.tournamentTeams).toHaveLength(1);
    });

    it('throws when dashboard missing pool', async () => {
      server.use(
        http.get(`${BASE}/pools/:id/dashboard`, () => {
          return HttpResponse.json({ investingOpen: true, totalPortfolios: 0, scoringRules: [], portfolios: [], investments: [], ownershipSummaries: [], ownershipDetails: [], schools: [], tournamentTeams: [], roundStandings: [] });
        }),
      );

      await expect(poolService.getPoolDashboard('pool-1')).rejects.toThrow();
    });
  });

  describe('getPoolsWithRankings', () => {
    it('returns pools with ranking data', async () => {
      const pools = await poolService.getPoolsWithRankings();

      expect(pools).toHaveLength(1);
      expect(pools[0].hasPortfolio).toBe(true);
      expect(pools[0].ranking?.rank).toBe(1);
    });
  });

  describe('updatePortfolio', () => {
    it('returns updated portfolio', async () => {
      const portfolio = await poolService.updatePortfolio('pool-1', 'portfolio-1', [{ teamId: 'team-1', credits: 20 }]);

      expect(portfolio.id).toBe('portfolio-1');
    });
  });

  describe('getPayouts', () => {
    it('returns parsed payouts', async () => {
      const result = await poolService.getPayouts('pool-1');

      expect(result).toHaveLength(1);
      expect(result[0].amountCents).toBe(10000);
    });
  });

  describe('replacePayouts', () => {
    it('returns replaced payouts', async () => {
      const result = await poolService.replacePayouts('pool-1', [{ position: 1, amountCents: 10000 }]);

      expect(result).toHaveLength(1);
    });
  });
});

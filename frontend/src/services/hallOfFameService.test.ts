import { describe, it, expect } from 'vitest';
import { http, HttpResponse } from 'msw';
import { server } from '../test/msw/server';
import { hallOfFameService } from './hallOfFameService';

const BASE = 'http://localhost:8080/api/v1';

describe('hallOfFameService', () => {
  describe('getBestTeams', () => {
    it('returns parsed best teams response', async () => {
      const result = await hallOfFameService.getBestTeams();

      expect(result.items).toHaveLength(1);
      expect(result.items[0].schoolName).toBe('Duke');
      expect(result.items[0].rawROI).toBe(2.0);
    });

    it('throws when team missing required field', async () => {
      server.use(
        http.get(`${BASE}/hall-of-fame/best-teams`, () => {
          return HttpResponse.json({ items: [{ schoolName: 'Duke' }] });
        }),
      );

      await expect(hallOfFameService.getBestTeams()).rejects.toThrow();
    });
  });

  describe('getBestInvestments', () => {
    it('returns parsed investments response', async () => {
      const result = await hallOfFameService.getBestInvestments();

      expect(result.items).toHaveLength(1);
      expect(result.items[0].entryName).toBe('Test Entry');
    });

    it('throws when response missing items key', async () => {
      server.use(
        http.get(`${BASE}/hall-of-fame/best-investments`, () => {
          return HttpResponse.json({ investments: [] });
        }),
      );

      await expect(hallOfFameService.getBestInvestments()).rejects.toThrow();
    });
  });

  describe('getBestEntries', () => {
    it('returns parsed entries response', async () => {
      const result = await hallOfFameService.getBestEntries();

      expect(result.items).toHaveLength(1);
      expect(result.items[0].normalizedReturns).toBe(1.5);
    });
  });

  describe('getBestCareers', () => {
    it('returns parsed careers response', async () => {
      const result = await hallOfFameService.getBestCareers();

      expect(result.items).toHaveLength(1);
      expect(result.items[0].wins).toBe(3);
      expect(result.items[0].careerEarningsCents).toBe(50000);
    });

    it('throws when career missing required boolean', async () => {
      server.use(
        http.get(`${BASE}/hall-of-fame/best-careers`, () => {
          return HttpResponse.json({
            items: [{ entryName: 'Test', wins: 1, years: 1, bestFinish: 1, podiums: 0, inTheMoneys: 0, top10s: 0, careerEarningsCents: 0 }],
          });
        }),
      );

      await expect(hallOfFameService.getBestCareers()).rejects.toThrow();
    });
  });
});

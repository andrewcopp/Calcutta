import { describe, it, expect } from 'vitest';
import { http, HttpResponse } from 'msw';
import { server } from '../test/msw/server';
import { calcuttaService } from './calcuttaService';

const BASE = 'http://localhost:8080/api';

describe('calcuttaService', () => {
  describe('getAllCalcuttas', () => {
    it('returns parsed calcuttas', async () => {
      const calcuttas = await calcuttaService.getAllCalcuttas();

      expect(calcuttas).toHaveLength(1);
      expect(calcuttas[0].name).toBe('Test Pool');
    });

    it('throws when calcutta missing required field', async () => {
      server.use(
        http.get(`${BASE}/calcuttas`, () => {
          return HttpResponse.json([{ id: 'calc-1' }]);
        }),
      );

      await expect(calcuttaService.getAllCalcuttas()).rejects.toThrow();
    });
  });

  describe('getCalcutta', () => {
    it('returns parsed calcutta', async () => {
      const calcutta = await calcuttaService.getCalcutta('calc-1');

      expect(calcutta.id).toBe('calc-1');
      expect(calcutta.budgetPoints).toBe(100);
    });
  });

  describe('createCalcutta', () => {
    it('returns created calcutta', async () => {
      const calcutta = await calcuttaService.createCalcutta(
        'New Pool',
        'tourn-1',
        [{ winIndex: 0, pointsAwarded: 10 }],
        3,
        10,
        50,
      );

      expect(calcutta.name).toBe('Test Pool');
    });
  });

  describe('updateCalcutta', () => {
    it('returns updated calcutta', async () => {
      const calcutta = await calcuttaService.updateCalcutta('calc-1', { name: 'Updated' });

      expect(calcutta.id).toBe('calc-1');
    });
  });

  describe('createEntry', () => {
    it('returns created entry', async () => {
      const entry = await calcuttaService.createEntry('calc-1', 'My Entry');

      expect(entry.id).toBe('entry-1');
      expect(entry.status).toBe('accepted');
    });
  });

  describe('getEntryTeams', () => {
    it('returns parsed entry teams', async () => {
      const teams = await calcuttaService.getEntryTeams('entry-1', 'calc-1');

      expect(teams).toHaveLength(1);
      expect(teams[0].bid).toBe(10);
    });
  });

  describe('getCalcuttaDashboard', () => {
    it('returns parsed dashboard with nested objects', async () => {
      const dashboard = await calcuttaService.getCalcuttaDashboard('calc-1');

      expect(dashboard.calcutta.id).toBe('calc-1');
      expect(dashboard.entries).toHaveLength(1);
      expect(dashboard.schools).toHaveLength(1);
      expect(dashboard.tournamentTeams).toHaveLength(1);
    });

    it('throws when dashboard missing calcutta', async () => {
      server.use(
        http.get(`${BASE}/calcuttas/:id/dashboard`, () => {
          return HttpResponse.json({ biddingOpen: true, totalEntries: 0, entries: [], entryTeams: [], portfolios: [], portfolioTeams: [], schools: [], tournamentTeams: [] });
        }),
      );

      await expect(calcuttaService.getCalcuttaDashboard('calc-1')).rejects.toThrow();
    });
  });

  describe('getCalcuttasWithRankings', () => {
    it('returns calcuttas with ranking data', async () => {
      const calcuttas = await calcuttaService.getCalcuttasWithRankings();

      expect(calcuttas).toHaveLength(1);
      expect(calcuttas[0].hasEntry).toBe(true);
      expect(calcuttas[0].ranking?.rank).toBe(1);
    });
  });

  describe('updateEntry', () => {
    it('returns updated entry', async () => {
      const entry = await calcuttaService.updateEntry('entry-1', [{ teamId: 'team-1', bid: 20 }]);

      expect(entry.id).toBe('entry-1');
    });
  });

  describe('getPayouts', () => {
    it('returns parsed payouts', async () => {
      const result = await calcuttaService.getPayouts('calc-1');

      expect(result.payouts).toHaveLength(1);
      expect(result.payouts[0].amountCents).toBe(10000);
    });
  });

  describe('replacePayouts', () => {
    it('returns replaced payouts', async () => {
      const result = await calcuttaService.replacePayouts('calc-1', [{ position: 1, amountCents: 10000 }]);

      expect(result.payouts).toHaveLength(1);
    });
  });
});

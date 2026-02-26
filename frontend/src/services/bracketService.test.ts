import { describe, it, expect } from 'vitest';
import { http, HttpResponse } from 'msw';
import { server } from '../test/msw/server';
import { bracketService } from './bracketService';

const BASE = 'http://localhost:8080/api/v1';

describe('bracketService', () => {
  describe('fetchBracket', () => {
    it('returns parsed bracket structure', async () => {
      const bracket = await bracketService.fetchBracket('tourn-1');

      expect(bracket.tournamentId).toBe('tourn-1');
      expect(bracket.regions).toEqual(['East', 'West', 'South', 'Midwest']);
      expect(bracket.games).toHaveLength(1);
    });

    it('throws when response missing regions', async () => {
      server.use(
        http.get(`${BASE}/tournaments/:id/bracket`, () => {
          return HttpResponse.json({ tournamentId: 'tourn-1', games: [] });
        }),
      );

      await expect(bracketService.fetchBracket('tourn-1')).rejects.toThrow();
    });
  });

  describe('selectWinner', () => {
    it('returns updated bracket with winner set', async () => {
      const bracket = await bracketService.selectWinner('tourn-1', 'game-1', 't-1');

      expect(bracket.games[0].winner?.teamId).toBe('t-1');
    });
  });

  describe('unselectWinner', () => {
    it('returns bracket without winner', async () => {
      const bracket = await bracketService.unselectWinner('tourn-1', 'game-1');

      expect(bracket.games[0].winner).toBeUndefined();
    });
  });

  describe('validateBracketSetup', () => {
    it('returns validation result', async () => {
      const result = await bracketService.validateBracketSetup('tourn-1');

      expect(result.valid).toBe(true);
      expect(result.errors).toEqual([]);
    });

    it('throws when response missing valid field', async () => {
      server.use(
        http.get(`${BASE}/tournaments/:id/bracket/validate`, () => {
          return HttpResponse.json({ errors: [] });
        }),
      );

      await expect(bracketService.validateBracketSetup('tourn-1')).rejects.toThrow();
    });
  });
});

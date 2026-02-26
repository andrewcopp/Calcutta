import { describe, it, expect } from 'vitest';
import { http, HttpResponse } from 'msw';
import { server } from '../test/msw/server';
import { tournamentService } from './tournamentService';

const BASE = 'http://localhost:8080/api/v1';

describe('tournamentService', () => {
  describe('getAllTournaments', () => {
    it('returns parsed tournaments', async () => {
      const tournaments = await tournamentService.getAllTournaments();

      expect(tournaments).toHaveLength(1);
      expect(tournaments[0].name).toBe('NCAA 2025');
    });

    it('throws when tournament missing required field', async () => {
      server.use(
        http.get(`${BASE}/tournaments`, () => {
          return HttpResponse.json({ items: [{ id: 'tourn-1' }] });
        }),
      );

      await expect(tournamentService.getAllTournaments()).rejects.toThrow();
    });
  });

  describe('getTournament', () => {
    it('returns single parsed tournament', async () => {
      const tournament = await tournamentService.getTournament('tourn-1');

      expect(tournament.id).toBe('tourn-1');
      expect(tournament.rounds).toBe(6);
    });
  });

  describe('getTournamentTeams', () => {
    it('returns parsed tournament teams with school', async () => {
      const teams = await tournamentService.getTournamentTeams('tourn-1');

      expect(teams).toHaveLength(1);
      expect(teams[0].school?.name).toBe('Duke');
    });
  });

  describe('createTournament', () => {
    it('returns created tournament', async () => {
      const tournament = await tournamentService.createTournament('NCAA', 2025, 6);

      expect(tournament.name).toBe('NCAA 2025');
    });
  });

  describe('updateTournament', () => {
    it('returns updated tournament', async () => {
      const tournament = await tournamentService.updateTournament('tourn-1', {
        startingAt: '2025-03-21T00:00:00Z',
      });

      expect(tournament.id).toBe('tourn-1');
    });
  });

  describe('replaceTeams', () => {
    it('returns replaced teams', async () => {
      const teams = await tournamentService.replaceTeams('tourn-1', [
        { schoolId: 'sch-1', seed: 1, region: 'East' },
      ]);

      expect(teams).toHaveLength(1);
      expect(teams[0].seed).toBe(1);
    });
  });

  describe('getCompetitions', () => {
    it('returns parsed competitions', async () => {
      const competitions = await tournamentService.getCompetitions();

      expect(competitions).toHaveLength(1);
      expect(competitions[0].name).toBe('NCAA');
    });
  });

  describe('getSeasons', () => {
    it('returns parsed seasons', async () => {
      const seasons = await tournamentService.getSeasons();

      expect(seasons).toHaveLength(1);
      expect(seasons[0].year).toBe(2025);
    });
  });

  describe('getTournamentModerators', () => {
    it('returns moderators array', async () => {
      const mods = await tournamentService.getTournamentModerators('tourn-1');

      expect(mods).toHaveLength(1);
      expect(mods[0].email).toBe('mod@test.com');
    });
  });

  describe('grantTournamentModerator', () => {
    it('returns new moderator', async () => {
      const mod = await tournamentService.grantTournamentModerator('tourn-1', 'new@test.com');

      expect(mod.email).toBe('new@test.com');
    });
  });

  describe('revokeTournamentModerator', () => {
    it('completes without error', async () => {
      await expect(tournamentService.revokeTournamentModerator('tourn-1', 'mod-1')).resolves.toBeUndefined();
    });
  });

  describe('updateKenPomStats', () => {
    it('returns teams with kenPom stats', async () => {
      const teams = await tournamentService.updateKenPomStats('tourn-1', [
        { teamId: 'team-1', netRtg: 25.5, oRtg: 118.0, dRtg: 92.5, adjT: 68.0 },
      ]);

      expect(teams[0].kenPom?.netRtg).toBe(25.5);
    });
  });
});

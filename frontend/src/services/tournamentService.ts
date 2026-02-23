import { z } from 'zod';
import {
  TournamentSchema,
  TournamentTeamSchema,
  TournamentModeratorsResponseSchema,
  GrantModeratorResponseSchema,
  CompetitionSchema,
  SeasonSchema,
  TournamentPredictionsSchema,
} from '../schemas/tournament';
import type { Tournament } from '../schemas/tournament';
import { apiClient } from '../api/apiClient';

export const tournamentService = {
  async getAllTournaments() {
    return apiClient.get('/tournaments', { schema: z.array(TournamentSchema) });
  },

  async getTournament(id: string) {
    return apiClient.get(`/tournaments/${id}`, { schema: TournamentSchema });
  },

  async getTournamentTeams(id: string) {
    return apiClient.get(`/tournaments/${id}/teams`, { schema: z.array(TournamentTeamSchema) });
  },

  async createTournament(competition: string, year: number, rounds: number) {
    return apiClient.post('/tournaments', { competition, year, rounds }, { schema: TournamentSchema });
  },

  async updateTournament(
    id: string,
    updates: {
      startingAt?: string | null;
      finalFourTopLeft?: string;
      finalFourBottomLeft?: string;
      finalFourTopRight?: string;
      finalFourBottomRight?: string;
    },
  ): Promise<Tournament> {
    return apiClient.patch(`/tournaments/${id}`, updates, { schema: TournamentSchema });
  },

  async replaceTeams(tournamentId: string, teams: { schoolId: string; seed: number; region: string }[]) {
    return apiClient.put(`/tournaments/${tournamentId}/teams`, { teams }, { schema: z.array(TournamentTeamSchema) });
  },

  async getCompetitions() {
    return apiClient.get('/competitions', { schema: z.array(CompetitionSchema) });
  },

  async getSeasons() {
    return apiClient.get('/seasons', { schema: z.array(SeasonSchema) });
  },

  async getTournamentModerators(tournamentId: string) {
    const res = await apiClient.get(`/tournaments/${tournamentId}/moderators`, {
      schema: TournamentModeratorsResponseSchema,
    });
    return res.moderators;
  },

  async grantTournamentModerator(tournamentId: string, email: string) {
    const res = await apiClient.post(`/tournaments/${tournamentId}/moderators`, { email }, {
      schema: GrantModeratorResponseSchema,
    });
    return res.moderator;
  },

  async revokeTournamentModerator(tournamentId: string, userId: string): Promise<void> {
    await apiClient.delete(`/tournaments/${tournamentId}/moderators/${userId}`);
  },

  async getTournamentPredictions(tournamentId: string) {
    return apiClient.get(`/tournaments/${tournamentId}/predictions`, {
      schema: TournamentPredictionsSchema,
    });
  },

  async updateKenPomStats(
    tournamentId: string,
    stats: { teamId: string; netRtg: number; oRtg: number; dRtg: number; adjT: number }[],
  ) {
    return apiClient.put(`/tournaments/${tournamentId}/kenpom`, { stats }, { schema: z.array(TournamentTeamSchema) });
  },
};

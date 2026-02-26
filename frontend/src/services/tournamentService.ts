import { z } from 'zod';
import {
  TournamentSchema,
  TournamentTeamSchema,
  TournamentModeratorsResponseSchema,
  GrantModeratorResponseSchema,
  CompetitionSchema,
  SeasonSchema,
  TournamentPredictionsSchema,
  PredictionBatchSchema,
} from '../schemas/tournament';
import type { Tournament } from '../schemas/tournament';
import { apiClient } from '../api/apiClient';

export const tournamentService = {
  async getAllTournaments() {
    const res = await apiClient.get('/tournaments', { schema: z.object({ items: z.array(TournamentSchema) }) });
    return res.items;
  },

  async getTournament(id: string) {
    return apiClient.get(`/tournaments/${id}`, { schema: TournamentSchema });
  },

  async getTournamentTeams(id: string) {
    const res = await apiClient.get(`/tournaments/${id}/teams`, { schema: z.object({ items: z.array(TournamentTeamSchema) }) });
    return res.items;
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
    const res = await apiClient.put(`/tournaments/${tournamentId}/teams`, { teams }, { schema: z.object({ items: z.array(TournamentTeamSchema) }) });
    return res.items;
  },

  async getCompetitions() {
    const res = await apiClient.get('/competitions', { schema: z.object({ items: z.array(CompetitionSchema) }) });
    return res.items;
  },

  async getSeasons() {
    const res = await apiClient.get('/seasons', { schema: z.object({ items: z.array(SeasonSchema) }) });
    return res.items;
  },

  async getTournamentModerators(tournamentId: string) {
    const res = await apiClient.get(`/tournaments/${tournamentId}/moderators`, {
      schema: TournamentModeratorsResponseSchema,
    });
    return res.items;
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

  async getPredictionBatches(tournamentId: string) {
    const res = await apiClient.get(`/tournaments/${tournamentId}/prediction-batches`, {
      schema: z.object({ items: z.array(PredictionBatchSchema) }),
    });
    return res.items;
  },

  async getTournamentPredictions(tournamentId: string, batchId?: string) {
    const params = batchId ? `?batch_id=${batchId}` : '';
    return apiClient.get(`/tournaments/${tournamentId}/predictions${params}`, {
      schema: TournamentPredictionsSchema,
    });
  },

  async updateKenPomStats(
    tournamentId: string,
    stats: { teamId: string; netRtg: number; oRtg: number; dRtg: number; adjT: number }[],
  ) {
    const res = await apiClient.put(`/tournaments/${tournamentId}/kenpom`, { stats }, { schema: z.object({ items: z.array(TournamentTeamSchema) }) });
    return res.items;
  },
};

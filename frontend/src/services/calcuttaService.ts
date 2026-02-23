import { z } from 'zod';
import {
  CalcuttaSchema,
  CalcuttaDashboardSchema,
  CalcuttaEntrySchema,
  CalcuttaEntryTeamSchema,
  CalcuttaWithRankingSchema,
  PayoutsResponseSchema,
} from '../schemas/calcutta';
import type { Calcutta } from '../schemas/calcutta';
import { apiClient } from '../api/apiClient';

export const calcuttaService = {
  async getAllCalcuttas() {
    return apiClient.get('/calcuttas', { schema: z.array(CalcuttaSchema) });
  },

  async getCalcutta(id: string) {
    return apiClient.get(`/calcuttas/${id}`, { schema: CalcuttaSchema });
  },

  async createCalcutta(
    name: string,
    tournamentId: string,
    scoringRules: Array<{ winIndex: number; pointsAwarded: number }>,
    minTeams: number,
    maxTeams: number,
    maxBidPoints: number,
  ) {
    return apiClient.post('/calcuttas', {
      name,
      tournamentId,
      scoringRules,
      minTeams,
      maxTeams,
      maxBidPoints,
    }, { schema: CalcuttaSchema });
  },

  async updateCalcutta(
    id: string,
    updates: Partial<Pick<Calcutta, 'name' | 'minTeams' | 'maxTeams' | 'maxBidPoints'>>,
  ) {
    return apiClient.patch(`/calcuttas/${id}`, updates, { schema: CalcuttaSchema });
  },

  async createEntry(calcuttaId: string, name: string) {
    return apiClient.post(`/calcuttas/${calcuttaId}/entries`, { name }, { schema: CalcuttaEntrySchema });
  },

  async getEntryTeams(entryId: string, calcuttaId: string) {
    return apiClient.get(`/calcuttas/${calcuttaId}/entries/${entryId}/teams`, {
      schema: z.array(CalcuttaEntryTeamSchema),
    });
  },

  async getCalcuttaDashboard(calcuttaId: string) {
    return apiClient.get(`/calcuttas/${calcuttaId}/dashboard`, { schema: CalcuttaDashboardSchema });
  },

  async getCalcuttasWithRankings() {
    return apiClient.get('/calcuttas/list-with-rankings', { schema: z.array(CalcuttaWithRankingSchema) });
  },

  async updateEntry(entryId: string, teams: Array<{ teamId: string; bid: number }>) {
    return apiClient.patch(`/entries/${entryId}`, { teams }, { schema: CalcuttaEntrySchema });
  },

  async getPayouts(calcuttaId: string) {
    return apiClient.get(`/calcuttas/${calcuttaId}/payouts`, { schema: PayoutsResponseSchema });
  },

  async replacePayouts(calcuttaId: string, payouts: Array<{ position: number; amountCents: number }>) {
    return apiClient.put(`/calcuttas/${calcuttaId}/payouts`, { payouts }, { schema: PayoutsResponseSchema });
  },
};

import { z } from 'zod';
import {
  PoolSchema,
  PoolDashboardSchema,
  PortfolioSchema,
  InvestmentSchema,
  PoolWithRankingSchema,
  PayoutsResponseSchema,
} from '../schemas/pool';
import type { Pool } from '../schemas/pool';
import { apiClient } from '../api/apiClient';

export const poolService = {
  async getAllPools() {
    const res = await apiClient.get('/pools', { schema: z.object({ items: z.array(PoolSchema) }) });
    return res.items;
  },

  async getPool(id: string) {
    return apiClient.get(`/pools/${id}`, { schema: PoolSchema });
  },

  async createPool(
    name: string,
    tournamentId: string,
    scoringRules: Array<{ winIndex: number; pointsAwarded: number }>,
    minTeams: number,
    maxTeams: number,
    maxInvestmentCredits: number,
  ) {
    return apiClient.post('/pools', {
      name,
      tournamentId,
      scoringRules,
      minTeams,
      maxTeams,
      maxInvestmentCredits,
    }, { schema: PoolSchema });
  },

  async updatePool(
    id: string,
    updates: Partial<Pick<Pool, 'name' | 'minTeams' | 'maxTeams' | 'maxInvestmentCredits'>>,
  ) {
    return apiClient.patch(`/pools/${id}`, updates, { schema: PoolSchema });
  },

  async createPortfolio(poolId: string, name: string) {
    return apiClient.post(`/pools/${poolId}/portfolios`, { name }, { schema: PortfolioSchema });
  },

  async getInvestments(portfolioId: string, poolId: string) {
    const res = await apiClient.get(`/pools/${poolId}/portfolios/${portfolioId}/investments`, {
      schema: z.object({ items: z.array(InvestmentSchema) }),
    });
    return res.items;
  },

  async getPoolDashboard(poolId: string) {
    return apiClient.get(`/pools/${poolId}/dashboard`, { schema: PoolDashboardSchema });
  },

  async getPoolsWithRankings() {
    const res = await apiClient.get('/pools?include=rankings', {
      schema: z.object({ items: z.array(PoolWithRankingSchema) }),
    });
    return res.items;
  },

  async updatePortfolio(poolId: string, portfolioId: string, teams: Array<{ teamId: string; credits: number }>) {
    return apiClient.patch(`/pools/${poolId}/portfolios/${portfolioId}`, { teams }, { schema: PortfolioSchema });
  },

  async getPayouts(poolId: string) {
    const res = await apiClient.get(`/pools/${poolId}/payouts`, { schema: PayoutsResponseSchema });
    return res.items;
  },

  async replacePayouts(poolId: string, payouts: Array<{ position: number; amountCents: number }>) {
    const res = await apiClient.put(`/pools/${poolId}/payouts`, { payouts }, { schema: PayoutsResponseSchema });
    return res.items;
  },
};

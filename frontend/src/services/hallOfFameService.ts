import { apiClient } from '../api/apiClient';
import {
  BestTeamsResponseSchema,
  InvestmentLeaderboardResponseSchema,
  EntryLeaderboardResponseSchema,
  CareerLeaderboardResponseSchema,
} from '../schemas/hallOfFame';

export const hallOfFameService = {
  async getBestTeams(limit = 200) {
    return apiClient.get(`/hall-of-fame/best-teams?limit=${limit}`, { schema: BestTeamsResponseSchema });
  },

  async getBestInvestments(limit = 200) {
    return apiClient.get(`/hall-of-fame/best-investments?limit=${limit}`, { schema: InvestmentLeaderboardResponseSchema });
  },

  async getBestEntries(limit = 200) {
    return apiClient.get(`/hall-of-fame/best-entries?limit=${limit}`, { schema: EntryLeaderboardResponseSchema });
  },

  async getBestCareers(limit = 200) {
    return apiClient.get(`/hall-of-fame/best-careers?limit=${limit}`, { schema: CareerLeaderboardResponseSchema });
  },
};

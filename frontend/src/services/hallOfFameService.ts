import { apiClient } from '../api/apiClient';
import {
  BestTeamsResponse,
  InvestmentLeaderboardResponse,
  EntryLeaderboardResponse,
  CareerLeaderboardResponse,
} from '../types/hallOfFame';

export const hallOfFameService = {
  async getBestTeams(limit = 200): Promise<BestTeamsResponse> {
    return apiClient.get<BestTeamsResponse>(`/hall-of-fame/best-teams?limit=${limit}`);
  },

  async getBestInvestments(limit = 200): Promise<InvestmentLeaderboardResponse> {
    return apiClient.get<InvestmentLeaderboardResponse>(`/hall-of-fame/best-investments?limit=${limit}`);
  },

  async getBestEntries(limit = 200): Promise<EntryLeaderboardResponse> {
    return apiClient.get<EntryLeaderboardResponse>(`/hall-of-fame/best-entries?limit=${limit}`);
  },

  async getBestCareers(limit = 200): Promise<CareerLeaderboardResponse> {
    return apiClient.get<CareerLeaderboardResponse>(`/hall-of-fame/best-careers?limit=${limit}`);
  },
};

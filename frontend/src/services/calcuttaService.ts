import { Calcutta, CalcuttaEntry, CalcuttaEntryTeam, CalcuttaPortfolio, CalcuttaPortfolioTeam, School, TournamentTeam } from '../types/calcutta';
import { apiClient } from '../api/apiClient';

export const calcuttaService = {
  async getAllCalcuttas(): Promise<Calcutta[]> {
    return apiClient.get<Calcutta[]>('/calcuttas');
  },

  async getCalcutta(id: string): Promise<Calcutta> {
    return apiClient.get<Calcutta>(`/calcuttas/${id}`);
  },

  async createCalcutta(name: string, tournamentId: string, ownerId: string): Promise<Calcutta> {
    return apiClient.post<Calcutta>('/calcuttas', {
      name,
      tournamentId,
      ownerId,
    });
  },

  async getCalcuttaEntries(calcuttaId: string): Promise<CalcuttaEntry[]> {
    return apiClient.get<CalcuttaEntry[]>(`/calcuttas/${calcuttaId}/entries`);
  },

  async getEntryTeams(entryId: string, calcuttaId: string): Promise<CalcuttaEntryTeam[]> {
    return apiClient.get<CalcuttaEntryTeam[]>(`/calcuttas/${calcuttaId}/entries/${entryId}/teams`);
  },

  async getSchools(): Promise<School[]> {
    return apiClient.get<School[]>('/schools');
  },

  async getTournamentTeams(tournamentId: string): Promise<TournamentTeam[]> {
    return apiClient.get<TournamentTeam[]>(`/tournaments/${tournamentId}/teams`);
  },

  async getPortfoliosByEntry(entryId: string): Promise<CalcuttaPortfolio[]> {
    return apiClient.get<CalcuttaPortfolio[]>(`/entries/${entryId}/portfolios`);
  },

  async getPortfolioTeams(portfolioId: string): Promise<CalcuttaPortfolioTeam[]> {
    return apiClient.get<CalcuttaPortfolioTeam[]>(`/portfolios/${portfolioId}/teams`);
  },

  async calculatePortfolioScores(portfolioId: string): Promise<void> {
    await apiClient.post<void>(`/portfolios/${portfolioId}/calculate-scores`);
  },

  async updatePortfolioTeamScores(
    portfolioId: string,
    teamId: string,
    expectedPoints: number,
    predictedPoints: number
  ): Promise<void> {
    await apiClient.put<void>(`/portfolios/${portfolioId}/teams/${teamId}/scores`, {
      expectedPoints,
      predictedPoints,
    });
  },

  async updatePortfolioMaximumScore(portfolioId: string, maximumPoints: number): Promise<void> {
    await apiClient.put<void>(`/portfolios/${portfolioId}/maximum-score`, {
      maximumPoints,
    });
  },
}; 
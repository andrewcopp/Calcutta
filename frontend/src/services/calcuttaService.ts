import { Calcutta, CalcuttaDashboard, CalcuttaEntry, CalcuttaEntryTeam, CalcuttaPortfolio, CalcuttaPortfolioTeam, CalcuttaWithRanking } from '../types/calcutta';
import { apiClient } from '../api/apiClient';

export const calcuttaService = {
  async getAllCalcuttas(): Promise<Calcutta[]> {
    return apiClient.get<Calcutta[]>('/calcuttas');
  },

  async getCalcutta(id: string): Promise<Calcutta> {
    return apiClient.get<Calcutta>(`/calcuttas/${id}`);
  },

  async createCalcutta(name: string, tournamentId: string): Promise<Calcutta> {
    return apiClient.post<Calcutta>('/calcuttas', {
      name,
      tournamentId,
    });
  },

  async updateCalcutta(id: string, updates: Partial<Pick<Calcutta, 'name' | 'minTeams' | 'maxTeams' | 'maxBid' | 'biddingOpen'>>): Promise<Calcutta> {
    return apiClient.patch<Calcutta>(`/calcuttas/${id}`, updates);
  },

  async getCalcuttaEntries(calcuttaId: string): Promise<CalcuttaEntry[]> {
    return apiClient.get<CalcuttaEntry[]>(`/calcuttas/${calcuttaId}/entries`);
  },

  async getEntryTeams(entryId: string, calcuttaId: string): Promise<CalcuttaEntryTeam[]> {
    return apiClient.get<CalcuttaEntryTeam[]>(`/calcuttas/${calcuttaId}/entries/${entryId}/teams`);
  },

  async getPortfoliosByEntry(entryId: string): Promise<CalcuttaPortfolio[]> {
    return apiClient.get<CalcuttaPortfolio[]>(`/entries/${entryId}/portfolios`);
  },

  async getPortfolioTeams(portfolioId: string): Promise<CalcuttaPortfolioTeam[]> {
    return apiClient.get<CalcuttaPortfolioTeam[]>(`/portfolios/${portfolioId}/teams`);
  },

  async getCalcuttaDashboard(calcuttaId: string): Promise<CalcuttaDashboard> {
    return apiClient.get<CalcuttaDashboard>(`/calcuttas/${calcuttaId}/dashboard`);
  },

  async getCalcuttasWithRankings(): Promise<CalcuttaWithRanking[]> {
    return apiClient.get<CalcuttaWithRanking[]>('/calcuttas/list-with-rankings');
  },

  async updateEntry(
    entryId: string,
    teams: Array<{ teamId: string; bid: number }>
  ): Promise<CalcuttaEntry> {
    return apiClient.patch<CalcuttaEntry>(`/entries/${entryId}`, {
      teams,
    });
  },

  async getPayouts(calcuttaId: string): Promise<{ payouts: Array<{ position: number; amountCents: number }> }> {
    return apiClient.get(`/calcuttas/${calcuttaId}/payouts`);
  },

  async replacePayouts(calcuttaId: string, payouts: Array<{ position: number; amountCents: number }>): Promise<{ payouts: Array<{ position: number; amountCents: number }> }> {
    return apiClient.put(`/calcuttas/${calcuttaId}/payouts`, { payouts });
  },
};
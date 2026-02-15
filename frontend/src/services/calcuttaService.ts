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

  async createEntry(calcuttaId: string, name: string, userId?: string): Promise<CalcuttaEntry> {
    return apiClient.post<CalcuttaEntry>(`/calcuttas/${calcuttaId}/entries`, {
      name,
      userId,
    });
  },

  async updateEntry(
    entryId: string,
    teams: Array<{ teamId: string; bid: number }>
  ): Promise<CalcuttaEntry> {
    return apiClient.patch<CalcuttaEntry>(`/entries/${entryId}`, {
      teams,
    });
  },
}; 
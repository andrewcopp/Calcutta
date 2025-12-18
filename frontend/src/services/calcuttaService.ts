import { Calcutta, CalcuttaEntry, CalcuttaEntryTeam, CalcuttaPortfolio, CalcuttaPortfolioTeam, School, TournamentTeam } from '../types/calcutta';

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';
const API_BASE_URL = `${API_URL}/api`;

export const calcuttaService = {
  async getAllCalcuttas(): Promise<Calcutta[]> {
    const response = await fetch(`${API_BASE_URL}/calcuttas`);
    if (!response.ok) {
      throw new Error('Failed to fetch calcuttas');
    }
    return response.json();
  },

  async getCalcutta(id: string): Promise<Calcutta> {
    const response = await fetch(`${API_BASE_URL}/calcuttas/${id}`);
    if (!response.ok) {
      throw new Error('Failed to fetch calcutta');
    }
    return response.json();
  },

  async createCalcutta(name: string, tournamentId: string, ownerId: string): Promise<Calcutta> {
    const response = await fetch(`${API_BASE_URL}/calcuttas`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        name,
        tournamentId,
        ownerId,
      }),
    });
    if (!response.ok) {
      throw new Error('Failed to create calcutta');
    }
    return response.json();
  },

  async getCalcuttaEntries(calcuttaId: string): Promise<CalcuttaEntry[]> {
    const response = await fetch(`${API_BASE_URL}/calcuttas/${calcuttaId}/entries`);
    if (!response.ok) {
      throw new Error('Failed to fetch calcutta entries');
    }
    return response.json();
  },

  async getEntryTeams(entryId: string, calcuttaId: string): Promise<CalcuttaEntryTeam[]> {
    const response = await fetch(`${API_BASE_URL}/calcuttas/${calcuttaId}/entries/${entryId}/teams`);
    if (!response.ok) {
      throw new Error('Failed to fetch entry teams');
    }
    return response.json();
  },

  async getSchools(): Promise<School[]> {
    const response = await fetch(`${API_BASE_URL}/schools`);
    if (!response.ok) {
      throw new Error('Failed to fetch schools');
    }
    return response.json();
  },

  async getTournamentTeams(tournamentId: string): Promise<TournamentTeam[]> {
    const response = await fetch(`${API_BASE_URL}/tournaments/${tournamentId}/teams`);
    if (!response.ok) {
      throw new Error('Failed to fetch tournament teams');
    }
    return response.json();
  },

  async getPortfoliosByEntry(entryId: string): Promise<CalcuttaPortfolio[]> {
    const response = await fetch(`${API_BASE_URL}/entries/${entryId}/portfolios`);
    if (!response.ok) {
      throw new Error('Failed to fetch portfolios');
    }
    return response.json();
  },

  async getPortfolioTeams(portfolioId: string): Promise<CalcuttaPortfolioTeam[]> {
    const response = await fetch(`${API_BASE_URL}/portfolios/${portfolioId}/teams`);
    if (!response.ok) {
      throw new Error('Failed to fetch portfolio teams');
    }
    return response.json();
  },

  async calculatePortfolioScores(portfolioId: string): Promise<void> {
    const response = await fetch(`${API_BASE_URL}/portfolios/${portfolioId}/calculate-scores`, {
      method: 'POST',
    });
    if (!response.ok) {
      throw new Error('Failed to calculate portfolio scores');
    }
  },

  async updatePortfolioTeamScores(
    portfolioId: string,
    teamId: string,
    expectedPoints: number,
    predictedPoints: number
  ): Promise<void> {
    const response = await fetch(`${API_BASE_URL}/portfolios/${portfolioId}/teams/${teamId}/scores`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        expectedPoints,
        predictedPoints,
      }),
    });
    if (!response.ok) {
      throw new Error('Failed to update portfolio team scores');
    }
  },

  async updatePortfolioMaximumScore(portfolioId: string, maximumPoints: number): Promise<void> {
    const response = await fetch(`${API_BASE_URL}/portfolios/${portfolioId}/maximum-score`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        maximumPoints,
      }),
    });
    if (!response.ok) {
      throw new Error('Failed to update portfolio maximum score');
    }
  },
}; 
import { Calcutta, CalcuttaEntry, CalcuttaEntryTeam } from '../types/calcutta';

const API_BASE_URL = 'http://localhost:8080/api';

export const calcuttaService = {
  async getAllCalcuttas(): Promise<Calcutta[]> {
    const response = await fetch(`${API_BASE_URL}/calcuttas`);
    if (!response.ok) {
      throw new Error('Failed to fetch calcuttas');
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
  }
}; 
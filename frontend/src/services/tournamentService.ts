import { Tournament, TournamentTeam } from '../types/calcutta';

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';
const API_BASE_URL = `${API_URL}/api`;

export const tournamentService = {
  async getAllTournaments(): Promise<Tournament[]> {
    const response = await fetch(`${API_BASE_URL}/tournaments`);
    if (!response.ok) {
      throw new Error('Failed to fetch tournaments');
    }
    return response.json();
  },

  async getTournament(id: string): Promise<Tournament> {
    const response = await fetch(`${API_BASE_URL}/tournaments/${id}`);
    if (!response.ok) {
      throw new Error('Failed to fetch tournament');
    }
    return response.json();
  },

  async getTournamentTeams(id: string): Promise<TournamentTeam[]> {
    const response = await fetch(`${API_BASE_URL}/tournaments/${id}/teams`);
    if (!response.ok) {
      throw new Error('Failed to fetch tournament teams');
    }
    return response.json();
  },

  async createTournament(name: string, rounds: number): Promise<Tournament> {
    const response = await fetch(`${API_BASE_URL}/tournaments`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ name, rounds }),
    });
    if (!response.ok) {
      throw new Error('Failed to create tournament');
    }
    return response.json();
  },

  async createTournamentTeam(
    tournamentId: string,
    schoolId: string,
    seed: number,
    region: string = 'Unknown'
  ): Promise<TournamentTeam> {
    const response = await fetch(`${API_BASE_URL}/tournaments/${tournamentId}/teams`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ schoolId, seed, region }),
    });
    if (!response.ok) {
      throw new Error('Failed to create tournament team');
    }
    return response.json();
  },

  async updateTournamentTeam(
    teamId: string,
    updates: Partial<TournamentTeam>
  ): Promise<TournamentTeam> {
    const response = await fetch(`${API_BASE_URL}/teams/${teamId}`, {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(updates),
    });
    if (!response.ok) {
      throw new Error('Failed to update tournament team');
    }
    return response.json();
  }
}; 
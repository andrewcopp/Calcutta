import { Tournament, TournamentTeam } from '../types/calcutta';
import { apiClient } from '../api/apiClient';

export const tournamentService = {
  async getAllTournaments(): Promise<Tournament[]> {
    return apiClient.get<Tournament[]>('/tournaments');
  },

  async getTournament(id: string): Promise<Tournament> {
    return apiClient.get<Tournament>(`/tournaments/${id}`);
  },

  async getTournamentTeams(id: string): Promise<TournamentTeam[]> {
    return apiClient.get<TournamentTeam[]>(`/tournaments/${id}/teams`);
  },

  async createTournament(name: string, rounds: number): Promise<Tournament> {
    return apiClient.post<Tournament>('/tournaments', { name, rounds });
  },

  async createTournamentTeam(
    tournamentId: string,
    schoolId: string,
    seed: number,
    region: string = 'Unknown'
  ): Promise<TournamentTeam> {
    return apiClient.post<TournamentTeam>(`/tournaments/${tournamentId}/teams`, { schoolId, seed, region });
  },

  async updateTournamentTeam(
    tournamentId: string,
    teamId: string,
    updates: Partial<TournamentTeam>
  ): Promise<TournamentTeam> {
    return apiClient.patch<TournamentTeam>(`/tournaments/${tournamentId}/teams/${teamId}`, updates);
  }
}; 
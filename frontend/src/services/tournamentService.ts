import { Tournament, TournamentTeam, TournamentModerator } from '../types/calcutta';
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
  },

  async getTournamentModerators(tournamentId: string): Promise<TournamentModerator[]> {
    const res = await apiClient.get<{ moderators: TournamentModerator[] }>(`/tournaments/${tournamentId}/moderators`);
    return res.moderators;
  },

  async grantTournamentModerator(tournamentId: string, email: string): Promise<TournamentModerator> {
    const res = await apiClient.post<{ moderator: TournamentModerator }>(`/tournaments/${tournamentId}/moderators`, { email });
    return res.moderator;
  },

  async revokeTournamentModerator(tournamentId: string, userId: string): Promise<void> {
    await apiClient.delete(`/tournaments/${tournamentId}/moderators/${userId}`);
  },
}; 
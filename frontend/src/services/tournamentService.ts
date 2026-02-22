import { Tournament, TournamentTeam, TournamentModerator, Competition, Season } from '../types/tournament';
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

  async createTournament(competition: string, year: number, rounds: number): Promise<Tournament> {
    return apiClient.post<Tournament>('/tournaments', { competition, year, rounds });
  },

  async updateTournament(
    id: string,
    updates: {
      startingAt?: string | null;
      finalFourTopLeft?: string;
      finalFourBottomLeft?: string;
      finalFourTopRight?: string;
      finalFourBottomRight?: string;
    }
  ): Promise<Tournament> {
    return apiClient.patch<Tournament>(`/tournaments/${id}`, updates);
  },

  async replaceTeams(
    tournamentId: string,
    teams: { schoolId: string; seed: number; region: string }[]
  ): Promise<TournamentTeam[]> {
    return apiClient.put<TournamentTeam[]>(`/tournaments/${tournamentId}/teams`, { teams });
  },

  async getCompetitions(): Promise<Competition[]> {
    return apiClient.get<Competition[]>('/competitions');
  },

  async getSeasons(): Promise<Season[]> {
    return apiClient.get<Season[]>('/seasons');
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

  async updateKenPomStats(
    tournamentId: string,
    stats: { teamId: string; netRtg: number; oRtg: number; dRtg: number; adjT: number }[]
  ): Promise<TournamentTeam[]> {
    return apiClient.put<TournamentTeam[]>(`/tournaments/${tournamentId}/kenpom`, { stats });
  },
};

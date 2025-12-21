import { TournamentTeam } from '../types/calcutta';
import { School } from '../types/school';
import { apiClient } from '../api/apiClient';

export async function fetchTournamentTeams(tournamentId: string): Promise<TournamentTeam[]> {
  return apiClient.get<TournamentTeam[]>(`/tournaments/${tournamentId}/teams`);
}

export async function updateTournamentTeam(tournamentId: string, teamId: string, updates: Partial<TournamentTeam>): Promise<TournamentTeam> {
  return apiClient.patch<TournamentTeam>(`/tournaments/${tournamentId}/teams/${teamId}`, updates);
}

export async function recalculatePortfolios(tournamentId: string): Promise<void> {
  await apiClient.post<void>(`/tournaments/${tournamentId}/recalculate-portfolios`);
}

interface CreateTournamentTeamData {
  schoolId: string;
  seed: number;
  region: string;
}

export async function createTournamentTeam(tournamentId: string, teamData: CreateTournamentTeamData): Promise<TournamentTeam> {
  return apiClient.post<TournamentTeam>(`/tournaments/${tournamentId}/teams`, teamData);
}

export const adminService = {
  async getAllSchools(): Promise<School[]> {
    return apiClient.get<School[]>('/schools');
  }
}; 
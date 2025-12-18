import { TournamentTeam } from '../types/calcutta';
import { School } from '../types/school';

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';
const API_BASE_URL = `${API_URL}/api`;

export async function fetchTournamentTeams(tournamentId: string): Promise<TournamentTeam[]> {
  const response = await fetch(`${API_BASE_URL}/tournaments/${tournamentId}/teams`, {
    headers: {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
    },
    credentials: 'include',
  });
  if (!response.ok) {
    throw new Error('Failed to fetch tournament teams');
  }
  return response.json();
}

export async function updateTournamentTeam(teamId: string, updates: Partial<TournamentTeam>): Promise<TournamentTeam> {
  const response = await fetch(`${API_BASE_URL}/teams/${teamId}`, {
    method: 'PATCH',
    headers: {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
    },
    credentials: 'include',
    body: JSON.stringify(updates),
  });
  
  if (!response.ok) {
    throw new Error('Failed to update team');
  }
  
  return response.json();
}

export async function recalculatePortfolios(tournamentId: string): Promise<void> {
  const response = await fetch(`${API_BASE_URL}/tournaments/${tournamentId}/recalculate-portfolios`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
    },
    credentials: 'include',
  });
  
  if (!response.ok) {
    throw new Error('Failed to recalculate portfolios');
  }
}

interface CreateTournamentTeamData {
  schoolId: string;
  seed: number;
  region: string;
}

export async function createTournamentTeam(tournamentId: string, teamData: CreateTournamentTeamData): Promise<TournamentTeam> {
  const response = await fetch(`${API_BASE_URL}/tournaments/${tournamentId}/teams`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
    },
    credentials: 'include',
    body: JSON.stringify(teamData),
  });
  
  if (!response.ok) {
    throw new Error('Failed to create tournament team');
  }
  
  return response.json();
}

export const adminService = {
  async getAllSchools(): Promise<School[]> {
    const response = await fetch(`${API_BASE_URL}/schools`);
    if (!response.ok) {
      throw new Error('Failed to fetch schools');
    }
    return response.json();
  }
}; 
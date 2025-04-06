import { TournamentTeam } from '../types/calcutta';

const API_BASE_URL = 'http://localhost:8080/api';

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
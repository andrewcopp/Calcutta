import { BracketStructure } from '../types/bracket';

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';
const API_BASE_URL = `${API_URL}/api`;

export async function fetchBracket(tournamentId: string): Promise<BracketStructure> {
  const response = await fetch(`${API_BASE_URL}/tournaments/${tournamentId}/bracket`, {
    headers: {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
    },
    credentials: 'include',
  });
  
  if (!response.ok) {
    throw new Error('Failed to fetch bracket');
  }
  
  return response.json();
}

export async function selectWinner(
  tournamentId: string,
  gameId: string,
  winnerTeamId: string
): Promise<BracketStructure> {
  const response = await fetch(
    `${API_BASE_URL}/tournaments/${tournamentId}/bracket/games/${gameId}/winner`,
    {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
      },
      credentials: 'include',
      body: JSON.stringify({ winnerTeamId }),
    }
  );
  
  if (!response.ok) {
    const error = await response.json().catch(() => ({ message: 'Failed to select winner' }));
    throw new Error(error.message || 'Failed to select winner');
  }
  
  return response.json();
}

export async function unselectWinner(
  tournamentId: string,
  gameId: string
): Promise<BracketStructure> {
  const response = await fetch(
    `${API_BASE_URL}/tournaments/${tournamentId}/bracket/games/${gameId}/winner`,
    {
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
      },
      credentials: 'include',
    }
  );
  
  if (!response.ok) {
    throw new Error('Failed to unselect winner');
  }
  
  return response.json();
}

export async function validateBracketSetup(tournamentId: string): Promise<{
  valid: boolean;
  errors: string[];
}> {
  const response = await fetch(
    `${API_BASE_URL}/tournaments/${tournamentId}/bracket/validate`,
    {
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
      },
      credentials: 'include',
    }
  );
  
  if (!response.ok) {
    throw new Error('Failed to validate bracket setup');
  }
  
  return response.json();
}

export const bracketService = {
  fetchBracket,
  selectWinner,
  unselectWinner,
  validateBracketSetup,
};

import { BracketStructure } from '../types/bracket';
import { apiClient } from '../api/apiClient';

export async function fetchBracket(tournamentId: string): Promise<BracketStructure> {
  return apiClient.get<BracketStructure>(`/tournaments/${tournamentId}/bracket`);
}

export async function selectWinner(
  tournamentId: string,
  gameId: string,
  winnerTeamId: string
): Promise<BracketStructure> {
  return apiClient.post<BracketStructure>(
    `/tournaments/${tournamentId}/bracket/games/${gameId}/winner`,
    { winnerTeamId }
  );
}

export async function unselectWinner(
  tournamentId: string,
  gameId: string
): Promise<BracketStructure> {
  return apiClient.delete<BracketStructure>(`/tournaments/${tournamentId}/bracket/games/${gameId}/winner`);
}

export async function validateBracketSetup(tournamentId: string): Promise<{
  valid: boolean;
  errors: string[];
}> {
  return apiClient.get<{ valid: boolean; errors: string[] }>(`/tournaments/${tournamentId}/bracket/validate`);
}

export const bracketService = {
  fetchBracket,
  selectWinner,
  unselectWinner,
  validateBracketSetup,
};

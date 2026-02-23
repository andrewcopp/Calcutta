import { BracketStructure } from '../types/bracket';
import { apiClient } from '../api/apiClient';

export const bracketService = {
  async fetchBracket(tournamentId: string): Promise<BracketStructure> {
    return apiClient.get<BracketStructure>(`/tournaments/${tournamentId}/bracket`);
  },

  async selectWinner(tournamentId: string, gameId: string, winnerTeamId: string): Promise<BracketStructure> {
    return apiClient.post<BracketStructure>(`/tournaments/${tournamentId}/bracket/games/${gameId}/winner`, {
      winnerTeamId,
    });
  },

  async unselectWinner(tournamentId: string, gameId: string): Promise<BracketStructure> {
    return apiClient.delete<BracketStructure>(`/tournaments/${tournamentId}/bracket/games/${gameId}/winner`);
  },

  async validateBracketSetup(tournamentId: string): Promise<{
    valid: boolean;
    errors: string[];
  }> {
    return apiClient.get<{ valid: boolean; errors: string[] }>(`/tournaments/${tournamentId}/bracket/validate`);
  },
};

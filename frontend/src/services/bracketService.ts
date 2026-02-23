import { BracketStructureSchema, BracketValidationSchema } from '../schemas/bracket';
import { apiClient } from '../api/apiClient';

export const bracketService = {
  async fetchBracket(tournamentId: string) {
    return apiClient.get(`/tournaments/${tournamentId}/bracket`, { schema: BracketStructureSchema });
  },

  async selectWinner(tournamentId: string, gameId: string, winnerTeamId: string) {
    return apiClient.post(`/tournaments/${tournamentId}/bracket/games/${gameId}/winner`, { winnerTeamId }, { schema: BracketStructureSchema });
  },

  async unselectWinner(tournamentId: string, gameId: string) {
    return apiClient.delete(`/tournaments/${tournamentId}/bracket/games/${gameId}/winner`, { schema: BracketStructureSchema });
  },

  async validateBracketSetup(tournamentId: string) {
    return apiClient.get(`/tournaments/${tournamentId}/bracket/validate`, { schema: BracketValidationSchema });
  },
};

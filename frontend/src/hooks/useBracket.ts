import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { BracketGame, BracketRound, ROUND_ORDER } from '../types/bracket';
import { bracketService } from '../services/bracketService';
import { queryKeys } from '../queryKeys';

function groupGamesByRound(games: BracketGame[]): Record<BracketRound, BracketGame[]> {
  const grouped = {} as Record<BracketRound, BracketGame[]>;

  ROUND_ORDER.forEach((round) => {
    grouped[round] = games
      .filter((game) => game.round === round)
      .sort((a, b) => a.sortOrder - b.sortOrder);
  });

  return grouped;
}

export function useBracket(tournamentId: string | undefined) {
  const [error, setError] = useState<string | null>(null);
  const queryClient = useQueryClient();

  const validationQuery = useQuery({
    queryKey: queryKeys.bracket.validation(tournamentId),
    enabled: Boolean(tournamentId),
    queryFn: () => bracketService.validateBracketSetup(tournamentId!),
  });

  const bracketQuery = useQuery({
    queryKey: queryKeys.bracket.detail(tournamentId),
    enabled: Boolean(tournamentId) && validationQuery.data?.valid === true,
    staleTime: 0,
    queryFn: () => bracketService.fetchBracket(tournamentId!),
  });

  const selectWinnerMutation = useMutation({
    mutationFn: async ({ gameId, teamId }: { gameId: string; teamId: string }) => {
      setError(null);
      return bracketService.selectWinner(tournamentId!, gameId, teamId);
    },
    onSuccess: (updated) => {
      queryClient.setQueryData(queryKeys.bracket.detail(tournamentId), updated);
    },
    onError: (err) => {
      setError(err instanceof Error ? err.message : 'Failed to select winner');
    },
  });

  const unselectWinnerMutation = useMutation({
    mutationFn: async ({ gameId }: { gameId: string }) => {
      setError(null);
      return bracketService.unselectWinner(tournamentId!, gameId);
    },
    onSuccess: (updated) => {
      queryClient.setQueryData(queryKeys.bracket.detail(tournamentId), updated);
    },
    onError: (err) => {
      setError(err instanceof Error ? err.message : 'Failed to unselect winner');
    },
  });

  const actionLoading = selectWinnerMutation.isPending || unselectWinnerMutation.isPending;

  const handleSelectWinner = (gameId: string, teamId: string) => {
    if (!tournamentId || actionLoading) return;
    selectWinnerMutation.mutate({ gameId, teamId });
  };

  const handleUnselectWinner = (gameId: string) => {
    if (!tournamentId || actionLoading) return;
    unselectWinnerMutation.mutate({ gameId });
  };

  const loadData = async () => {
    setError(null);
    const validationResult = await validationQuery.refetch();
    if (validationResult.data?.valid === true) {
      await bracketQuery.refetch();
    }
  };

  const validationErrors =
    validationQuery.data?.valid === false ? validationQuery.data.errors : [];

  const bracket = bracketQuery.data || null;
  const gamesByRound = bracket ? groupGamesByRound(bracket.games) : null;

  return {
    validationErrors,
    validationLoading: validationQuery.isLoading,
    bracket,
    bracketLoading: validationQuery.data?.valid === true && bracketQuery.isLoading,
    gamesByRound,
    error,
    actionLoading,
    handleSelectWinner,
    handleUnselectWinner,
    loadData,
  };
}

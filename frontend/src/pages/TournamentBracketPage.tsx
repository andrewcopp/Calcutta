import React, { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useParams, Link } from 'react-router-dom';
import { BracketStructure, BracketGame, ROUND_LABELS, ROUND_ORDER, BracketRound } from '../types/bracket';
import { Tournament } from '../types/calcutta';
import { bracketService } from '../services/bracketService';
import { tournamentService } from '../services/tournamentService';
import { BracketGameCard } from '../components/BracketGameCard';
import { queryKeys } from '../queryKeys';
import { Alert } from '../components/ui/Alert';

export const TournamentBracketPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const [error, setError] = useState<string | null>(null);
  const queryClient = useQueryClient();

  // Fetch tournament info
  const tournamentQuery = useQuery({
    queryKey: queryKeys.tournaments.detail(id),
    enabled: Boolean(id),
    staleTime: 30_000,
    queryFn: () => tournamentService.getTournament(id!),
  });

  // Validate bracket setup
  const validationQuery = useQuery({
    queryKey: queryKeys.bracket.validation(id),
    enabled: Boolean(id),
    staleTime: 30_000,
    queryFn: () => bracketService.validateBracketSetup(id!),
  });

  // Fetch bracket
  const bracketQuery = useQuery({
    queryKey: queryKeys.bracket.detail(id),
    enabled: Boolean(id) && validationQuery.data?.valid === true,
    staleTime: 0,
    queryFn: () => bracketService.fetchBracket(id!),
  });

  const selectWinnerMutation = useMutation({
    mutationFn: async ({ gameId, teamId }: { gameId: string; teamId: string }) => {
      setError(null);
      return bracketService.selectWinner(id!, gameId, teamId);
    },
    onSuccess: (updated) => {
      queryClient.setQueryData(queryKeys.bracket.detail(id), updated);
    },
    onError: (err) => {
      setError(err instanceof Error ? err.message : 'Failed to select winner');
    },
  });

  const unselectWinnerMutation = useMutation({
    mutationFn: async ({ gameId }: { gameId: string }) => {
      setError(null);
      return bracketService.unselectWinner(id!, gameId);
    },
    onSuccess: (updated) => {
      queryClient.setQueryData(queryKeys.bracket.detail(id), updated);
    },
    onError: (err) => {
      setError(err instanceof Error ? err.message : 'Failed to unselect winner');
    },
  });

  const loadData = async () => {
    setError(null);

    await tournamentQuery.refetch();
    const validationResult = await validationQuery.refetch();
    if (validationResult.data?.valid === true) {
      await bracketQuery.refetch();
    }
  };

  const handleSelectWinner = async (gameId: string, teamId: string) => {
    if (!id) return;
    if (selectWinnerMutation.isPending || unselectWinnerMutation.isPending) return;
    selectWinnerMutation.mutate({ gameId, teamId });
  };

  const handleUnselectWinner = async (gameId: string) => {
    if (!id) return;
    if (selectWinnerMutation.isPending || unselectWinnerMutation.isPending) return;
    unselectWinnerMutation.mutate({ gameId });
  };

  const groupGamesByRound = (games: BracketGame[]): Record<BracketRound, BracketGame[]> => {
    const grouped = {} as Record<BracketRound, BracketGame[]>;
    
    ROUND_ORDER.forEach(round => {
      grouped[round] = games
        .filter(game => game.round === round)
        .sort((a, b) => a.sortOrder - b.sortOrder);
    });
    
    return grouped;
  };

  if (!id) {
    return <Alert variant="error">Missing required parameters</Alert>;
  }

  if (tournamentQuery.isLoading || validationQuery.isLoading || (validationQuery.data?.valid === true && bracketQuery.isLoading)) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">Loading bracket...</p>
        </div>
      </div>
    );
  }

  const validationErrors = validationQuery.data?.valid === false ? validationQuery.data.errors : [];

  if (validationErrors.length > 0) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="mb-6">
          <Link to={`/admin/tournaments/${id}`} className="text-blue-600 hover:underline">
            ← Back to Tournament
          </Link>
        </div>
        
        <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-6">
          <h2 className="text-xl font-bold text-yellow-800 mb-4">
            Bracket Setup Incomplete
          </h2>
          <p className="text-yellow-700 mb-4">
            The tournament is not ready for bracket management. Please fix the following issues:
          </p>
          <ul className="list-disc list-inside space-y-2">
            {validationErrors.map((error, index) => (
              <li key={index} className="text-yellow-700">{error}</li>
            ))}
          </ul>
          <div className="mt-6">
            <Link
              to={`/admin/tournaments/${id}/teams/add`}
              className="inline-block px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
            >
              Add Teams
            </Link>
          </div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="bg-red-50 border border-red-200 rounded-lg p-6">
          <h2 className="text-xl font-bold text-red-800 mb-2">Error</h2>
          <p className="text-red-700">{error}</p>
          <button
            onClick={loadData}
            className="mt-4 px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
          >
            Retry
          </button>
        </div>
      </div>
    );
  }

  if (tournamentQuery.isError || validationQuery.isError || bracketQuery.isError) {
    const message =
      (tournamentQuery.error instanceof Error ? tournamentQuery.error.message : null) ||
      (validationQuery.error instanceof Error ? validationQuery.error.message : null) ||
      (bracketQuery.error instanceof Error ? bracketQuery.error.message : null) ||
      'Failed to load bracket';

    return (
      <div className="container mx-auto px-4 py-8">
        <div className="bg-red-50 border border-red-200 rounded-lg p-6">
          <h2 className="text-xl font-bold text-red-800 mb-2">Error</h2>
          <p className="text-red-700">{message}</p>
          <button
            onClick={loadData}
            className="mt-4 px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
          >
            Retry
          </button>
        </div>
      </div>
    );
  }

  const tournament: Tournament | null = tournamentQuery.data || null;
  const bracket: BracketStructure | null = bracketQuery.data || null;

  const actionLoading = selectWinnerMutation.isPending || unselectWinnerMutation.isPending;

  if (!bracket) {
    return (
      <div className="container mx-auto px-4 py-8">
        <p className="text-gray-600">No bracket data available</p>
      </div>
    );
  }

  const gamesByRound = groupGamesByRound(bracket.games);

  return (
    <div className="container mx-auto px-4 py-8">
      {/* Header */}
      <div className="mb-6">
        <Link to={`/admin/tournaments/${id}`} className="text-blue-600 hover:underline mb-4 inline-block">
          ← Back to Tournament
        </Link>
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold">{tournament?.name} - Bracket</h1>
            <p className="text-gray-600 mt-1">
              Click on a team to select them as the winner. Click again to undo.
            </p>
          </div>
          <button
            onClick={loadData}
            disabled={actionLoading}
            className="px-4 py-2 bg-gray-200 text-gray-700 rounded hover:bg-gray-300 disabled:opacity-50"
          >
            Refresh
          </button>
        </div>
      </div>

      {/* Action Loading Indicator */}
      {actionLoading && (
        <div className="mb-4 bg-blue-50 border border-blue-200 rounded-lg p-4">
          <div className="flex items-center gap-3">
            <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-blue-600"></div>
            <span className="text-blue-700">Updating bracket...</span>
          </div>
        </div>
      )}

      {/* Rounds */}
      <div className="space-y-8">
        {ROUND_ORDER.map(round => {
          const games = gamesByRound[round];
          if (games.length === 0) return null;

          return (
            <div key={round} className="bg-gray-50 rounded-lg p-6">
              <h2 className="text-2xl font-bold mb-4 text-gray-800">
                {ROUND_LABELS[round]}
                <span className="ml-3 text-sm font-normal text-gray-500">
                  ({games.length} {games.length === 1 ? 'game' : 'games'})
                </span>
              </h2>
              
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
                {games.map(game => (
                  <BracketGameCard
                    key={game.gameId}
                    game={game}
                    onSelectWinner={handleSelectWinner}
                    onUnselectWinner={handleUnselectWinner}
                    isLoading={actionLoading}
                  />
                ))}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};

export default TournamentBracketPage;

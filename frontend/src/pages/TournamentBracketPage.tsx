import React, { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useParams, useNavigate } from 'react-router-dom';
import { BracketStructure, BracketGame, ROUND_LABELS, ROUND_ORDER, BracketRound } from '../types/bracket';
import { Tournament } from '../types/calcutta';
import { bracketService } from '../services/bracketService';
import { tournamentService } from '../services/tournamentService';
import { BracketGameCard } from '../components/BracketGameCard';
import { queryKeys } from '../queryKeys';
import { Alert } from '../components/ui/Alert';
import { Breadcrumb } from '../components/ui/Breadcrumb';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';

export const TournamentBracketPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [error, setError] = useState<string | null>(null);
  const queryClient = useQueryClient();

  // Fetch tournament info
  const tournamentQuery = useQuery({
    queryKey: queryKeys.tournaments.detail(id),
    enabled: Boolean(id),
    queryFn: () => tournamentService.getTournament(id!),
  });

  // Validate bracket setup
  const validationQuery = useQuery({
    queryKey: queryKeys.bracket.validation(id),
    enabled: Boolean(id),
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
    return (
      <PageContainer>
        <Alert variant="error">Missing required parameters</Alert>
      </PageContainer>
    );
  }

  if (tournamentQuery.isLoading || validationQuery.isLoading || (validationQuery.data?.valid === true && bracketQuery.isLoading)) {
    return (
      <PageContainer>
        <LoadingState label="Loading bracket..." size="lg" />
      </PageContainer>
    );
  }

  const validationErrors = validationQuery.data?.valid === false ? validationQuery.data.errors : [];

  if (validationErrors.length > 0) {
    return (
      <PageContainer>
        <PageHeader
          title="Bracket"
          subtitle="Bracket setup incomplete"
          actions={
            <>
              <Button variant="ghost" onClick={() => navigate(`/admin/tournaments/${id}`)}>
                Back
              </Button>
              <Button onClick={() => navigate(`/admin/tournaments/${id}/teams/add`)}>Add Teams</Button>
            </>
          }
        />

        <Alert variant="warning" className="mb-4">
          The tournament is not ready for bracket management. Please fix the following issues:
        </Alert>

        <Card>
          <ul className="list-disc list-inside space-y-2">
            {validationErrors.map((error, index) => (
              <li key={index}>{error}</li>
            ))}
          </ul>
        </Card>
      </PageContainer>
    );
  }

  if (error) {
    return (
      <PageContainer>
        <Alert variant="error" className="mb-4">
          {error}
        </Alert>
        <Button onClick={loadData}>Retry</Button>
      </PageContainer>
    );
  }

  if (tournamentQuery.isError || validationQuery.isError || bracketQuery.isError) {
    const message =
      (tournamentQuery.error instanceof Error ? tournamentQuery.error.message : null) ||
      (validationQuery.error instanceof Error ? validationQuery.error.message : null) ||
      (bracketQuery.error instanceof Error ? bracketQuery.error.message : null) ||
      'Failed to load bracket';

    return (
      <PageContainer>
        <Alert variant="error" className="mb-4">
          {message}
        </Alert>
        <Button onClick={loadData}>Retry</Button>
      </PageContainer>
    );
  }

  const tournament: Tournament | null = tournamentQuery.data || null;
  const bracket: BracketStructure | null = bracketQuery.data || null;

  const actionLoading = selectWinnerMutation.isPending || unselectWinnerMutation.isPending;

  if (!bracket) {
    return (
      <PageContainer>
        <Alert variant="info">No bracket data available</Alert>
      </PageContainer>
    );
  }

  const gamesByRound = groupGamesByRound(bracket.games);

  return (
    <PageContainer>
      <Breadcrumb
        items={[
          { label: 'Tournaments', href: '/admin/tournaments' },
          { label: tournament?.name ?? 'Tournament', href: `/admin/tournaments/${id}` },
          { label: 'Bracket' },
        ]}
      />
      {/* Header */}
      <PageHeader
        title={`${tournament?.name ?? 'Tournament'} - Bracket`}
        subtitle="Click on a team to select them as the winner. Click again to undo."
        actions={
          <>
            <Button variant="ghost" onClick={() => navigate(`/admin/tournaments/${id}`)}>
              Back
            </Button>
            <Button variant="secondary" onClick={loadData} disabled={actionLoading}>
              Refresh
            </Button>
          </>
        }
      />

      {/* Action Loading Indicator */}
      {actionLoading ? (
        <Alert variant="info" className="mb-4">
          <LoadingState layout="inline" label="Updating bracket..." />
        </Alert>
      ) : null}

      {/* Rounds */}
      <div className="space-y-8">
        {ROUND_ORDER.map(round => {
          const games = gamesByRound[round];
          if (games.length === 0) return null;

          return (
            <Card key={round} className="bg-gray-50">
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
            </Card>
          );
        })}
      </div>
    </PageContainer>
  );
 };

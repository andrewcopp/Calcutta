import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useParams, Link } from 'react-router-dom';
import { ROUND_LABELS, ROUND_ORDER } from '../schemas/bracket';
import { tournamentService } from '../services/tournamentService';
import { queryKeys } from '../queryKeys';
import { useBracket } from '../hooks/useBracket';
import { BracketGameCard } from '../components/BracketGameCard';
import { Alert } from '../components/ui/Alert';
import { Breadcrumb } from '../components/ui/Breadcrumb';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { ErrorState } from '../components/ui/ErrorState';
import { Input } from '../components/ui/Input';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { formatDate, toDatetimeLocalValue } from '../utils/format';
import { ModeratorsSection } from './Tournament/ModeratorsSection';

export function TournamentViewPage() {
  const { id } = useParams<{ id: string }>();
  const queryClient = useQueryClient();
  const [editingStartTime, setEditingStartTime] = useState(false);
  const [startTimeValue, setStartTimeValue] = useState('');

  const startTimeMutation = useMutation({
    mutationFn: (startingAt: string) =>
      tournamentService.updateTournament(id!, { startingAt: new Date(startingAt).toISOString() }),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournaments.detail(id) });
      setEditingStartTime(false);
    },
  });

  const tournamentQuery = useQuery({
    queryKey: queryKeys.tournaments.detail(id),
    enabled: Boolean(id),
    queryFn: () => tournamentService.getTournament(id!),
  });

  const {
    validationErrors,
    validationLoading,
    bracketLoading,
    gamesByRound,
    error: bracketError,
    actionLoading,
    handleSelectWinner,
    handleUnselectWinner,
    loadData,
  } = useBracket(id);

  if (!id) {
    return (
      <PageContainer>
        <Alert variant="error">Missing tournament id</Alert>
      </PageContainer>
    );
  }

  if (tournamentQuery.isLoading || validationLoading) {
    return (
      <PageContainer>
        <LoadingState label="Loading tournament..." />
      </PageContainer>
    );
  }

  if (tournamentQuery.isError) {
    return (
      <PageContainer>
        <ErrorState error={tournamentQuery.error} onRetry={() => tournamentQuery.refetch()} />
      </PageContainer>
    );
  }

  const tournament = tournamentQuery.data || null;

  if (!tournament) {
    return (
      <PageContainer>
        <Alert variant="error">Tournament not found</Alert>
      </PageContainer>
    );
  }

  return (
    <PageContainer>
      <Breadcrumb items={[{ label: 'Tournaments', href: '/admin/tournaments' }, { label: tournament.name }]} />
      <PageHeader
        title={tournament.name}
        subtitle={`${tournament.rounds} rounds • ${tournament.startingAt ? `Starts ${formatDate(tournament.startingAt, true)}` : 'No start time set'}`}
        actions={
          <div className="flex gap-2">
            <Link to={`/admin/tournaments/${id}/teams/setup`}>
              <Button variant="secondary">Edit Field</Button>
            </Link>
            <Button variant="outline" onClick={loadData} disabled={actionLoading}>
              Refresh
            </Button>
          </div>
        }
      />

      <Card>
        <div className="flex items-center gap-4">
          <span className="text-sm font-medium text-foreground">Start Time</span>
          {editingStartTime ? (
            <>
              <Input
                type="datetime-local"
                value={startTimeValue}
                onChange={(e) => setStartTimeValue(e.target.value)}
                className="w-auto"
              />
              <Button
                size="sm"
                disabled={!startTimeValue || startTimeMutation.isPending}
                loading={startTimeMutation.isPending}
                onClick={() => startTimeMutation.mutate(startTimeValue)}
              >
                Save
              </Button>
              <Button size="sm" variant="outline" onClick={() => setEditingStartTime(false)}>
                Cancel
              </Button>
            </>
          ) : (
            <>
              <span className="text-sm text-muted-foreground">
                {tournament.startingAt ? formatDate(tournament.startingAt, true) : 'Not set'}
              </span>
              <Button
                size="sm"
                variant="outline"
                onClick={() => {
                  setStartTimeValue(tournament.startingAt ? toDatetimeLocalValue(tournament.startingAt) : '');
                  setEditingStartTime(true);
                }}
              >
                {tournament.startingAt ? 'Change' : 'Set'}
              </Button>
            </>
          )}
        </div>
        {startTimeMutation.isError && (
          <Alert variant="error" className="mt-2">
            {startTimeMutation.error instanceof Error ? startTimeMutation.error.message : 'Failed to update start time'}
          </Alert>
        )}
      </Card>

      <ModeratorsSection tournamentId={id} />

      {validationErrors.length > 0 && (
        <Alert variant="warning">
          <div className="font-semibold mb-1">Bracket setup incomplete</div>
          <ul className="list-disc list-inside text-sm space-y-1">
            {validationErrors.map((err, i) => (
              <li key={i}>{err}</li>
            ))}
          </ul>
          <div className="mt-2">
            <Link to={`/admin/tournaments/${id}/teams/setup`}>
              <Button size="sm" variant="outline">
                Setup Teams
              </Button>
            </Link>
          </div>
        </Alert>
      )}

      {bracketError && <Alert variant="error">{bracketError}</Alert>}

      {bracketLoading && <LoadingState label="Loading bracket..." />}

      {actionLoading && (
        <Alert variant="info">
          <LoadingState layout="inline" label="Updating bracket..." />
        </Alert>
      )}

      {gamesByRound && (
        <div className="space-y-8">
          {ROUND_ORDER.map((round) => {
            const games = gamesByRound[round];
            if (games.length === 0) return null;

            return (
              <Card key={round} className="bg-accent">
                <h2 className="text-2xl font-bold mb-4 text-foreground">
                  {ROUND_LABELS[round]}
                  <span className="ml-3 text-sm font-normal text-muted-foreground">
                    ({games.length} {games.length === 1 ? 'game' : 'games'})
                  </span>
                </h2>

                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
                  {games.map((game) => (
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
      )}
    </PageContainer>
  );
}

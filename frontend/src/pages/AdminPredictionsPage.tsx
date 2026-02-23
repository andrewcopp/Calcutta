import { useState, useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { tournamentService } from '../services/tournamentService';
import { queryKeys } from '../queryKeys';
import { TeamPrediction } from '../schemas/tournament';
import { Breadcrumb } from '../components/ui/Breadcrumb';
import { Card } from '../components/ui/Card';
import { Select } from '../components/ui/Select';
import { LoadingState } from '../components/ui/LoadingState';
import { ErrorState } from '../components/ui/ErrorState';
import { PageContainer, PageHeader } from '../components/ui/Page';

const ROUND_HEADERS = ['R1', 'R2', 'R3', 'R4', 'R5', 'R6', 'R7'] as const;

function getDisplayProbability(team: TeamPrediction, round: number): { value: number; style: string } {
  const progress = team.wins + team.byes;
  if (progress >= round) {
    return { value: 1, style: 'text-green-600 font-semibold' };
  }
  if (team.isEliminated) {
    return { value: 0, style: 'text-muted-foreground' };
  }
  const pKey = `pRound${round}` as keyof TeamPrediction;
  const raw = team[pKey] as number;
  return { value: raw, style: '' };
}

function formatPercent(value: number): string {
  if (value === 1) return '100%';
  if (value === 0) return '0%';
  return `${(value * 100).toFixed(1)}%`;
}

function formatBatchLabel(createdAt: string, isLatest: boolean): string {
  const date = new Date(createdAt);
  const label = date.toLocaleString(undefined, {
    month: 'short',
    day: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
  });
  return isLatest ? `${label} (latest)` : label;
}

export function AdminPredictionsPage() {
  const [selectedTournamentId, setSelectedTournamentId] = useState<string>('');
  const [selectedBatchId, setSelectedBatchId] = useState<string>('');

  const tournamentsQuery = useQuery({
    queryKey: queryKeys.tournaments.all(),
    queryFn: () => tournamentService.getAllTournaments(),
  });

  const batchesQuery = useQuery({
    queryKey: queryKeys.tournaments.predictionBatches(selectedTournamentId || undefined),
    queryFn: () => tournamentService.getPredictionBatches(selectedTournamentId),
    enabled: !!selectedTournamentId,
  });

  // Auto-select latest batch when batches load or tournament changes
  useEffect(() => {
    if (batchesQuery.data && batchesQuery.data.length > 0) {
      setSelectedBatchId(batchesQuery.data[0].id);
    } else {
      setSelectedBatchId('');
    }
  }, [batchesQuery.data]);

  // Reset batch when tournament changes
  useEffect(() => {
    setSelectedBatchId('');
  }, [selectedTournamentId]);

  const predictionsQuery = useQuery({
    queryKey: queryKeys.tournaments.predictions(selectedTournamentId || undefined, selectedBatchId || undefined),
    queryFn: () => tournamentService.getTournamentPredictions(selectedTournamentId, selectedBatchId || undefined),
    enabled: !!selectedTournamentId && !!selectedBatchId,
  });

  const teams = predictionsQuery.data?.teams ?? [];
  const sortedTeams = [...teams].sort((a, b) => {
    if (a.region < b.region) return -1;
    if (a.region > b.region) return 1;
    return a.seed - b.seed;
  });

  return (
    <PageContainer>
      <Breadcrumb items={[{ label: 'Admin', href: '/admin' }, { label: 'Predictions' }]} />
      <PageHeader title="Tournament Predictions" subtitle="Round-by-round advancement probabilities across prediction snapshots" />

      <Card>
        <div className="flex flex-col sm:flex-row gap-4 mb-6">
          <div className="flex-1">
            <label htmlFor="tournament-select" className="block text-sm font-medium text-foreground mb-2">
              Tournament
            </label>
            {tournamentsQuery.isLoading ? (
              <LoadingState />
            ) : tournamentsQuery.isError ? (
              <ErrorState error="Failed to load tournaments" onRetry={() => tournamentsQuery.refetch()} />
            ) : (
              <Select
                id="tournament-select"
                value={selectedTournamentId}
                onChange={(e) => setSelectedTournamentId(e.target.value)}
              >
                <option value="">Select a tournament...</option>
                {tournamentsQuery.data?.map((t) => (
                  <option key={t.id} value={t.id}>
                    {t.name}
                  </option>
                ))}
              </Select>
            )}
          </div>

          {selectedTournamentId && (
            <div className="flex-1">
              <label htmlFor="batch-select" className="block text-sm font-medium text-foreground mb-2">
                Prediction Batch
              </label>
              {batchesQuery.isLoading ? (
                <LoadingState />
              ) : batchesQuery.isError ? (
                <ErrorState error="Failed to load batches" onRetry={() => batchesQuery.refetch()} />
              ) : (batchesQuery.data?.length ?? 0) === 0 ? (
                <p className="text-sm text-muted-foreground">No prediction batches available</p>
              ) : (
                <Select
                  id="batch-select"
                  value={selectedBatchId}
                  onChange={(e) => setSelectedBatchId(e.target.value)}
                >
                  {batchesQuery.data?.map((batch, i) => (
                    <option key={batch.id} value={batch.id}>
                      {formatBatchLabel(batch.createdAt, i === 0)}
                    </option>
                  ))}
                </Select>
              )}
            </div>
          )}
        </div>

        {selectedTournamentId && selectedBatchId && predictionsQuery.isLoading && <LoadingState />}
        {selectedTournamentId && selectedBatchId && predictionsQuery.isError && (
          <ErrorState error="Failed to load predictions" onRetry={() => predictionsQuery.refetch()} />
        )}

        {sortedTeams.length > 0 && (
          <div className="overflow-x-auto">
            <table className="min-w-full text-sm">
              <thead>
                <tr className="border-b text-left text-muted-foreground">
                  <th className="px-3 py-2 font-medium">Region</th>
                  <th className="px-3 py-2 font-medium">Seed</th>
                  <th className="px-3 py-2 font-medium">School</th>
                  {ROUND_HEADERS.map((label) => (
                    <th key={label} className="px-2 py-2 font-medium text-right whitespace-nowrap">
                      {label}
                    </th>
                  ))}
                  <th className="px-3 py-2 font-medium text-right">E[Pts]</th>
                  <th className="px-3 py-2 font-medium text-center">Status</th>
                </tr>
              </thead>
              <tbody>
                {sortedTeams.map((team) => {
                  const progress = team.wins + team.byes;
                  let status = '';
                  if (team.isEliminated) {
                    status = 'Eliminated';
                  } else if (progress > 0) {
                    status = `W${team.wins}` + (team.byes > 0 ? ` B${team.byes}` : '');
                  }
                  return (
                    <tr key={team.teamId} className="border-b">
                      <td className="px-3 py-2">{team.region}</td>
                      <td className="px-3 py-2">{team.seed}</td>
                      <td className="px-3 py-2">{team.schoolName || '—'}</td>
                      {ROUND_HEADERS.map((_, i) => {
                        const round = i + 1;
                        const { value, style } = getDisplayProbability(team, round);
                        return (
                          <td key={round} className={`px-2 py-2 text-right tabular-nums ${style}`}>
                            {formatPercent(value)}
                          </td>
                        );
                      })}
                      <td className="px-3 py-2 text-right tabular-nums font-medium">
                        {team.expectedPoints.toFixed(1)}
                      </td>
                      <td className="px-3 py-2 text-center text-muted-foreground text-xs">{status}</td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </Card>
    </PageContainer>
  );
}

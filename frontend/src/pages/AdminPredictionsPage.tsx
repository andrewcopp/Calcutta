import { useState } from 'react';
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

const ROUNDS = [
  { key: 1, label: 'First Four' },
  { key: 2, label: 'First Round' },
  { key: 3, label: 'Second Round' },
  { key: 4, label: 'Sweet 16' },
  { key: 5, label: 'Elite 8' },
  { key: 6, label: 'Final Four' },
  { key: 7, label: 'Championship' },
] as const;

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

export function AdminPredictionsPage() {
  const [selectedTournamentId, setSelectedTournamentId] = useState<string>('');
  const [selectedRound, setSelectedRound] = useState(2);

  const tournamentsQuery = useQuery({
    queryKey: queryKeys.tournaments.all(),
    queryFn: () => tournamentService.getAllTournaments(),
  });

  const predictionsQuery = useQuery({
    queryKey: queryKeys.tournaments.predictions(selectedTournamentId || undefined),
    queryFn: () => tournamentService.getTournamentPredictions(selectedTournamentId),
    enabled: !!selectedTournamentId,
  });

  const teams = predictionsQuery.data?.teams ?? [];
  const sortedTeams = [...teams].sort((a, b) => {
    const aProb = getDisplayProbability(a, selectedRound).value;
    const bProb = getDisplayProbability(b, selectedRound).value;
    if (bProb !== aProb) return bProb - aProb;
    if (a.region < b.region) return -1;
    if (a.region > b.region) return 1;
    return a.seed - b.seed;
  });

  return (
    <PageContainer>
      <Breadcrumb items={[{ label: 'Admin', href: '/admin' }, { label: 'Predictions' }]} />
      <PageHeader title="Tournament Predictions" subtitle="Pre-tournament advancement probabilities with realized results" />

      <Card>
        <div className="mb-6">
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

        {selectedTournamentId && predictionsQuery.isLoading && <LoadingState />}
        {selectedTournamentId && predictionsQuery.isError && (
          <ErrorState error="Failed to load predictions" onRetry={() => predictionsQuery.refetch()} />
        )}

        {sortedTeams.length > 0 && (
          <>
            <div className="flex flex-wrap gap-2 mb-4">
              {ROUNDS.map((round) => (
                <button
                  key={round.key}
                  onClick={() => setSelectedRound(round.key)}
                  className={`px-3 py-1.5 rounded-md text-sm font-medium transition-colors ${
                    selectedRound === round.key
                      ? 'bg-primary text-primary-foreground'
                      : 'bg-muted text-muted-foreground hover:bg-muted/80'
                  }`}
                >
                  {round.label}
                </button>
              ))}
            </div>

            <div className="overflow-x-auto">
              <table className="min-w-full text-sm">
                <thead>
                  <tr className="border-b text-left text-muted-foreground">
                    <th className="px-3 py-2 font-medium">Region</th>
                    <th className="px-3 py-2 font-medium">Seed</th>
                    <th className="px-3 py-2 font-medium">School</th>
                    <th className="px-3 py-2 font-medium text-right">Advancement %</th>
                    <th className="px-3 py-2 font-medium text-center">Status</th>
                  </tr>
                </thead>
                <tbody>
                  {sortedTeams.map((team) => {
                    const { value, style } = getDisplayProbability(team, selectedRound);
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
                        <td className={`px-3 py-2 text-right ${style}`}>{formatPercent(value)}</td>
                        <td className="px-3 py-2 text-center text-muted-foreground text-xs">{status}</td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          </>
        )}
      </Card>
    </PageContainer>
  );
}

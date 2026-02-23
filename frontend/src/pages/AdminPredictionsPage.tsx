import { useState, useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { tournamentService } from '../services/tournamentService';
import { queryKeys } from '../queryKeys';
import { Breadcrumb } from '../components/ui/Breadcrumb';
import { Card } from '../components/ui/Card';
import { Select } from '../components/ui/Select';
import { LoadingState } from '../components/ui/LoadingState';
import { ErrorState } from '../components/ui/ErrorState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { getConditionalProbability, formatPercent } from '../utils/predictions';

const ROUND_ABBREVS = ['FF', 'R64', 'R32', 'S16', 'E8', 'F4', 'NCG'] as const;

const THROUGH_ROUND_LABELS: Record<number, string> = {
  0: 'Pre-tournament',
  1: 'After First Four',
  2: 'After Round of 64',
  3: 'After Round of 32',
  4: 'After Sweet 16',
  5: 'After Elite 8',
  6: 'After Final Four',
  7: 'After Championship',
};

export function AdminPredictionsPage() {
  const [selectedTournamentId, setSelectedTournamentId] = useState<string>('');
  const [throughRound, setThroughRound] = useState<number | null>(null);
  const [sortRound, setSortRound] = useState(7);

  const tournamentsQuery = useQuery({
    queryKey: queryKeys.tournaments.all(),
    queryFn: () => tournamentService.getAllTournaments(),
  });

  const predictionsQuery = useQuery({
    queryKey: queryKeys.tournaments.predictions(selectedTournamentId || undefined),
    queryFn: () => tournamentService.getTournamentPredictions(selectedTournamentId),
    enabled: !!selectedTournamentId,
  });

  const teams = useMemo(() => predictionsQuery.data?.teams ?? [], [predictionsQuery.data]);

  // Compute current max progress to determine default throughRound and available options
  const maxProgress = useMemo(
    () => teams.reduce((max, t) => Math.max(max, t.wins + t.byes), 0),
    [teams],
  );

  // Auto-set throughRound when data loads
  const effectiveThroughRound = throughRound ?? maxProgress;

  // Build through-round options: 0 through maxProgress
  const throughRoundOptions = useMemo(() => {
    const opts: { label: string; value: number }[] = [];
    for (let i = 0; i <= maxProgress; i++) {
      opts.push({ label: THROUGH_ROUND_LABELS[i] ?? `After Round ${i}`, value: i });
    }
    return opts;
  }, [maxProgress]);

  // Compute round sums for surviving teams to normalize conditional probabilities
  const roundSums = useMemo(() => {
    if (effectiveThroughRound === 0) return undefined;
    const sums: Record<number, number> = {};
    const survivors = teams.filter((t) => t.wins + t.byes >= effectiveThroughRound);
    for (let round = 1; round <= 7; round++) {
      const pKey = `pRound${round}` as keyof typeof survivors[0];
      sums[round] = survivors.reduce((sum, t) => sum + (t[pKey] as number), 0);
    }
    return sums;
  }, [teams, effectiveThroughRound]);

  const sortedTeams = useMemo(() => {
    return [...teams].sort((a, b) => {
      const aProb = getConditionalProbability(a, sortRound, effectiveThroughRound, roundSums);
      const bProb = getConditionalProbability(b, sortRound, effectiveThroughRound, roundSums);
      if (bProb.value !== aProb.value) return bProb.value - aProb.value;
      if (a.region < b.region) return -1;
      if (a.region > b.region) return 1;
      return a.seed - b.seed;
    });
  }, [teams, sortRound, effectiveThroughRound, roundSums]);

  return (
    <PageContainer>
      <Breadcrumb items={[{ label: 'Admin', href: '/admin' }, { label: 'Predictions' }]} />
      <PageHeader title="Tournament Predictions" subtitle="Conditional advancement probabilities by tournament progression" />

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
                onChange={(e) => {
                  setSelectedTournamentId(e.target.value);
                  setThroughRound(null);
                }}
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

          {teams.length > 0 && (
            <div className="flex-1">
              <label htmlFor="round-select" className="block text-sm font-medium text-foreground mb-2">
                Through Round
              </label>
              <Select
                id="round-select"
                value={effectiveThroughRound}
                onChange={(e) => setThroughRound(Number(e.target.value))}
              >
                {throughRoundOptions.map((opt) => (
                  <option key={opt.value} value={opt.value}>
                    {opt.label}
                  </option>
                ))}
              </Select>
            </div>
          )}
        </div>

        {selectedTournamentId && predictionsQuery.isLoading && <LoadingState />}
        {selectedTournamentId && predictionsQuery.isError && (
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
                  {ROUND_ABBREVS.map((label, i) => {
                    const round = i + 1;
                    const isActive = sortRound === round;
                    return (
                      <th
                        key={label}
                        className={`px-2 py-2 font-medium text-right whitespace-nowrap cursor-pointer select-none hover:text-foreground ${isActive ? 'text-foreground underline' : ''}`}
                        onClick={() => setSortRound(round)}
                      >
                        {label}{isActive ? ' \u2193' : ''}
                      </th>
                    );
                  })}
                </tr>
              </thead>
              <tbody>
                {sortedTeams.map((team) => (
                  <tr key={team.teamId} className="border-b">
                    <td className="px-3 py-2">{team.region}</td>
                    <td className="px-3 py-2">{team.seed}</td>
                    <td className="px-3 py-2">{team.schoolName || '—'}</td>
                    {ROUND_ABBREVS.map((_, i) => {
                      const round = i + 1;
                      const { value, style } = getConditionalProbability(team, round, effectiveThroughRound, roundSums);
                      return (
                        <td key={round} className={`px-2 py-2 text-right tabular-nums ${style}`}>
                          {formatPercent(value)}
                        </td>
                      );
                    })}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </Card>
    </PageContainer>
  );
}

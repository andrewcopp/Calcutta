import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { Tournament } from '../types/tournament';
import { calcuttaService } from '../services/calcuttaService';
import { tournamentService } from '../services/tournamentService';
import { queryKeys } from '../queryKeys';
import { Alert } from '../components/ui/Alert';
import { Breadcrumb } from '../components/ui/Breadcrumb';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { Input } from '../components/ui/Input';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { Select } from '../components/ui/Select';

const ROUND_LABELS: Record<number, string> = {
  1: 'First Four Win',
  2: 'Round of 64 Win',
  3: 'Round of 32 Win',
  4: 'Sweet 16 Win',
  5: 'Elite 8 Win',
  6: 'Final Four Win',
  7: 'Championship Win',
};

const DEFAULT_POINTS = [0, 50, 100, 150, 200, 250, 300];

interface ScoringRule {
  winIndex: number;
  pointsAwarded: number;
}

function buildDefaultScoringRules(roundCount: number): ScoringRule[] {
  return Array.from({ length: roundCount }, (_, i) => ({
    winIndex: i + 1,
    pointsAwarded: DEFAULT_POINTS[i] ?? (i + 1) * 50,
  }));
}

export function CreateCalcuttaPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [error, setError] = useState<string | null>(null);
  const [name, setName] = useState('');
  const [tournamentId, setTournamentId] = useState('');
  const [scoringRules, setScoringRules] = useState<ScoringRule[]>([]);
  const [minTeams, setMinTeams] = useState(3);
  const [maxTeams, setMaxTeams] = useState(10);
  const [maxBid, setMaxBid] = useState(50);

  const tournamentsQuery = useQuery({
    queryKey: queryKeys.tournaments.all(),
    queryFn: () => tournamentService.getAllTournaments(),
  });

  const createCalcuttaMutation = useMutation({
    mutationFn: async (params: { name: string; tournamentId: string; scoringRules: ScoringRule[]; minTeams: number; maxTeams: number; maxBid: number }) => {
      return calcuttaService.createCalcutta(params.name, params.tournamentId, params.scoringRules, params.minTeams, params.maxTeams, params.maxBid);
    },
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: queryKeys.calcuttas.all() }),
        queryClient.invalidateQueries({ queryKey: ['calcuttasWithRankings'] }),
      ]);
      navigate('/calcuttas');
    },
    onError: (error) => {
      setError(error instanceof Error ? error.message : 'Failed to create calcutta');
    },
  });

  const handleTournamentChange = (newTournamentId: string) => {
    setTournamentId(newTournamentId);
    if (!newTournamentId) {
      setScoringRules([]);
      return;
    }
    const tournament = tournaments.find((t) => t.id === newTournamentId);
    const roundCount = tournament?.rounds ?? 7;
    setScoringRules(buildDefaultScoringRules(roundCount));
  };

  const handlePointsChange = (winIndex: number, value: string) => {
    const points = parseInt(value, 10);
    if (isNaN(points) || points < 0) return;
    setScoringRules((prev) =>
      prev.map((rule) => (rule.winIndex === winIndex ? { ...rule, pointsAwarded: points } : rule))
    );
  };

  const handleCreateCalcutta = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    createCalcuttaMutation.mutate({
      name,
      tournamentId,
      scoringRules,
      minTeams,
      maxTeams,
      maxBid,
    });
  };

  if (tournamentsQuery.isLoading) {
    return (
      <PageContainer>
        <LoadingState label="Loading tournaments..." />
      </PageContainer>
    );
  }

  if (tournamentsQuery.isError) {
    const message = tournamentsQuery.error instanceof Error ? tournamentsQuery.error.message : 'Failed to fetch tournaments';
    return (
      <PageContainer>
        <Alert variant="error">{message}</Alert>
      </PageContainer>
    );
  }

  const tournaments: Tournament[] = tournamentsQuery.data || [];

  return (
    <PageContainer>
      <div className="max-w-2xl mx-auto">
        <Breadcrumb
          items={[
            { label: 'Calcuttas', href: '/calcuttas' },
            { label: 'Create' },
          ]}
        />
        <PageHeader
          title="Create New Calcutta"
          actions={
            <Button variant="ghost" onClick={() => navigate('/calcuttas')}>
              ← Back to Calcuttas
            </Button>
          }
        />

        {error ? <Alert variant="error" className="mb-4">{error}</Alert> : null}

        <Card>
          <form onSubmit={handleCreateCalcutta}>
            <div className="space-y-6">
              <div>
                <label htmlFor="name" className="block text-sm font-medium text-gray-700 mb-1">
                  Calcutta Name
                </label>
                <Input
                  type="text"
                  id="name"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="Enter a name for your Calcutta"
                  required
                />
              </div>

              <div>
                <label htmlFor="tournament" className="block text-sm font-medium text-gray-700 mb-1">
                  Tournament
                </label>
                <Select
                  id="tournament"
                  value={tournamentId}
                  onChange={(e) => handleTournamentChange(e.target.value)}
                  required
                >
                  <option value="">Select a tournament</option>
                  {tournaments.map((tournament) => (
                    <option key={tournament.id} value={tournament.id}>
                      {tournament.name}
                    </option>
                  ))}
                </Select>
                <p className="mt-1 text-sm text-gray-500">Select the tournament this Calcutta will be based on</p>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-3">
                  Pool Settings
                </label>
                <div className="space-y-3">
                  <div className="flex items-center gap-3">
                    <label htmlFor="minTeams" className="text-sm text-gray-600 w-44 shrink-0">
                      Min Teams per Entry
                    </label>
                    <Input
                      type="number"
                      id="minTeams"
                      min={1}
                      max={68}
                      value={minTeams}
                      onChange={(e) => setMinTeams(parseInt(e.target.value, 10) || 0)}
                      className="w-28"
                    />
                  </div>
                  <div className="flex items-center gap-3">
                    <label htmlFor="maxTeams" className="text-sm text-gray-600 w-44 shrink-0">
                      Max Teams per Entry
                    </label>
                    <Input
                      type="number"
                      id="maxTeams"
                      min={1}
                      max={68}
                      value={maxTeams}
                      onChange={(e) => setMaxTeams(parseInt(e.target.value, 10) || 0)}
                      className="w-28"
                    />
                  </div>
                  <div className="flex items-center gap-3">
                    <label htmlFor="maxBid" className="text-sm text-gray-600 w-44 shrink-0">
                      Max Bid per Team
                    </label>
                    <Input
                      type="number"
                      id="maxBid"
                      min={1}
                      value={maxBid}
                      onChange={(e) => setMaxBid(parseInt(e.target.value, 10) || 0)}
                      className="w-28"
                    />
                    <span className="text-sm text-gray-500">pts</span>
                  </div>
                </div>
                <p className="mt-2 text-sm text-gray-500">Constraints for each participant's entry</p>
              </div>

              {scoringRules.length > 0 ? (
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-3">
                    Scoring Rules
                  </label>
                  <div className="space-y-3">
                    {scoringRules.map((rule) => (
                      <div key={rule.winIndex} className="flex items-center gap-3">
                        <label
                          htmlFor={`scoring-${rule.winIndex}`}
                          className="text-sm text-gray-600 w-44 shrink-0"
                        >
                          {ROUND_LABELS[rule.winIndex] ?? `Win ${rule.winIndex}`}
                        </label>
                        <Input
                          type="number"
                          id={`scoring-${rule.winIndex}`}
                          min={0}
                          value={rule.pointsAwarded}
                          onChange={(e) => handlePointsChange(rule.winIndex, e.target.value)}
                          className="w-28"
                        />
                        <span className="text-sm text-gray-500">pts</span>
                      </div>
                    ))}
                  </div>
                  <p className="mt-2 text-sm text-gray-500">Points awarded for each win in the tournament</p>
                </div>
              ) : null}

              <div className="pt-2">
                <Button type="submit" className="w-full" loading={createCalcuttaMutation.isPending}>
                  Create Calcutta
                </Button>
              </div>
            </div>
          </form>
        </Card>
      </div>
    </PageContainer>
  );
}

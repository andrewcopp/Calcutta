import React, { useEffect, useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link, useParams, useSearchParams } from 'react-router-dom';

import { Card } from '../components/ui/Card';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { Select } from '../components/ui/Select';
import { analyticsService } from '../services/analyticsService';

type Algorithm = {
  id: string;
  kind: string;
  name: string;
  description?: string | null;
  params_json?: unknown;
  created_at?: string;
};

type GameOutcomeRun = {
  id: string;
  algorithm_id: string;
  tournament_id: string;
  params_json?: unknown;
  git_sha?: string | null;
  created_at: string;
};

type AlgorithmCoverageTournament = {
  tournament_id: string;
  tournament_name: string;
  starting_at?: string | null;
  last_run_at?: string | null;
};

type AlgorithmCoverageDetailResponse = {
  algorithm: {
    id: string;
    name: string;
    description?: string | null;
  };
  covered: number;
  total: number;
  items: AlgorithmCoverageTournament[];
};

type TeamPredictedAdvancement = {
  team_id: string;
  school_name: string;
  seed: number;
  region: string;
  prob_pi: number;
  reach_r64: number;
  reach_r32: number;
  reach_s16: number;
  reach_e8: number;
  reach_ff: number;
  reach_champ: number;
  win_champ: number;
};

export function LabAdvancementAlgorithmDetailPage() {
  const { algorithmId } = useParams<{ algorithmId: string }>();
  const [searchParams, setSearchParams] = useSearchParams();
  const [selectedTournamentId, setSelectedTournamentId] = useState<string>(() => searchParams.get('tournamentId') || '');
  const [selectedGameOutcomeRunId, setSelectedGameOutcomeRunId] = useState<string>(() => searchParams.get('gameOutcomeRunId') || '');

  const algorithmsQuery = useQuery<{ items: Algorithm[] } | null>({
    queryKey: ['analytics', 'algorithms', 'game_outcomes'],
    queryFn: async () => {
      const filtered = await analyticsService.listAlgorithms<{ items: Algorithm[] }>('game_outcomes');
      if (filtered?.items?.length) return filtered;
      return analyticsService.listAlgorithms<{ items: Algorithm[] }>();
    },
  });

  const algorithm = useMemo(() => {
    const items = algorithmsQuery.data?.items ?? [];
    return items.find((a) => a.id === algorithmId) ?? null;
  }, [algorithmsQuery.data?.items, algorithmId]);

  const coverageDetailQuery = useQuery<AlgorithmCoverageDetailResponse | null>({
    queryKey: ['analytics', 'coverage-detail', 'game-outcomes', algorithmId],
    queryFn: async () => {
      if (!algorithmId) return null;
      return analyticsService.getGameOutcomesAlgorithmCoverageDetail<AlgorithmCoverageDetailResponse>(algorithmId);
    },
    enabled: Boolean(algorithmId),
  });

  const runsQuery = useQuery<{ runs: GameOutcomeRun[] } | null>({
    queryKey: ['analytics', 'game-outcome-runs', selectedTournamentId],
    queryFn: async () => {
      if (!selectedTournamentId) return null;
      return analyticsService.listGameOutcomeRunsForTournament<{ runs: GameOutcomeRun[] }>(selectedTournamentId);
    },
    enabled: Boolean(selectedTournamentId),
  });

  const runs = useMemo(() => {
    const all = runsQuery.data?.runs ?? [];
    if (!algorithmId) return [];
    return all.filter((r) => r.algorithm_id === algorithmId);
  }, [runsQuery.data?.runs, algorithmId]);

  useEffect(() => {
    // When switching tournaments (or after loading runs), default to the latest run for this algorithm.
    if (!selectedTournamentId) {
      if (selectedGameOutcomeRunId) setSelectedGameOutcomeRunId('');
      return;
    }
    if (selectedGameOutcomeRunId) return;
    if (runs.length > 0) {
      setSelectedGameOutcomeRunId(runs[0].id);
    }
  }, [runs, selectedTournamentId, selectedGameOutcomeRunId]);

  useEffect(() => {
    const next = new URLSearchParams(searchParams);
    if (selectedTournamentId) {
      next.set('tournamentId', selectedTournamentId);
    } else {
      next.delete('tournamentId');
    }
    if (selectedGameOutcomeRunId) {
      next.set('gameOutcomeRunId', selectedGameOutcomeRunId);
    } else {
      next.delete('gameOutcomeRunId');
    }

    if (next.toString() !== searchParams.toString()) {
      setSearchParams(next, { replace: true });
    }
  }, [searchParams, selectedTournamentId, selectedGameOutcomeRunId, setSearchParams]);

  const predictedAdvancementQuery = useQuery<{ teams: TeamPredictedAdvancement[] } | null>({
    queryKey: ['analytics', 'predicted-advancement', selectedTournamentId, selectedGameOutcomeRunId],
    queryFn: async () => {
      if (!selectedTournamentId) return null;
      return analyticsService.getTournamentPredictedAdvancement<{ teams: TeamPredictedAdvancement[] }>({
        tournamentId: selectedTournamentId,
        gameOutcomeRunId: selectedGameOutcomeRunId || undefined,
      });
    },
    enabled: Boolean(selectedTournamentId),
  });

  const tournaments = coverageDetailQuery.data?.items ?? [];
  const sortedTournaments = useMemo(() => {
    return tournaments
      .slice()
      .sort((a, b) => {
        const aHas = Boolean(a.last_run_at);
        const bHas = Boolean(b.last_run_at);
        if (aHas !== bHas) return aHas ? -1 : 1;
        const aLast = a.last_run_at ?? '';
        const bLast = b.last_run_at ?? '';
        if (bLast !== aLast) return bLast.localeCompare(aLast);
        const aStart = a.starting_at ?? '';
        const bStart = b.starting_at ?? '';
        if (bStart !== aStart) return bStart.localeCompare(aStart);
        return a.tournament_name.localeCompare(b.tournament_name);
      });
  }, [tournaments]);
  const selectedTournament = useMemo(() => {
    return tournaments.find((t) => t.tournament_id === selectedTournamentId) ?? null;
  }, [tournaments, selectedTournamentId]);

  const formatProb = (p: number) => `${(p * 100).toFixed(1)}%`;
  const tournamentYear = (startingAt?: string | null) => {
    if (!startingAt) return '—';
    const d = new Date(startingAt);
    if (Number.isNaN(d.getTime())) return '—';
    return String(d.getFullYear());
  };
  const tournamentYearOrName = (t: AlgorithmCoverageTournament) => {
    const byStart = tournamentYear(t.starting_at);
    if (byStart !== '—') return byStart;
    const m = t.tournament_name.match(/\b(19\d{2}|20\d{2})\b/);
    return m ? m[1] : '—';
  };
  const fmtDateTime = (iso?: string | null) => {
    if (!iso) return 'Never';
    const d = new Date(iso);
    if (Number.isNaN(d.getTime())) return '—';
    return d.toLocaleString();
  };

  return (
    <PageContainer className="max-w-none">
      <PageHeader
        title="Advancements Algorithm"
        subtitle={algorithm ? algorithm.name : algorithmId}
        actions={
          <Link to="/lab" className="text-blue-600 hover:text-blue-800">
            ← Back to Lab
          </Link>
        }
      />

      <div className="space-y-6">
        <Card>
          <h2 className="text-xl font-semibold mb-2">Algorithm</h2>
          {algorithmsQuery.isLoading ? <div className="text-gray-500">Loading algorithm...</div> : null}
          {algorithm ? (
            <div className="text-sm text-gray-700">
              <div className="font-medium text-gray-900">{algorithm.name}</div>
              {algorithm.description ? <div className="text-gray-600">{algorithm.description}</div> : null}
              <div className="text-gray-500 mt-1">id={algorithm.id}</div>
            </div>
          ) : !algorithmsQuery.isLoading ? (
            <div className="text-sm text-gray-700">Algorithm not found in registry.</div>
          ) : null}

          {coverageDetailQuery.data ? (
            <div className="text-sm text-gray-700 mt-3">
              <div className="text-gray-600">coverage={coverageDetailQuery.data.covered}/{coverageDetailQuery.data.total}</div>
            </div>
          ) : null}
        </Card>

        <Card>
          <h2 className="text-xl font-semibold mb-4">Tournaments</h2>

          {coverageDetailQuery.isLoading ? <div className="text-gray-500">Loading coverage...</div> : null}

          {!coverageDetailQuery.isLoading && tournaments.length === 0 ? (
            <div className="text-gray-700">No tournaments found.</div>
          ) : null}

          {!coverageDetailQuery.isLoading && tournaments.length > 0 ? (
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Year</th>
                    <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Tournament</th>
                    <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Last Run</th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {sortedTournaments.map((t) => {
                    const isSelected = t.tournament_id === selectedTournamentId;
                    return (
                      <tr
                        key={t.tournament_id}
                        className={`hover:bg-gray-50 cursor-pointer ${isSelected ? 'bg-blue-50' : ''}`}
                        onClick={() => {
                          setSelectedTournamentId(t.tournament_id);
                          setSelectedGameOutcomeRunId('');
                        }}
                      >
                        <td className="px-3 py-2 text-sm text-gray-700">{tournamentYearOrName(t)}</td>
                        <td className="px-3 py-2 text-sm text-gray-900">{t.tournament_name}</td>
                        <td className="px-3 py-2 text-sm text-gray-700">{fmtDateTime(t.last_run_at)}</td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          ) : null}
        </Card>

        {selectedTournamentId ? (
          <Card>
            <h2 className="text-xl font-semibold mb-4">{selectedTournament ? selectedTournament.tournament_name : 'Tournament'}</h2>

            <div className="flex items-center gap-4">
              <label htmlFor="game-outcome-run-select" className="text-lg font-semibold whitespace-nowrap">
                Run:
              </label>

              {runsQuery.isLoading ? (
                <div className="text-gray-500">Loading runs...</div>
              ) : runs.length ? (
                <Select
                  id="game-outcome-run-select"
                  value={selectedGameOutcomeRunId}
                  onChange={(e) => setSelectedGameOutcomeRunId(e.target.value)}
                  className="flex-1 max-w-xl"
                >
                  {runs
                    .slice()
                    .sort((a, b) => b.created_at.localeCompare(a.created_at))
                    .map((run) => (
                      <option key={run.id} value={run.id}>
                        {new Date(run.created_at).toLocaleString()} ({run.id.slice(0, 8)})
                      </option>
                    ))}
                </Select>
              ) : (
                <div className="text-gray-700">No runs found for this algorithm + tournament.</div>
              )}
            </div>

            <div className="mt-4 text-sm text-gray-600">Per-team cumulative probability of reaching each round (no points).</div>

            <div className="mt-4">
              {predictedAdvancementQuery.isLoading ? <div className="text-gray-500">Loading predicted advancement...</div> : null}

              {!predictedAdvancementQuery.isLoading && predictedAdvancementQuery.data?.teams ? (
                <div className="overflow-x-auto">
                  <table className="min-w-full divide-y divide-gray-200">
                    <thead className="bg-gray-50">
                      <tr>
                        <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider sticky left-0 bg-gray-50">
                          Team
                        </th>
                        <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Seed</th>
                        <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Region</th>
                        <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">PI</th>
                        <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">R64</th>
                        <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">R32</th>
                        <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">S16</th>
                        <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">E8</th>
                        <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">FF</th>
                        <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Champ</th>
                        <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider bg-blue-50">
                          Win
                        </th>
                      </tr>
                    </thead>
                    <tbody className="bg-white divide-y divide-gray-200">
                      {predictedAdvancementQuery.data.teams.map((team) => (
                        <tr key={team.team_id} className="hover:bg-gray-50">
                          <td className="px-4 py-3 text-sm font-medium text-gray-900 sticky left-0 bg-white">{team.school_name}</td>
                          <td className="px-4 py-3 text-sm text-center text-gray-700">{team.seed}</td>
                          <td className="px-4 py-3 text-sm text-center text-gray-600">{team.region}</td>
                          <td className="px-4 py-3 text-sm text-center text-gray-700">{formatProb(team.prob_pi)}</td>
                          <td className="px-4 py-3 text-sm text-center text-gray-700">{formatProb(team.reach_r64)}</td>
                          <td className="px-4 py-3 text-sm text-center text-gray-700">{formatProb(team.reach_r32)}</td>
                          <td className="px-4 py-3 text-sm text-center text-gray-700">{formatProb(team.reach_s16)}</td>
                          <td className="px-4 py-3 text-sm text-center text-gray-700">{formatProb(team.reach_e8)}</td>
                          <td className="px-4 py-3 text-sm text-center text-gray-700">{formatProb(team.reach_ff)}</td>
                          <td className="px-4 py-3 text-sm text-center text-gray-700">{formatProb(team.reach_champ)}</td>
                          <td className="px-4 py-3 text-sm text-center font-semibold text-blue-700 bg-blue-50">{formatProb(team.win_champ)}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              ) : !predictedAdvancementQuery.isLoading ? (
                <div className="text-gray-500">No predicted advancement data available.</div>
              ) : null}
            </div>
          </Card>
        ) : null}
      </div>
    </PageContainer>
  );
}

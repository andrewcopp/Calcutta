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

export function LabAdvancementTournamentDetailPage() {
  const { algorithmId, tournamentId } = useParams<{ algorithmId: string; tournamentId: string }>();
  const [searchParams, setSearchParams] = useSearchParams();
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

  const tournament = useMemo(() => {
    const items = coverageDetailQuery.data?.items ?? [];
    return items.find((t) => t.tournament_id === tournamentId) ?? null;
  }, [coverageDetailQuery.data?.items, tournamentId]);

  const runsQuery = useQuery<{ runs: GameOutcomeRun[] } | null>({
    queryKey: ['analytics', 'game-outcome-runs', tournamentId],
    queryFn: async () => {
      if (!tournamentId) return null;
      return analyticsService.listGameOutcomeRunsForTournament<{ runs: GameOutcomeRun[] }>(tournamentId);
    },
    enabled: Boolean(tournamentId),
  });

  const runs = useMemo(() => {
    const all = runsQuery.data?.runs ?? [];
    if (!algorithmId) return [];
    return all.filter((r) => r.algorithm_id === algorithmId).slice().sort((a, b) => b.created_at.localeCompare(a.created_at));
  }, [runsQuery.data?.runs, algorithmId]);

  useEffect(() => {
    if (selectedGameOutcomeRunId) return;
    if (runs.length > 0) {
      setSelectedGameOutcomeRunId(runs[0].id);
    }
  }, [runs, selectedGameOutcomeRunId]);

  useEffect(() => {
    const next = new URLSearchParams(searchParams);
    if (selectedGameOutcomeRunId) {
      next.set('gameOutcomeRunId', selectedGameOutcomeRunId);
    } else {
      next.delete('gameOutcomeRunId');
    }
    if (next.toString() !== searchParams.toString()) {
      setSearchParams(next, { replace: true });
    }
  }, [searchParams, selectedGameOutcomeRunId, setSearchParams]);

  const predictedAdvancementQuery = useQuery<{ teams: TeamPredictedAdvancement[] } | null>({
    queryKey: ['analytics', 'predicted-advancement', tournamentId, selectedGameOutcomeRunId],
    queryFn: async () => {
      if (!tournamentId) return null;
      return analyticsService.getTournamentPredictedAdvancement<{ teams: TeamPredictedAdvancement[] }>({
        tournamentId,
        gameOutcomeRunId: selectedGameOutcomeRunId || undefined,
      });
    },
    enabled: Boolean(tournamentId),
  });

  const formatProb = (p: number) => `${(p * 100).toFixed(1)}%`;

  return (
    <PageContainer className="max-w-none">
      <PageHeader
        title="Advancements"
        subtitle={tournament ? tournament.tournament_name : tournamentId}
        leftActions={
          <Link to={`/lab/advancements/algorithms/${encodeURIComponent(algorithmId || '')}`} className="text-blue-600 hover:text-blue-800">
            ← Back to Algorithm
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
        </Card>

        <Card>
          <h2 className="text-xl font-semibold mb-4">Run</h2>

          <div className="flex items-center gap-4">
            <label htmlFor="game-outcome-run-select" className="text-lg font-semibold whitespace-nowrap">
              Game Outcomes Run:
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
                {runs.map((run) => (
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
        </Card>

        <Card>
          <h2 className="text-xl font-semibold mb-4">Advancement</h2>

          {predictedAdvancementQuery.isLoading ? <div className="text-gray-500">Loading predicted advancement...</div> : null}

          {!predictedAdvancementQuery.isLoading && predictedAdvancementQuery.data?.teams ? (
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider sticky left-0 bg-gray-50">Team</th>
                    <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Seed</th>
                    <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Region</th>
                    <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">PI</th>
                    <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">R64</th>
                    <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">R32</th>
                    <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">S16</th>
                    <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">E8</th>
                    <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">FF</th>
                    <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Champ</th>
                    <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider bg-blue-50">Win</th>
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
        </Card>
      </div>
    </PageContainer>
  );
}

import React, { useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link, useParams } from 'react-router-dom';

import { Card } from '../components/ui/Card';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { Select } from '../components/ui/Select';
import { analyticsService } from '../services/analyticsService';
import { tournamentService } from '../services/tournamentService';
import type { Tournament } from '../types/calcutta';

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

export function LabAdvancementAlgorithmDetailPage() {
  const { algorithmId } = useParams<{ algorithmId: string }>();
  const [selectedTournamentId, setSelectedTournamentId] = useState<string>('');

  const { data: tournaments = [], isLoading: tournamentsLoading } = useQuery<Tournament[]>({
    queryKey: ['tournaments', 'all'],
    queryFn: tournamentService.getAllTournaments,
  });

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
        </Card>

        <Card>
          <h2 className="text-xl font-semibold mb-4">Tournaments with runs</h2>

          <div className="flex items-center gap-4">
            <label htmlFor="tournament-select" className="text-lg font-semibold whitespace-nowrap">
              Select Tournament:
            </label>

            {tournamentsLoading ? (
              <div className="text-gray-500">Loading tournaments...</div>
            ) : tournaments.length === 0 ? (
              <div className="text-gray-500">No tournaments found</div>
            ) : (
              <Select
                id="tournament-select"
                value={selectedTournamentId}
                onChange={(e) => setSelectedTournamentId(e.target.value)}
                className="flex-1 max-w-md"
              >
                <option value="">-- Select a tournament --</option>
                {tournaments.map((tournament) => (
                  <option key={tournament.id} value={tournament.id}>
                    {tournament.name}
                  </option>
                ))}
              </Select>
            )}
          </div>

          {selectedTournamentId ? (
            <div className="mt-4">
              {runsQuery.isLoading ? <div className="text-gray-500">Loading runs...</div> : null}

              {!runsQuery.isLoading && runs.length === 0 ? (
                <div className="text-gray-700">No runs found for this algorithm + tournament.</div>
              ) : null}

              {!runsQuery.isLoading && runs.length > 0 ? (
                <div className="overflow-x-auto">
                  <table className="min-w-full divide-y divide-gray-200">
                    <thead className="bg-gray-50">
                      <tr>
                        <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Created</th>
                        <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Run ID</th>
                        <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
                      </tr>
                    </thead>
                    <tbody className="bg-white divide-y divide-gray-200">
                      {runs
                        .slice()
                        .sort((a, b) => b.created_at.localeCompare(a.created_at))
                        .map((run) => {
                          const labUrl = `/lab?tab=advancements&algorithmId=${encodeURIComponent(algorithmId || '')}&tournamentId=${encodeURIComponent(
                            selectedTournamentId
                          )}&gameOutcomeRunId=${encodeURIComponent(run.id)}`;

                          return (
                            <tr key={run.id} className="hover:bg-gray-50">
                              <td className="px-3 py-2 text-sm text-gray-700">{new Date(run.created_at).toLocaleString()}</td>
                              <td className="px-3 py-2 text-sm text-gray-900 font-mono">{run.id}</td>
                              <td className="px-3 py-2 text-sm">
                                <Link to={labUrl} className="text-blue-600 hover:text-blue-800">
                                  Open in Lab
                                </Link>
                              </td>
                            </tr>
                          );
                        })}
                    </tbody>
                  </table>
                </div>
              ) : null}
            </div>
          ) : null}
        </Card>
      </div>
    </PageContainer>
  );
}

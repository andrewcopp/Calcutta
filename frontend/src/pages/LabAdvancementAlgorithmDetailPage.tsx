import React, { useEffect, useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link, useNavigate, useParams } from 'react-router-dom';

import { Alert } from '../components/ui/Alert';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { PageContainer, PageHeader } from '../components/ui/Page';
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
  const navigate = useNavigate();

	const [runAllError, setRunAllError] = useState<string | null>(null);
	const [runAllLoading, setRunAllLoading] = useState(false);

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

	const runAll = async () => {
		if (!algorithmId) return;
		setRunAllError(null);
		setRunAllLoading(true);
		try {
			await analyticsService.bulkCreateGameOutcomeRunsForAlgorithm(algorithmId);
			await coverageDetailQuery.refetch();
		} catch (e) {
			const msg = e instanceof Error ? e.message : 'Failed to enqueue runs';
			setRunAllError(msg);
		} finally {
			setRunAllLoading(false);
		}
	};

  return (
    <PageContainer className="max-w-none">
      <PageHeader
        title="Advancements Algorithm"
        subtitle={algorithm ? algorithm.name : algorithmId}
        leftActions={
          <Link to="/lab" className="text-blue-600 hover:text-blue-800">
            ← Back to Lab
          </Link>
        }
		actions={
			<Button onClick={runAll} loading={runAllLoading} disabled={!algorithmId || algorithmsQuery.isLoading}>
				Run for all tournaments
			</Button>
		}
      />

      <div className="space-y-6">
			{runAllError ? <Alert variant="error">{runAllError}</Alert> : null}

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
                    return (
                      <tr
                        key={t.tournament_id}
                        className="hover:bg-gray-50 cursor-pointer"
                        onClick={() => {
                          navigate(
                            `/lab/advancements/algorithms/${encodeURIComponent(algorithmId || '')}/tournaments/${encodeURIComponent(
                              t.tournament_id
                            )}`
                          );
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
      </div>
    </PageContainer>
  );
}

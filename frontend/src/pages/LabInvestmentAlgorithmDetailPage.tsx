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

type MarketShareRun = {
  id: string;
  algorithm_id: string;
  calcutta_id: string;
  params_json?: unknown;
  git_sha?: string | null;
  created_at: string;
};

type AlgorithmCoverageCalcutta = {
  calcutta_id: string;
  calcutta_name: string;
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
  items: AlgorithmCoverageCalcutta[];
};

type TeamPredictedMarketShare = {
  team_id: string;
  school_name: string;
  seed: number;
  region: string;
  rational_share: number;
  predicted_share: number;
  delta_percent: number;
};

export function LabInvestmentAlgorithmDetailPage() {
  const { algorithmId } = useParams<{ algorithmId: string }>();
  const [searchParams, setSearchParams] = useSearchParams();
  const [selectedCalcuttaId, setSelectedCalcuttaId] = useState<string>(() => searchParams.get('calcuttaId') || '');
  const [selectedMarketShareRunId, setSelectedMarketShareRunId] = useState<string>(() => searchParams.get('marketShareRunId') || '');
  const [selectedGameOutcomeRunId, setSelectedGameOutcomeRunId] = useState<string>(() => searchParams.get('gameOutcomeRunId') || '');

  const algorithmsQuery = useQuery<{ items: Algorithm[] } | null>({
    queryKey: ['analytics', 'algorithms', 'market_share'],
    queryFn: async () => {
      const filtered = await analyticsService.listAlgorithms<{ items: Algorithm[] }>('market_share');
      if (filtered?.items?.length) return filtered;
      return analyticsService.listAlgorithms<{ items: Algorithm[] }>();
    },
  });

  const algorithm = useMemo(() => {
    const items = algorithmsQuery.data?.items ?? [];
    return items.find((a) => a.id === algorithmId) ?? null;
  }, [algorithmsQuery.data?.items, algorithmId]);

  const coverageDetailQuery = useQuery<AlgorithmCoverageDetailResponse | null>({
    queryKey: ['analytics', 'coverage-detail', 'market-share', algorithmId],
    queryFn: async () => {
      if (!algorithmId) return null;
      return analyticsService.getMarketShareAlgorithmCoverageDetail<AlgorithmCoverageDetailResponse>(algorithmId);
    },
    enabled: Boolean(algorithmId),
  });

  const marketShareRunsQuery = useQuery<{ runs: MarketShareRun[] } | null>({
    queryKey: ['analytics', 'market-share-runs', selectedCalcuttaId],
    queryFn: async () => {
      if (!selectedCalcuttaId) return null;
      return analyticsService.listMarketShareRunsForCalcutta<{ runs: MarketShareRun[] }>(selectedCalcuttaId);
    },
    enabled: Boolean(selectedCalcuttaId),
  });

  const marketShareRuns = useMemo(() => {
    const all = marketShareRunsQuery.data?.runs ?? [];
    if (!algorithmId) return [];
    return all.filter((r) => r.algorithm_id === algorithmId);
  }, [marketShareRunsQuery.data?.runs, algorithmId]);

  const selectedCalcutta = useMemo(() => {
    const items = coverageDetailQuery.data?.items ?? [];
    return items.find((c) => c.calcutta_id === selectedCalcuttaId) ?? null;
  }, [coverageDetailQuery.data?.items, selectedCalcuttaId]);

  const gameOutcomeRunsQuery = useQuery<{ runs: GameOutcomeRun[] } | null>({
    queryKey: ['analytics', 'game-outcome-runs', selectedCalcutta?.tournament_id],
    queryFn: async () => {
      if (!selectedCalcutta?.tournament_id) return null;
      return analyticsService.listGameOutcomeRunsForTournament<{ runs: GameOutcomeRun[] }>(selectedCalcutta.tournament_id);
    },
    enabled: Boolean(selectedCalcutta?.tournament_id),
  });

  useEffect(() => {
    if (!selectedCalcuttaId) {
      if (selectedMarketShareRunId) setSelectedMarketShareRunId('');
      if (selectedGameOutcomeRunId) setSelectedGameOutcomeRunId('');
      return;
    }

    if (!selectedMarketShareRunId && marketShareRuns.length > 0) {
      setSelectedMarketShareRunId(marketShareRuns[0].id);
    }

    const goRuns = gameOutcomeRunsQuery.data?.runs ?? [];
    if (!selectedGameOutcomeRunId && goRuns.length > 0) {
      setSelectedGameOutcomeRunId(goRuns[0].id);
    }
  }, [
    selectedCalcuttaId,
    selectedMarketShareRunId,
    selectedGameOutcomeRunId,
    marketShareRuns,
    gameOutcomeRunsQuery.data?.runs,
  ]);

  useEffect(() => {
    const next = new URLSearchParams(searchParams);
    if (selectedCalcuttaId) {
      next.set('calcuttaId', selectedCalcuttaId);
    } else {
      next.delete('calcuttaId');
    }
    if (selectedMarketShareRunId) {
      next.set('marketShareRunId', selectedMarketShareRunId);
    } else {
      next.delete('marketShareRunId');
    }
    if (selectedGameOutcomeRunId) {
      next.set('gameOutcomeRunId', selectedGameOutcomeRunId);
    } else {
      next.delete('gameOutcomeRunId');
    }

    if (next.toString() !== searchParams.toString()) {
      setSearchParams(next, { replace: true });
    }
  }, [searchParams, selectedCalcuttaId, selectedMarketShareRunId, selectedGameOutcomeRunId, setSearchParams]);

  const predictedMarketShareQuery = useQuery<{ teams: TeamPredictedMarketShare[] } | null>({
    queryKey: [
      'analytics',
      'predicted-market-share',
      selectedCalcuttaId,
      selectedMarketShareRunId,
      selectedGameOutcomeRunId,
    ],
    queryFn: async () => {
      if (!selectedCalcuttaId) return null;
      return analyticsService.getCalcuttaPredictedMarketShare<{ teams: TeamPredictedMarketShare[] }>({
        calcuttaId: selectedCalcuttaId,
        marketShareRunId: selectedMarketShareRunId || undefined,
        gameOutcomeRunId: selectedGameOutcomeRunId || undefined,
      });
    },
    enabled: Boolean(selectedCalcuttaId),
  });

  const calcuttas = coverageDetailQuery.data?.items ?? [];
  const sortedCalcuttas = useMemo(() => {
    return calcuttas
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
        return a.calcutta_name.localeCompare(b.calcutta_name);
      });
  }, [calcuttas]);
  const fmtDateTime = (iso?: string | null) => {
    if (!iso) return 'Never';
    const d = new Date(iso);
    if (Number.isNaN(d.getTime())) return '—';
    return d.toLocaleString();
  };
  const calcuttaYear = (startingAt?: string | null) => {
    if (!startingAt) return '—';
    const d = new Date(startingAt);
    if (Number.isNaN(d.getTime())) return '—';
    return String(d.getFullYear());
  };
  const calcuttaYearOrName = (c: AlgorithmCoverageCalcutta) => {
    const byStart = calcuttaYear(c.starting_at);
    if (byStart !== '—') return byStart;
    const m = c.tournament_name.match(/\b(19\d{2}|20\d{2})\b/);
    return m ? m[1] : '—';
  };

  const formatShare = (p: number) => `${(p * 100).toFixed(3)}%`;
  const formatDeltaPercent = (p: number) => {
    if (!Number.isFinite(p)) return '—';
    const formatted = p.toFixed(1);
    return p > 0 ? `+${formatted}%` : `${formatted}%`;
  };

  const deltaClass = (p: number) => {
    if (!Number.isFinite(p)) return 'text-gray-700';
    if (p < 0) return 'text-green-700 font-semibold';
    if (p > 0) return 'text-red-700 font-semibold';
    return 'text-gray-700';
  };

  return (
    <PageContainer className="max-w-none">
      <PageHeader
        title="Investments Algorithm"
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
          <h2 className="text-xl font-semibold mb-4">Calcuttas</h2>

          {coverageDetailQuery.isLoading ? <div className="text-gray-500">Loading coverage...</div> : null}

          {!coverageDetailQuery.isLoading && calcuttas.length === 0 ? (
            <div className="text-gray-700">No calcuttas found.</div>
          ) : null}

          {!coverageDetailQuery.isLoading && calcuttas.length > 0 ? (
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Year</th>
                    <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Calcutta</th>
                    <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Tournament</th>
                    <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Last Run</th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {sortedCalcuttas.map((c) => {
                    const isSelected = c.calcutta_id === selectedCalcuttaId;
                    return (
                      <tr
                        key={c.calcutta_id}
                        className={`hover:bg-gray-50 cursor-pointer ${isSelected ? 'bg-blue-50' : ''}`}
                        onClick={() => {
                          setSelectedCalcuttaId(c.calcutta_id);
                          setSelectedMarketShareRunId('');
                          setSelectedGameOutcomeRunId('');
                        }}
                      >
                        <td className="px-3 py-2 text-sm text-gray-700">{calcuttaYearOrName(c)}</td>
                        <td className="px-3 py-2 text-sm text-gray-900">{c.calcutta_name}</td>
                        <td className="px-3 py-2 text-sm text-gray-700">{c.tournament_name}</td>
                        <td className="px-3 py-2 text-sm text-gray-700">{fmtDateTime(c.last_run_at)}</td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          ) : null}
        </Card>

        {selectedCalcuttaId ? (
          <Card>
            <h2 className="text-xl font-semibold mb-4">{selectedCalcutta ? selectedCalcutta.calcutta_name : 'Calcutta'}</h2>

            <div className="flex items-center gap-4">
              <label htmlFor="market-share-run-select" className="text-lg font-semibold whitespace-nowrap">
                Market Share Run:
              </label>

              {marketShareRunsQuery.isLoading ? (
                <div className="text-gray-500">Loading runs...</div>
              ) : marketShareRuns.length ? (
                <Select
                  id="market-share-run-select"
                  value={selectedMarketShareRunId}
                  onChange={(e) => setSelectedMarketShareRunId(e.target.value)}
                  className="flex-1 max-w-xl"
                >
                  {marketShareRuns
                    .slice()
                    .sort((a, b) => b.created_at.localeCompare(a.created_at))
                    .map((run) => (
                      <option key={run.id} value={run.id}>
                        {new Date(run.created_at).toLocaleString()} ({run.id.slice(0, 8)})
                      </option>
                    ))}
                </Select>
              ) : (
                <div className="text-gray-700">No market share runs found for this algorithm + calcutta.</div>
              )}
            </div>

            <div className="mt-4 flex items-center gap-4">
              <label htmlFor="game-outcome-run-select" className="text-lg font-semibold whitespace-nowrap">
                Game Outcomes Run:
              </label>

              {gameOutcomeRunsQuery.isLoading ? (
                <div className="text-gray-500">Loading runs...</div>
              ) : gameOutcomeRunsQuery.data?.runs?.length ? (
                <Select
                  id="game-outcome-run-select"
                  value={selectedGameOutcomeRunId}
                  onChange={(e) => setSelectedGameOutcomeRunId(e.target.value)}
                  className="flex-1 max-w-xl"
                >
                  {gameOutcomeRunsQuery.data.runs
                    .slice()
                    .sort((a, b) => b.created_at.localeCompare(a.created_at))
                    .map((run) => (
                      <option key={run.id} value={run.id}>
                        {new Date(run.created_at).toLocaleString()} ({run.id.slice(0, 8)})
                      </option>
                    ))}
                </Select>
              ) : (
                <div className="text-gray-700">No game outcome runs found for this tournament.</div>
              )}
            </div>

            <div className="mt-4 text-sm text-gray-600">Per-team predicted and rational market share (with delta percent).</div>

            <div className="mt-4">
              {predictedMarketShareQuery.isLoading ? <div className="text-gray-500">Loading predicted market share...</div> : null}

              {!predictedMarketShareQuery.isLoading && predictedMarketShareQuery.data?.teams ? (
                <div className="overflow-x-auto">
                  <table className="min-w-full divide-y divide-gray-200">
                    <thead className="bg-gray-50">
                      <tr>
                        <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider sticky left-0 bg-gray-50">
                          Team
                        </th>
                        <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Seed</th>
                        <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Region</th>
                        <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Rational</th>
                        <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider bg-green-50">
                          Predicted
                        </th>
                        <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Delta</th>
                      </tr>
                    </thead>
                    <tbody className="bg-white divide-y divide-gray-200">
                      {predictedMarketShareQuery.data.teams.map((team) => (
                        <tr key={team.team_id} className="hover:bg-gray-50">
                          <td className="px-4 py-3 text-sm font-medium text-gray-900 sticky left-0 bg-white">{team.school_name}</td>
                          <td className="px-4 py-3 text-sm text-center text-gray-700">{team.seed}</td>
                          <td className="px-4 py-3 text-sm text-center text-gray-600">{team.region}</td>
                          <td className="px-4 py-3 text-sm text-center text-gray-700">{formatShare(team.rational_share)}</td>
                          <td className="px-4 py-3 text-sm text-center font-semibold text-green-700 bg-green-50">
                            {formatShare(team.predicted_share)}
                          </td>
                          <td className={`px-4 py-3 text-sm text-center ${deltaClass(team.delta_percent)}`}>{formatDeltaPercent(team.delta_percent)}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              ) : !predictedMarketShareQuery.isLoading ? (
                <div className="text-gray-500">No predicted market share data available.</div>
              ) : null}
            </div>
          </Card>
        ) : null}
      </div>
    </PageContainer>
  );
}

import React, { useEffect, useMemo, useState } from 'react';
import { Navigate, useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { mlAnalyticsService } from '../services/mlAnalyticsService';
import { RunViewerHeader } from '../components/RunViewerHeader';

const formatPercent = (value: number) => `${(value * 100).toFixed(1)}%`;

export function RunReturnsPage() {
  const { year, runId } = useParams<{ year: string; runId: string }>();
  const parsedYear = year ? Number(year) : undefined;

  if (!parsedYear || Number.isNaN(parsedYear) || !runId) {
    return <Navigate to="/runs" replace />;
  }

  const decodedRunId = useMemo(() => decodeURIComponent(runId), [runId]);

  const ourEntryQuery = useQuery({
    queryKey: ['mlAnalytics', 'ourEntryDetails', parsedYear, decodedRunId],
    queryFn: () => mlAnalyticsService.getOurEntryDetails(parsedYear, decodedRunId),
  });

  const calcuttaId = ourEntryQuery.data?.run.calcutta_id ?? null;

  const strategyRunsQuery = useQuery({
    queryKey: ['analytics', 'strategyGenerationRuns', calcuttaId],
    queryFn: () => mlAnalyticsService.listStrategyGenerationRuns(calcuttaId as string),
    enabled: Boolean(calcuttaId),
  });

  const defaultStrategyGenerationRunId = useMemo(() => {
    const runs = strategyRunsQuery.data?.runs ?? [];
    const match = runs.find((r) => r.run_key === decodedRunId);
    return match?.id ?? null;
  }, [strategyRunsQuery.data, decodedRunId]);

  const [selectedStrategyGenerationRunId, setSelectedStrategyGenerationRunId] = useState<string | null>(null);

  useEffect(() => {
    if (selectedStrategyGenerationRunId !== null) return;
    if (!strategyRunsQuery.data) return;
    if (defaultStrategyGenerationRunId) {
      setSelectedStrategyGenerationRunId(defaultStrategyGenerationRunId);
    }
  }, [defaultStrategyGenerationRunId, selectedStrategyGenerationRunId, strategyRunsQuery.data]);

  const [sortKey, setSortKey] = useState<'expected_value' | 'prob_champ' | 'seed' | 'school_name'>('expected_value');
  const [sortDir, setSortDir] = useState<'desc' | 'asc'>('desc');

  const returnsQuery = useQuery({
    queryKey: ['analytics', 'predictedReturns', calcuttaId, selectedStrategyGenerationRunId],
    queryFn: () =>
      mlAnalyticsService.getCalcuttaPredictedReturns({
        calcuttaId: calcuttaId as string,
        strategyGenerationRunId: selectedStrategyGenerationRunId ?? undefined,
      }),
    enabled: Boolean(calcuttaId),
  });

  const sortedTeams = useMemo(() => {
    const teams = returnsQuery.data?.teams ?? [];
    const mult = sortDir === 'asc' ? 1 : -1;

    return [...teams].sort((a, b) => {
      if (sortKey === 'school_name') {
        return mult * a.school_name.localeCompare(b.school_name);
      }
      if (sortKey === 'seed') {
        return mult * (a.seed - b.seed);
      }
      if (sortKey === 'prob_champ') {
        return mult * (a.prob_champ - b.prob_champ);
      }
      return mult * (a.expected_value - b.expected_value);
    });
  }, [returnsQuery.data, sortDir, sortKey]);

  return (
    <div className="container mx-auto px-4 py-8">
      <RunViewerHeader year={parsedYear} runId={decodedRunId} runName={ourEntryQuery.data?.run.name} activeTab="returns" />

      <div className="bg-white rounded-lg shadow p-6">
        {!calcuttaId && ourEntryQuery.isSuccess && <div className="text-gray-600">No calcutta_id found for this run.</div>}
        {ourEntryQuery.isLoading && <div className="text-gray-600">Loading run context…</div>}
        {ourEntryQuery.isError && <div className="text-red-600">Failed to load run context.</div>}

        {returnsQuery.isLoading && calcuttaId && <div className="text-gray-600">Loading returns…</div>}
        {returnsQuery.isError && calcuttaId && <div className="text-red-600">Failed to load returns.</div>}

        {returnsQuery.data && (
          <div className="overflow-x-auto">
            <div className="flex flex-col md:flex-row md:items-end md:justify-between gap-3 mb-4">
              <div className="text-sm text-gray-700">
                <div>Calcutta: {calcuttaId}</div>
                <div>Strategy run id: {returnsQuery.data.strategy_generation_run_id || 'latest'}</div>
              </div>

              <div className="flex flex-col md:flex-row gap-2">
                <label className="text-sm text-gray-700">
                  Strategy run
                  <select
                    className="ml-2 border rounded px-2 py-1 text-sm"
                    value={selectedStrategyGenerationRunId ?? ''}
                    onChange={(e) => setSelectedStrategyGenerationRunId(e.target.value || null)}
                  >
                    <option value="">Latest</option>
                    {(strategyRunsQuery.data?.runs ?? []).map((r) => (
                      <option key={r.id} value={r.id}>
                        {r.run_key}
                      </option>
                    ))}
                  </select>
                </label>

                <label className="text-sm text-gray-700">
                  Sort
                  <select
                    className="ml-2 border rounded px-2 py-1 text-sm"
                    value={sortKey}
                    onChange={(e) => setSortKey(e.target.value as typeof sortKey)}
                  >
                    <option value="expected_value">EV</option>
                    <option value="prob_champ">P(Champ)</option>
                    <option value="seed">Seed</option>
                    <option value="school_name">Team</option>
                  </select>
                </label>

                <label className="text-sm text-gray-700">
                  Dir
                  <select className="ml-2 border rounded px-2 py-1 text-sm" value={sortDir} onChange={(e) => setSortDir(e.target.value as typeof sortDir)}>
                    <option value="desc">Desc</option>
                    <option value="asc">Asc</option>
                  </select>
                </label>
              </div>
            </div>

            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Team</th>
                  <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Seed</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Region</th>
                  <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">EV</th>
                  <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">P(Champ)</th>
                  <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">P(FF)</th>
                  <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">P(E8)</th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {sortedTeams.map((t) => (
                  <tr key={t.team_id} className="hover:bg-gray-50">
                    <td className="px-4 py-3 text-sm text-gray-900">{t.school_name}</td>
                    <td className="px-4 py-3 text-sm text-right text-gray-700">{t.seed}</td>
                    <td className="px-4 py-3 text-sm text-gray-700">{t.region}</td>
                    <td className="px-4 py-3 text-sm text-right text-gray-700">{t.expected_value.toFixed(2)}</td>
                    <td className="px-4 py-3 text-sm text-right text-gray-700">{formatPercent(t.prob_champ)}</td>
                    <td className="px-4 py-3 text-sm text-right text-gray-700">{formatPercent(t.prob_ff)}</td>
                    <td className="px-4 py-3 text-sm text-right text-gray-700">{formatPercent(t.prob_e8)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}

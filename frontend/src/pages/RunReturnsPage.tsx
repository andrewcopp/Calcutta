import React, { useEffect, useMemo, useState } from 'react';
import { Navigate, useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { mlAnalyticsService } from '../services/mlAnalyticsService';
import { RunViewerHeader } from '../components/RunViewerHeader';
import { Alert } from '../components/ui/Alert';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer } from '../components/ui/Page';
import { Select } from '../components/ui/Select';

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
    <PageContainer>
      <RunViewerHeader year={parsedYear} runId={decodedRunId} runName={ourEntryQuery.data?.run.name} activeTab="returns" />

      <Card>
        {!calcuttaId && ourEntryQuery.isSuccess ? (
          <Alert variant="info" className="mb-4">
            No calcutta_id found for this run.
          </Alert>
        ) : null}

        {ourEntryQuery.isLoading ? <LoadingState label="Loading run context..." layout="inline" /> : null}

        {ourEntryQuery.isError ? (
          <Alert variant="error" className="mt-3">
            <div className="font-semibold mb-1">Failed to load run context</div>
            <div className="mb-3">{ourEntryQuery.error instanceof Error ? ourEntryQuery.error.message : 'An error occurred'}</div>
            <Button size="sm" onClick={() => ourEntryQuery.refetch()}>
              Retry
            </Button>
          </Alert>
        ) : null}

        {returnsQuery.isLoading && calcuttaId ? <LoadingState label="Loading returns..." layout="inline" /> : null}

        {returnsQuery.isError && calcuttaId ? (
          <Alert variant="error" className="mt-3">
            <div className="font-semibold mb-1">Failed to load returns</div>
            <div className="mb-3">{returnsQuery.error instanceof Error ? returnsQuery.error.message : 'An error occurred'}</div>
            <Button size="sm" onClick={() => returnsQuery.refetch()}>
              Retry
            </Button>
          </Alert>
        ) : null}

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
                  <Select
                    className="ml-2"
                    value={selectedStrategyGenerationRunId ?? ''}
                    onChange={(e) => setSelectedStrategyGenerationRunId(e.target.value || null)}
                  >
                    <option value="">Latest</option>
                    {(strategyRunsQuery.data?.runs ?? []).map((r) => (
                      <option key={r.id} value={r.id}>
                        {r.run_key}
                      </option>
                    ))}
                  </Select>
                </label>

                <label className="text-sm text-gray-700">
                  Sort
                  <Select
                    className="ml-2"
                    value={sortKey}
                    onChange={(e) => setSortKey(e.target.value as typeof sortKey)}
                  >
                    <option value="expected_value">EV</option>
                    <option value="prob_champ">P(Champ)</option>
                    <option value="seed">Seed</option>
                    <option value="school_name">Team</option>
                  </Select>
                </label>

                <label className="text-sm text-gray-700">
                  Dir
                  <Select className="ml-2" value={sortDir} onChange={(e) => setSortDir(e.target.value as typeof sortDir)}>
                    <option value="desc">Desc</option>
                    <option value="asc">Asc</option>
                  </Select>
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

        {!returnsQuery.isLoading && !returnsQuery.isError && calcuttaId && !returnsQuery.data ? (
          <Alert variant="info" className="mt-3">
            No returns data available.
          </Alert>
        ) : null}
      </Card>
    </PageContainer>
  );
}

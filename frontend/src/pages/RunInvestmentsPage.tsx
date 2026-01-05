import React, { useEffect, useMemo, useState } from 'react';
import { Navigate, useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { mlAnalyticsService } from '../services/mlAnalyticsService';
import { RunViewerHeader } from '../components/RunViewerHeader';
import { Alert } from '../components/ui/Alert';
import { Card } from '../components/ui/Card';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer } from '../components/ui/Page';
import { Select } from '../components/ui/Select';

export function RunInvestmentsPage() {
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

  const [sortKey, setSortKey] = useState<'delta' | 'predicted' | 'rational' | 'seed' | 'school_name'>('delta');
  const [sortDir, setSortDir] = useState<'desc' | 'asc'>('desc');

  const investmentsQuery = useQuery({
    queryKey: ['analytics', 'predictedInvestment', calcuttaId, selectedStrategyGenerationRunId],
    queryFn: () =>
      mlAnalyticsService.getCalcuttaPredictedInvestment({
        calcuttaId: calcuttaId as string,
        strategyGenerationRunId: selectedStrategyGenerationRunId ?? undefined,
      }),
    enabled: Boolean(calcuttaId),
  });

  const sortedTeams = useMemo(() => {
    const teams = investmentsQuery.data?.teams ?? [];
    const mult = sortDir === 'asc' ? 1 : -1;

    return [...teams].sort((a, b) => {
      if (sortKey === 'school_name') {
        return mult * a.school_name.localeCompare(b.school_name);
      }
      if (sortKey === 'seed') {
        return mult * (a.seed - b.seed);
      }
      if (sortKey === 'rational') {
        return mult * (a.rational - b.rational);
      }
      if (sortKey === 'predicted') {
        return mult * (a.predicted - b.predicted);
      }
      return mult * (a.delta - b.delta);
    });
  }, [investmentsQuery.data, sortDir, sortKey]);

  return (
    <PageContainer>
      <RunViewerHeader year={parsedYear} runId={decodedRunId} runName={ourEntryQuery.data?.run.name} activeTab="investments" />

      <Card>
        {!calcuttaId && ourEntryQuery.isSuccess && <div className="text-gray-600">No calcutta_id found for this run.</div>}
        {ourEntryQuery.isLoading && <LoadingState label="Loading run context..." layout="inline" />}
        {ourEntryQuery.isError && <Alert variant="error">Failed to load run context.</Alert>}

        {investmentsQuery.isLoading && calcuttaId && <LoadingState label="Loading investments..." layout="inline" />}
        {investmentsQuery.isError && calcuttaId && <Alert variant="error">Failed to load investments.</Alert>}

        {investmentsQuery.data && (
          <div className="overflow-x-auto">
            <div className="flex flex-col md:flex-row md:items-end md:justify-between gap-3 mb-4">
              <div className="text-sm text-gray-700">
                <div>Calcutta: {calcuttaId}</div>
                <div>Strategy run id: {investmentsQuery.data.strategy_generation_run_id || 'latest'}</div>
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
                    <option value="delta">Delta</option>
                    <option value="predicted">Predicted</option>
                    <option value="rational">Rational</option>
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
                  <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Rational</th>
                  <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Predicted</th>
                  <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Delta %</th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {sortedTeams.map((t) => (
                  <tr key={t.team_id} className="hover:bg-gray-50">
                    <td className="px-4 py-3 text-sm text-gray-900">{t.school_name}</td>
                    <td className="px-4 py-3 text-sm text-right text-gray-700">{t.seed}</td>
                    <td className="px-4 py-3 text-sm text-gray-700">{t.region}</td>
                    <td className="px-4 py-3 text-sm text-right text-gray-700">{t.rational.toFixed(1)}</td>
                    <td className="px-4 py-3 text-sm text-right text-gray-700">{t.predicted.toFixed(1)}</td>
                    <td className={`px-4 py-3 text-sm text-right ${t.delta >= 0 ? 'text-emerald-700' : 'text-red-700'}`}>
                      {t.delta.toFixed(1)}
                    </td>
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

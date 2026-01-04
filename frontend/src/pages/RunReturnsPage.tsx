import React, { useMemo } from 'react';
import { Link, Navigate, useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { mlAnalyticsService } from '../services/mlAnalyticsService';

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

  const strategyGenerationRunId = useMemo(() => {
    const runs = strategyRunsQuery.data?.runs ?? [];
    const match = runs.find((r) => r.run_key === decodedRunId);
    return match?.id ?? null;
  }, [strategyRunsQuery.data, decodedRunId]);

  const returnsQuery = useQuery({
    queryKey: ['analytics', 'predictedReturns', calcuttaId, strategyGenerationRunId],
    queryFn: () =>
      mlAnalyticsService.getCalcuttaPredictedReturns({
        calcuttaId: calcuttaId as string,
        strategyGenerationRunId: strategyGenerationRunId ?? undefined,
      }),
    enabled: Boolean(calcuttaId),
  });

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-6">
        <Link to={`/runs/${parsedYear}/${encodeURIComponent(runId)}`} className="text-blue-600 hover:text-blue-800">
          ← Back to Run
        </Link>
      </div>

      <div className="mb-8">
        <h1 className="text-3xl font-bold mb-2">Returns</h1>
        <div className="text-gray-600">
          <div>Year: {parsedYear}</div>
          <div>Run: {decodedRunId}</div>
        </div>
      </div>

      <div className="bg-white rounded-lg shadow p-6">
        {!calcuttaId && ourEntryQuery.isSuccess && <div className="text-gray-600">No calcutta_id found for this run.</div>}
        {ourEntryQuery.isLoading && <div className="text-gray-600">Loading run context…</div>}
        {ourEntryQuery.isError && <div className="text-red-600">Failed to load run context.</div>}

        {returnsQuery.isLoading && calcuttaId && <div className="text-gray-600">Loading returns…</div>}
        {returnsQuery.isError && calcuttaId && <div className="text-red-600">Failed to load returns.</div>}

        {returnsQuery.data && (
          <div className="overflow-x-auto">
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
                {returnsQuery.data.teams.map((t) => (
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

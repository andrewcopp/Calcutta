import React, { useMemo } from 'react';
import { Link, Navigate, useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { mlAnalyticsService } from '../services/mlAnalyticsService';
import { RunViewerHeader } from '../components/RunViewerHeader';

const formatPercent = (value: number) => `${(value * 100).toFixed(1)}%`;

export function RunRankingsPage() {
  const { year, runId } = useParams<{ year: string; runId: string }>();
  const parsedYear = year ? Number(year) : undefined;

  if (!parsedYear || Number.isNaN(parsedYear) || !runId) {
    return <Navigate to="/runs" replace />;
  }

  const decodedRunId = useMemo(() => decodeURIComponent(runId), [runId]);

  const rankingsQuery = useQuery({
    queryKey: ['mlAnalytics', 'entryRankings', parsedYear, decodedRunId],
    queryFn: () => mlAnalyticsService.getEntryRankings(parsedYear, decodedRunId, 200, 0),
  });

  const ourEntryQuery = useQuery({
    queryKey: ['mlAnalytics', 'ourEntryDetails', parsedYear, decodedRunId],
    queryFn: () => mlAnalyticsService.getOurEntryDetails(parsedYear, decodedRunId),
    retry: false,
  });

  const encodedRunId = useMemo(() => encodeURIComponent(decodedRunId), [decodedRunId]);

  return (
    <div className="container mx-auto px-4 py-8">
      <RunViewerHeader year={parsedYear} runId={decodedRunId} runName={ourEntryQuery.data?.run.name} activeTab="rankings" />

      <div className="mb-8">
        {ourEntryQuery.data && (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="bg-white rounded-lg shadow p-4">
              <div className="text-sm text-gray-500">Run Summary</div>
              <div className="mt-1 text-sm text-gray-700">
                <div>Strategy: {ourEntryQuery.data.run.strategy}</div>
                <div>Simulations: {ourEntryQuery.data.run.n_sims.toLocaleString()}</div>
                <div>Budget: {ourEntryQuery.data.run.budget_points} points</div>
                <div>Created: {new Date(ourEntryQuery.data.run.created_at).toLocaleString()}</div>
              </div>
            </div>

            <div className="bg-amber-50 rounded-lg shadow p-4">
              <div className="flex items-center justify-between">
                <div>
                  <div className="text-sm text-amber-700">Our Strategy</div>
                  <div className="text-xs text-amber-700/80">Aggregated performance</div>
                </div>
                <Link to={`/runs/${parsedYear}/${encodedRunId}/entries/${encodeURIComponent('our_strategy')}`} className="text-sm text-blue-700 hover:underline">
                  View portfolio
                </Link>
              </div>

              <div className="mt-2 grid grid-cols-2 gap-2 text-sm">
                <div>
                  <div className="text-gray-600">Mean normalized</div>
                  <div className="font-medium text-gray-900">{ourEntryQuery.data.summary.mean_normalized_payout.toFixed(3)}</div>
                </div>
                <div>
                  <div className="text-gray-600">Percentile</div>
                  <div className="font-medium text-gray-900">{formatPercent(ourEntryQuery.data.summary.percentile_rank)}</div>
                </div>
                <div>
                  <div className="text-gray-600">P(Top1)</div>
                  <div className="font-medium text-gray-900">{formatPercent(ourEntryQuery.data.summary.p_top1)}</div>
                </div>
                <div>
                  <div className="text-gray-600">P(In Money)</div>
                  <div className="font-medium text-gray-900">{formatPercent(ourEntryQuery.data.summary.p_in_money)}</div>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>

      <div className="bg-white rounded-lg shadow p-6">
        {rankingsQuery.isLoading && <div className="text-gray-600">Loading…</div>}
        {rankingsQuery.isError && <div className="text-red-600">Failed to load rankings.</div>}

        {rankingsQuery.data && (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Rank</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Entry</th>
                  <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Mean Normalized</th>
                  <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">P(Top1)</th>
                  <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">P(In Money)</th>
                  <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Percentile</th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {rankingsQuery.data.entries.map((e) => (
                  <tr key={`${e.entry_key}-${e.rank}`} className={e.is_our_strategy ? 'bg-amber-50 hover:bg-amber-100' : 'hover:bg-gray-50'}>
                    <td className="px-4 py-3 text-sm text-gray-700">{e.rank}</td>
                    <td className="px-4 py-3 text-sm font-medium text-blue-700">
                      <Link
                        to={`/runs/${parsedYear}/${encodedRunId}/entries/${encodeURIComponent(e.entry_key)}`}
                        className="hover:underline"
                      >
                        {e.entry_key}
                      </Link>
                    </td>
                    <td className="px-4 py-3 text-sm text-right text-gray-700">{e.mean_normalized_payout.toFixed(3)}</td>
                    <td className="px-4 py-3 text-sm text-right text-gray-700">{formatPercent(e.p_top1)}</td>
                    <td className="px-4 py-3 text-sm text-right text-gray-700">{formatPercent(e.p_in_money)}</td>
                    <td className="px-4 py-3 text-sm text-right text-gray-700">{formatPercent(e.percentile_rank)}</td>
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

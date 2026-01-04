import React, { useMemo } from 'react';
import { Link, Navigate, useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { mlAnalyticsService } from '../services/mlAnalyticsService';

const formatPercent = (value: number) => `${(value * 100).toFixed(1)}%`;

export function RunRankingsPage() {
  const { year, runId } = useParams<{ year: string; runId: string }>();
  const parsedYear = year ? Number(year) : undefined;

  if (!parsedYear || Number.isNaN(parsedYear) || !runId) {
    return <Navigate to="/runs" replace />;
  }

  const rankingsQuery = useQuery({
    queryKey: ['mlAnalytics', 'entryRankings', parsedYear, runId],
    queryFn: () => mlAnalyticsService.getEntryRankings(parsedYear, runId, 200, 0),
  });

  const title = useMemo(() => decodeURIComponent(runId), [runId]);

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-6">
        <Link to={`/runs/${parsedYear}`} className="text-blue-600 hover:text-blue-800">
          ← Back to Runs
        </Link>
      </div>

      <div className="mb-8">
        <h1 className="text-3xl font-bold mb-2">Run Rankings</h1>
        <div className="text-gray-600">
          <div>Year: {parsedYear}</div>
          <div>Run: {title}</div>
        </div>
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
                        to={`/runs/${parsedYear}/${encodeURIComponent(runId)}/entries/${encodeURIComponent(e.entry_key)}`}
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

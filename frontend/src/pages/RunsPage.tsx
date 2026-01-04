import React, { useMemo } from 'react';
import { Link, Navigate, useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { mlAnalyticsService } from '../services/mlAnalyticsService';

export function RunsPage() {
  const { year } = useParams<{ year: string }>();
  const parsedYear = year ? Number(year) : undefined;

  const fallbackYear = useMemo(() => new Date().getFullYear(), []);

  if (!parsedYear || Number.isNaN(parsedYear)) {
    return <Navigate to={`/runs/${fallbackYear}`} replace />;
  }

  const runsQuery = useQuery({
    queryKey: ['mlAnalytics', 'optimizationRuns', parsedYear],
    queryFn: () => mlAnalyticsService.getOptimizationRuns(parsedYear),
  });

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold mb-2">Runs</h1>
        <p className="text-gray-600">Read-only viewer for recent ML pipeline runs.</p>
      </div>

      <div className="bg-white rounded-lg shadow p-6">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-xl font-semibold">{parsedYear}</h2>
          <div className="flex gap-2">
            <Link
              to={`/runs/${parsedYear - 1}`}
              className="px-3 py-2 border rounded text-sm text-gray-700 hover:bg-gray-50"
            >
              Prev
            </Link>
            <Link
              to={`/runs/${parsedYear + 1}`}
              className="px-3 py-2 border rounded text-sm text-gray-700 hover:bg-gray-50"
            >
              Next
            </Link>
          </div>
        </div>

        {runsQuery.isLoading && <div className="text-gray-600">Loading…</div>}
        {runsQuery.isError && (
          <div className="text-red-600">Failed to load runs.</div>
        )}

        {runsQuery.data && (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Run</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Strategy</th>
                  <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Simulations</th>
                  <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Budget</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Created</th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {runsQuery.data.runs.map((run) => (
                  <tr key={run.run_id} className="hover:bg-gray-50">
                    <td className="px-4 py-3 text-sm font-medium text-blue-700">
                      <Link to={`/runs/${parsedYear}/${encodeURIComponent(run.run_id)}`} className="hover:underline">
                        {run.name || run.run_id}
                      </Link>
                      {run.name && (
                        <div className="text-xs text-gray-500 font-normal mt-0.5">{run.run_id}</div>
                      )}
                    </td>
                    <td className="px-4 py-3 text-sm text-gray-700">{run.strategy}</td>
                    <td className="px-4 py-3 text-sm text-right text-gray-700">{run.n_sims.toLocaleString()}</td>
                    <td className="px-4 py-3 text-sm text-right text-gray-700">{run.budget_points}</td>
                    <td className="px-4 py-3 text-sm text-gray-500">{new Date(run.created_at).toLocaleString()}</td>
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

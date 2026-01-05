import React, { useMemo } from 'react';
import { Link, Navigate, useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { mlAnalyticsService } from '../services/mlAnalyticsService';
import { Alert } from '../components/ui/Alert';
import { Card } from '../components/ui/Card';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';

export function RunsPage() {
  const { year } = useParams<{ year: string }>();
  const fallbackYear = useMemo(() => new Date().getFullYear(), []);
  const yearNumber = year ? Number(year) : NaN;
  const parsedYear = Number.isFinite(yearNumber) ? yearNumber : null;
  const hasValidYear = parsedYear !== null;

  const runsQuery = useQuery({
    queryKey: ['mlAnalytics', 'strategyRuns', parsedYear],
    queryFn: () => mlAnalyticsService.getStrategyRuns(parsedYear as number),
    enabled: hasValidYear,
  });

  if (!hasValidYear) {
    return <Navigate to={`/runs/${fallbackYear}`} replace />;
  }

  return (
    <PageContainer>
      <PageHeader title="Runs" subtitle="Read-only viewer for recent strategy runs." />

      <Card>
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

        {runsQuery.isLoading && <LoadingState label="Loading runs..." layout="inline" />}
        {runsQuery.isError && <Alert variant="error">Failed to load runs.</Alert>}

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
      </Card>
    </PageContainer>
  );
}

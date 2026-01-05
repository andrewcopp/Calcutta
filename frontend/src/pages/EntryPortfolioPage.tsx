import React, { useMemo } from 'react';
import { Link, Navigate, useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { mlAnalyticsService } from '../services/mlAnalyticsService';
import { Alert } from '../components/ui/Alert';
import { Card } from '../components/ui/Card';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';

export function EntryPortfolioPage() {
  const { year, runId, entryKey } = useParams<{ year: string; runId: string; entryKey: string }>();
  const yearNumber = year ? Number(year) : NaN;
  const parsedYear = Number.isFinite(yearNumber) ? yearNumber : null;
  const decodedRunId = useMemo(() => (runId ? decodeURIComponent(runId) : ''), [runId]);
  const decodedEntryKey = useMemo(() => (entryKey ? decodeURIComponent(entryKey) : ''), [entryKey]);
  const encodedRunId = useMemo(() => encodeURIComponent(decodedRunId), [decodedRunId]);

  const hasValidParams = parsedYear !== null && Boolean(runId) && Boolean(entryKey);

  const portfolioQuery = useQuery({
    queryKey: ['mlAnalytics', 'entryPortfolio', parsedYear, decodedRunId, decodedEntryKey],
    queryFn: () => mlAnalyticsService.getEntryPortfolio(parsedYear as number, decodedRunId, decodedEntryKey),
    enabled: hasValidParams,
  });

  if (!hasValidParams) {
    return <Navigate to="/runs" replace />;
  }

  return (
    <PageContainer>
      <PageHeader
        title="Entry Portfolio"
        subtitle={
          <div>
            <div>Year: {parsedYear}</div>
            <div>Run: {decodedRunId}</div>
            <div>Entry: {decodedEntryKey}</div>
          </div>
        }
        actions={
          <Link to={`/runs/${parsedYear}/${encodedRunId}`} className="text-blue-600 hover:text-blue-800">
            ← Back to Run Rankings
          </Link>
        }
      />

      <Card>
        {portfolioQuery.isLoading ? <LoadingState label="Loading…" /> : null}

        {portfolioQuery.isError ? <Alert variant="error">Failed to load portfolio.</Alert> : null}

        {portfolioQuery.data ? (
          <div className="overflow-x-auto">
            <div className="mb-4 text-sm text-gray-700">
              <div>Total bid: {portfolioQuery.data.total_bid} points</div>
              <div>Teams: {portfolioQuery.data.n_teams}</div>
            </div>
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Team</th>
                  <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Seed</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Region</th>
                  <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Bid</th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {portfolioQuery.data.teams.map((t) => (
                  <tr key={t.team_id} className="hover:bg-gray-50">
                    <td className="px-4 py-3 text-sm text-gray-900">{t.school_name}</td>
                    <td className="px-4 py-3 text-sm text-right text-gray-700">{t.seed}</td>
                    <td className="px-4 py-3 text-sm text-gray-700">{t.region}</td>
                    <td className="px-4 py-3 text-sm text-right text-gray-700">{t.bid_points}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : null}
      </Card>
    </PageContainer>
  );
 }

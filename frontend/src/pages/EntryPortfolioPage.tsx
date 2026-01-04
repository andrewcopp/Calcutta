import React from 'react';
import { Link, Navigate, useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { mlAnalyticsService } from '../services/mlAnalyticsService';

export function EntryPortfolioPage() {
  const { year, runId, entryKey } = useParams<{ year: string; runId: string; entryKey: string }>();
  const parsedYear = year ? Number(year) : undefined;

  if (!parsedYear || Number.isNaN(parsedYear) || !runId || !entryKey) {
    return <Navigate to="/runs" replace />;
  }

  const portfolioQuery = useQuery({
    queryKey: ['mlAnalytics', 'entryPortfolio', parsedYear, runId, entryKey],
    queryFn: () => mlAnalyticsService.getEntryPortfolio(parsedYear, runId, entryKey),
  });

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-6">
        <Link to={`/runs/${parsedYear}/${encodeURIComponent(runId)}`} className="text-blue-600 hover:text-blue-800">
          ← Back to Run Rankings
        </Link>
      </div>

      <div className="mb-8">
        <h1 className="text-3xl font-bold mb-2">Entry Portfolio</h1>
        <div className="text-gray-600">
          <div>Year: {parsedYear}</div>
          <div>Run: {decodeURIComponent(runId)}</div>
          <div>Entry: {decodeURIComponent(entryKey)}</div>
        </div>
      </div>

      <div className="bg-white rounded-lg shadow p-6">
        {portfolioQuery.isLoading && <div className="text-gray-600">Loading…</div>}
        {portfolioQuery.isError && <div className="text-red-600">Failed to load portfolio.</div>}

        {portfolioQuery.data && (
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
                    <td className="px-4 py-3 text-sm text-right text-gray-700">{t.bid_amount_points}</td>
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

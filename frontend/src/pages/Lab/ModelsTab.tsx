import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';

import { Alert } from '../../components/ui/Alert';
import { Card } from '../../components/ui/Card';
import { LoadingState } from '../../components/ui/LoadingState';
import { labService, LeaderboardResponse } from '../../services/labService';

export function ModelsTab() {
  const navigate = useNavigate();

  const leaderboardQuery = useQuery<LeaderboardResponse | null>({
    queryKey: ['lab', 'models', 'leaderboard'],
    queryFn: () => labService.getLeaderboard(),
  });

  const items = leaderboardQuery.data?.items ?? [];

  const formatPct = (val?: number | null) => {
    if (val == null) return '-';
    return `${(val * 100).toFixed(1)}%`;
  };

  const formatPayoutX = (val?: number | null) => {
    if (val == null) return '-';
    return `${val.toFixed(2)}x`;
  };

  return (
    <Card>
      <h2 className="text-xl font-semibold mb-4">Model Leaderboard</h2>
      <p className="text-sm text-gray-600 mb-4">
        Investment models ranked by average normalized payout across all evaluations.
      </p>

      {leaderboardQuery.isLoading ? <LoadingState label="Loading models..." layout="inline" /> : null}

      {leaderboardQuery.isError ? (
        <Alert variant="error">Failed to load leaderboard.</Alert>
      ) : null}

      {!leaderboardQuery.isLoading && !leaderboardQuery.isError && items.length === 0 ? (
        <Alert variant="info">No models found. Register models via Python to see them here.</Alert>
      ) : null}

      {!leaderboardQuery.isLoading && !leaderboardQuery.isError && items.length > 0 ? (
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Model
                </th>
                <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Kind
                </th>
                <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Entries
                </th>
                <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Evaluations
                </th>
                <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Avg Payout
                </th>
                <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                  P(Top 1)
                </th>
                <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                  P(In Money)
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {items.map((row) => (
                <tr
                  key={row.investment_model_id}
                  className="hover:bg-gray-50 cursor-pointer"
                  onClick={() => navigate(`/lab/models/${encodeURIComponent(row.investment_model_id)}`)}
                >
                  <td className="px-3 py-2 text-sm text-gray-900 font-medium">{row.model_name}</td>
                  <td className="px-3 py-2 text-sm text-gray-600">{row.model_kind}</td>
                  <td className="px-3 py-2 text-sm text-gray-700 text-right">{row.n_entries}</td>
                  <td className="px-3 py-2 text-sm text-gray-700 text-right">{row.n_evaluations}</td>
                  <td className="px-3 py-2 text-sm text-gray-900 text-right font-medium">
                    {formatPayoutX(row.avg_mean_payout)}
                  </td>
                  <td className="px-3 py-2 text-sm text-gray-700 text-right">{formatPct(row.avg_p_top1)}</td>
                  <td className="px-3 py-2 text-sm text-gray-700 text-right">{formatPct(row.avg_p_in_money)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : null}
    </Card>
  );
}

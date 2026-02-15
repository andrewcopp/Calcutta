import React, { useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';

import { Alert } from '../../components/ui/Alert';
import { Card } from '../../components/ui/Card';
import { Input } from '../../components/ui/Input';
import { LoadingState } from '../../components/ui/LoadingState';
import { labService, ListEvaluationsResponse } from '../../services/labService';

export function EvaluationsTab() {
  const navigate = useNavigate();
  const [modelFilter, setModelFilter] = useState('');

  const evaluationsQuery = useQuery<ListEvaluationsResponse | null>({
    queryKey: ['lab', 'evaluations'],
    queryFn: () => labService.listEvaluations({ limit: 100 }),
  });

  const filteredItems = useMemo(() => {
    const items = evaluationsQuery.data?.items ?? [];
    if (!modelFilter.trim()) return items;
    const lower = modelFilter.toLowerCase();
    return items.filter(
      (e) =>
        e.model_name.toLowerCase().includes(lower) ||
        e.calcutta_name.toLowerCase().includes(lower)
    );
  }, [evaluationsQuery.data, modelFilter]);

  const formatPct = (val?: number | null) => {
    if (val == null) return '-';
    return `${(val * 100).toFixed(1)}%`;
  };

  const formatPayoutX = (val?: number | null) => {
    if (val == null) return '-';
    return `${val.toFixed(3)}x`;
  };

  const formatDate = (dateStr: string) => {
    const d = new Date(dateStr);
    return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
  };

  return (
    <Card>
      <h2 className="text-xl font-semibold mb-4">Evaluations</h2>
      <p className="text-sm text-gray-600 mb-4">
        Simulation results showing how each entry performed. Sorted by mean normalized payout.
      </p>

      <div className="mb-4">
        <Input
          type="text"
          placeholder="Filter by model or calcutta name..."
          value={modelFilter}
          onChange={(e) => setModelFilter(e.target.value)}
          className="max-w-sm"
        />
      </div>

      {evaluationsQuery.isLoading ? <LoadingState label="Loading evaluations..." layout="inline" /> : null}

      {evaluationsQuery.isError ? (
        <Alert variant="error">Failed to load evaluations.</Alert>
      ) : null}

      {!evaluationsQuery.isLoading && !evaluationsQuery.isError && filteredItems.length === 0 ? (
        <Alert variant="info">No evaluations found.</Alert>
      ) : null}

      {!evaluationsQuery.isLoading && !evaluationsQuery.isError && filteredItems.length > 0 ? (
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Model
                </th>
                <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Calcutta
                </th>
                <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Sims
                </th>
                <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Mean Payout
                </th>
                <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Median Payout
                </th>
                <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                  P(Top 1)
                </th>
                <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                  P(In Money)
                </th>
                <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Created
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {filteredItems.map((row) => (
                <tr
                  key={row.id}
                  className="hover:bg-gray-50 cursor-pointer"
                  onClick={() => navigate(`/lab/evaluations/${encodeURIComponent(row.id)}`)}
                >
                  <td className="px-3 py-2 text-sm text-gray-900">
                    <div className="font-medium">{row.model_name}</div>
                    <div className="text-xs text-gray-500">{row.model_kind}</div>
                  </td>
                  <td className="px-3 py-2 text-sm text-gray-700">{row.calcutta_name}</td>
                  <td className="px-3 py-2 text-sm text-gray-600 text-right">{row.n_sims.toLocaleString()}</td>
                  <td className="px-3 py-2 text-sm text-gray-900 text-right font-medium">
                    {formatPayoutX(row.mean_normalized_payout)}
                  </td>
                  <td className="px-3 py-2 text-sm text-gray-700 text-right">
                    {formatPayoutX(row.median_normalized_payout)}
                  </td>
                  <td className="px-3 py-2 text-sm text-gray-700 text-right">{formatPct(row.p_top1)}</td>
                  <td className="px-3 py-2 text-sm text-gray-700 text-right">{formatPct(row.p_in_money)}</td>
                  <td className="px-3 py-2 text-sm text-gray-500">{formatDate(row.created_at)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : null}
    </Card>
  );
}

import React from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';

import { Alert } from '../../../components/ui/Alert';
import { Card } from '../../../components/ui/Card';
import { LoadingState } from '../../../components/ui/LoadingState';
import { labService, EvaluationEntryResult } from '../../../services/labService';
import { cn } from '../../../lib/cn';
import { queryKeys } from '../../../queryKeys';

interface Evaluation {
  id: string;
  n_sims: number;
  mean_normalized_payout?: number | null;
  p_top1?: number | null;
  p_in_money?: number | null;
  created_at: string;
}

interface EvaluationsTabProps {
  evaluation: Evaluation | null;
  isLoading: boolean;
  modelName: string;
  calcuttaId: string;
}

function formatDate(dateStr: string): string {
  const d = new Date(dateStr);
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
}

function formatPayoutX(val?: number | null): string {
  if (val == null) return '-';
  return `${val.toFixed(3)}x`;
}

function formatPct(val?: number | null): string {
  if (val == null) return '-';
  return `${(val * 100).toFixed(1)}%`;
}

function getPayoutColor(payout?: number | null): string {
  if (payout == null) return 'text-gray-400';
  if (payout >= 1.2) return 'text-green-700 font-bold';
  if (payout >= 0.9) return 'text-yellow-700';
  return 'text-red-700';
}

export function EvaluationsTab({ evaluation, isLoading, modelName, calcuttaId }: EvaluationsTabProps) {
  const navigate = useNavigate();

  // Fetch entry results when we have an evaluation
  const entryResultsQuery = useQuery<EvaluationEntryResult[]>({
    queryKey: queryKeys.lab.evaluations.entries(evaluation?.id),
    queryFn: () => (evaluation?.id ? labService.getEvaluationEntryResults(evaluation.id) : Promise.resolve([])),
    enabled: Boolean(evaluation?.id),
  });

  const entryResults = entryResultsQuery.data ?? [];

  const handleEntryClick = (entryResultId: string) => {
    navigate(
      `/lab/models/${encodeURIComponent(modelName)}/calcutta/${encodeURIComponent(calcuttaId)}/entry-results/${encodeURIComponent(entryResultId)}`
    );
  };

  if (isLoading) {
    return <LoadingState label="Loading evaluations..." layout="inline" />;
  }

  if (!evaluation) {
    return <Alert variant="info">No evaluations yet. Run simulations to see how this entry would perform.</Alert>;
  }

  return (
    <div className="space-y-4">
      {/* Compact evaluation metrics header */}
      <div className="bg-white rounded-lg border border-gray-200 p-4">
        <div className="flex flex-wrap items-center gap-6">
          <div>
            <span className="text-xs text-gray-500 uppercase mr-2">Sims:</span>
            <span className="font-medium">{evaluation.n_sims.toLocaleString()}</span>
          </div>
          <div>
            <span className="text-xs text-gray-500 uppercase mr-2">Mean Payout:</span>
            <span className={cn('font-semibold', getPayoutColor(evaluation.mean_normalized_payout))}>
              {formatPayoutX(evaluation.mean_normalized_payout)}
            </span>
          </div>
          <div>
            <span className="text-xs text-gray-500 uppercase mr-2">P(Top 1):</span>
            <span className="font-medium">{formatPct(evaluation.p_top1)}</span>
          </div>
          <div>
            <span className="text-xs text-gray-500 uppercase mr-2">P(In Money):</span>
            <span className="font-medium">{formatPct(evaluation.p_in_money)}</span>
          </div>
          <div className="text-gray-400 text-sm">
            Created: {formatDate(evaluation.created_at)}
          </div>
        </div>
      </div>

      {/* All Entries Ranked by Mean Payout */}
      <Card>
        <h2 className="text-lg font-semibold mb-3">All Entries Ranked by Mean Payout</h2>
        {entryResultsQuery.isLoading ? (
          <LoadingState label="Loading entry results..." layout="inline" />
        ) : entryResults.length === 0 ? (
          <p className="text-gray-500 text-sm">No entry results available for this evaluation.</p>
        ) : (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase">Rank</th>
                  <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase">Entry Name</th>
                  <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase">Mean Payout</th>
                  <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase">P(Top 1)</th>
                  <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase">P(In Money)</th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {entryResults.map((entry) => {
                  const isOurStrategy = entry.entry_name === 'Our Strategy';
                  return (
                    <tr
                      key={entry.entry_name}
                      className={cn(
                        'hover:bg-gray-50 cursor-pointer',
                        isOurStrategy && 'bg-blue-50 font-semibold hover:bg-blue-100'
                      )}
                      onClick={() => handleEntryClick(entry.id)}
                    >
                      <td className="px-3 py-2 text-sm text-gray-700">#{entry.rank}</td>
                      <td className={cn('px-3 py-2 text-sm', isOurStrategy ? 'text-blue-900' : 'text-gray-900')}>
                        {entry.entry_name}
                      </td>
                      <td className="px-3 py-2 text-sm text-gray-700 text-right">{formatPayoutX(entry.mean_normalized_payout)}</td>
                      <td className="px-3 py-2 text-sm text-gray-700 text-right">{formatPct(entry.p_top1)}</td>
                      <td className="px-3 py-2 text-sm text-gray-700 text-right">{formatPct(entry.p_in_money)}</td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </Card>
    </div>
  );
}

import React from 'react';
import { useNavigate } from 'react-router-dom';

import { Alert } from '../../../components/ui/Alert';
import { Card } from '../../../components/ui/Card';
import { LoadingState } from '../../../components/ui/LoadingState';
import { cn } from '../../../lib/cn';

interface Evaluation {
  id: string;
  n_sims: number;
  mean_normalized_payout?: number | null;
  p_top1?: number | null;
  p_in_money?: number | null;
  created_at: string;
}

interface EvaluationsTabProps {
  evaluations: Evaluation[];
  isLoading: boolean;
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

export function EvaluationsTab({ evaluations, isLoading }: EvaluationsTabProps) {
  const navigate = useNavigate();

  // Calculate summary stats
  const latestEval = evaluations.length > 0 ? evaluations[0] : null;
  const bestEval = evaluations.length > 0
    ? [...evaluations].sort((a, b) => (b.mean_normalized_payout ?? 0) - (a.mean_normalized_payout ?? 0))[0]
    : null;

  return (
    <div className="space-y-4">
      {/* Evaluation-specific stats */}
      {!isLoading && evaluations.length > 0 && (
        <div className="grid grid-cols-2 gap-4">
          <div className="bg-white rounded-lg border border-gray-200 p-3">
            <div className="text-xs text-gray-500 uppercase">Latest Run</div>
            <div className={cn('text-lg font-semibold', getPayoutColor(latestEval?.mean_normalized_payout))}>
              {formatPayoutX(latestEval?.mean_normalized_payout)}
            </div>
            <div className="text-xs text-gray-400">
              {latestEval ? formatDate(latestEval.created_at) : '-'}
            </div>
          </div>
          <div className="bg-white rounded-lg border border-gray-200 p-3">
            <div className="text-xs text-gray-500 uppercase">Best Run</div>
            <div className={cn('text-lg font-semibold', getPayoutColor(bestEval?.mean_normalized_payout))}>
              {formatPayoutX(bestEval?.mean_normalized_payout)}
            </div>
            <div className="text-xs text-gray-400">
              {bestEval?.n_sims.toLocaleString()} sims
            </div>
          </div>
        </div>
      )}

      <Card>
        <h2 className="text-lg font-semibold mb-3">Evaluations ({evaluations.length})</h2>
        {isLoading ? <LoadingState label="Loading evaluations..." layout="inline" /> : null}
        {!isLoading && evaluations.length === 0 ? (
          <Alert variant="info">No evaluations yet. Run simulations to see how this entry would perform.</Alert>
        ) : null}
        {!isLoading && evaluations.length > 0 ? (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase">Sims</th>
                  <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase">Mean Payout</th>
                  <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase">P(Top 1)</th>
                  <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase">P(In Money)</th>
                  <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase">Created</th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {evaluations.map((ev) => (
                  <tr
                    key={ev.id}
                    className="hover:bg-gray-50 cursor-pointer"
                    onClick={() => navigate(`/lab/evaluations/${encodeURIComponent(ev.id)}`)}
                  >
                    <td className="px-3 py-2 text-sm text-gray-700 text-right">{ev.n_sims.toLocaleString()}</td>
                    <td className={cn('px-3 py-2 text-sm text-right', getPayoutColor(ev.mean_normalized_payout))}>
                      {formatPayoutX(ev.mean_normalized_payout)}
                    </td>
                    <td className="px-3 py-2 text-sm text-gray-700 text-right">{formatPct(ev.p_top1)}</td>
                    <td className="px-3 py-2 text-sm text-gray-700 text-right">{formatPct(ev.p_in_money)}</td>
                    <td className="px-3 py-2 text-sm text-gray-500">{formatDate(ev.created_at)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : null}
      </Card>
    </div>
  );
}

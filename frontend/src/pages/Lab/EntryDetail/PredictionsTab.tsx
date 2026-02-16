import React from 'react';

import { Card } from '../../../components/ui/Card';
import { cn } from '../../../lib/cn';
import type { EnrichedPrediction } from '../../../types/lab';

type PredSortKey = 'seed' | 'team' | 'rational' | 'predicted' | 'edge';
type SortDir = 'asc' | 'desc';

interface PredictionsTabProps {
  predictions: EnrichedPrediction[];
  sortKey: PredSortKey;
  sortDir: SortDir;
  onSort: (key: PredSortKey) => void;
  optimizerParams?: Record<string, unknown>;
}

function getEdgeColor(edge: number): string {
  if (edge >= 25) return 'bg-green-100';
  if (edge >= 10) return 'bg-green-50';
  if (edge <= -25) return 'bg-red-100';
  if (edge <= -10) return 'bg-red-50';
  return '';
}

function getEdgeTextColor(edge: number): string {
  if (edge >= 10) return 'text-green-700';
  if (edge <= -10) return 'text-red-700';
  return 'text-gray-600';
}

function formatEdge(edge: number): string {
  const sign = edge >= 0 ? '+' : '';
  return `${sign}${edge.toFixed(1)}%`;
}

export function PredictionsTab({ predictions, sortKey, sortDir, onSort, optimizerParams }: PredictionsTabProps) {
  const SortHeader = ({ label, sortKeyValue }: { label: string; sortKeyValue: PredSortKey }) => (
    <button
      type="button"
      onClick={() => onSort(sortKeyValue)}
      className={cn(
        'flex items-center gap-1 text-xs font-medium uppercase tracking-wider',
        sortKey === sortKeyValue ? 'text-blue-600' : 'text-gray-500 hover:text-gray-700'
      )}
    >
      {label}
      {sortKey === sortKeyValue ? (sortDir === 'desc' ? ' ▼' : ' ▲') : ''}
    </button>
  );

  // Extract prediction inputs from optimizer params
  const estimatedParticipants = optimizerParams?.estimated_participants as number | undefined;
  const excludedEntryName = optimizerParams?.excluded_entry_name as string | undefined;

  // Summary stats
  const totalRational = predictions.reduce((sum, p) => sum + p.naive_points, 0);
  const totalPredicted = predictions.reduce((sum, p) => sum + p.predicted_bid_points, 0);

  // Find biggest edge opportunities (where market might undervalue)
  const sortedByEdge = [...predictions].sort((a, b) => b.edge_percent - a.edge_percent);
  const topUndervalued = sortedByEdge.filter(p => p.edge_percent > 0).slice(0, 3);
  const topOvervalued = sortedByEdge.filter(p => p.edge_percent < 0).slice(-3).reverse();

  return (
    <div className="space-y-4">
      {/* Prediction inputs */}
      {(estimatedParticipants != null || excludedEntryName) && (
        <div className="bg-gray-50 rounded-lg border border-gray-200 p-3">
          <div className="text-xs text-gray-500 uppercase mb-2">Prediction Inputs</div>
          <div className="flex gap-6 text-sm">
            {estimatedParticipants != null && (
              <div>
                <span className="text-gray-500">Estimated Participants:</span>{' '}
                <span className="font-medium">{estimatedParticipants}</span>
              </div>
            )}
            {excludedEntryName && (
              <div>
                <span className="text-gray-500">Excluded Entry:</span>{' '}
                <span className="font-medium">{excludedEntryName}</span>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Summary stats */}
      <div className="grid grid-cols-2 gap-4">
        <div className="bg-white rounded-lg border border-gray-200 p-3">
          <div className="text-xs text-gray-500 uppercase">Total Rational Budget</div>
          <div className="text-lg font-semibold">{totalRational.toLocaleString()} pts</div>
        </div>
        <div className="bg-white rounded-lg border border-gray-200 p-3">
          <div className="text-xs text-gray-500 uppercase">Total Predicted Market</div>
          <div className="text-lg font-semibold">{totalPredicted.toLocaleString()} pts</div>
        </div>
      </div>

      {/* Top opportunities */}
      {(topUndervalued.length > 0 || topOvervalued.length > 0) && (
        <div className="grid grid-cols-2 gap-4">
          {topUndervalued.length > 0 && (
            <div className="bg-green-50 rounded-lg border border-green-200 p-3">
              <div className="text-xs text-green-700 uppercase mb-2">Potentially Undervalued</div>
              {topUndervalued.map(p => (
                <div key={p.team_id} className="text-sm text-green-800">
                  {p.school_name} ({p.seed}) {formatEdge(p.edge_percent)}
                </div>
              ))}
            </div>
          )}
          {topOvervalued.length > 0 && (
            <div className="bg-red-50 rounded-lg border border-red-200 p-3">
              <div className="text-xs text-red-700 uppercase mb-2">Potentially Overvalued</div>
              {topOvervalued.map(p => (
                <div key={p.team_id} className="text-sm text-red-800">
                  {p.school_name} ({p.seed}) {formatEdge(p.edge_percent)}
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      <Card>
        <h2 className="text-lg font-semibold mb-3">Market Predictions</h2>
        <p className="text-sm text-gray-600 mb-3">
          Comparing rational vs predicted market behavior.
          <strong className="ml-2">Rational</strong> = what perfect-info bidders would bid (proportional to expected value).
          <strong className="ml-2">Predicted</strong> = model's prediction of actual market bids (blind auction).
          <strong className="ml-2">Edge</strong> = (rational - predicted) / rational — where market may be over/undervaluing.
        </p>
        <div className="overflow-x-auto">
          <table className="min-w-full">
            <thead className="bg-gray-50 border-b border-gray-200">
              <tr>
                <th className="px-3 py-2 text-left">
                  <SortHeader label="Team" sortKeyValue="team" />
                </th>
                <th className="px-3 py-2 text-center">
                  <SortHeader label="Seed" sortKeyValue="seed" />
                </th>
                <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase">Region</th>
                <th className="px-3 py-2 text-right">
                  <SortHeader label="Rational" sortKeyValue="rational" />
                </th>
                <th className="px-3 py-2 text-right">
                  <SortHeader label="Predicted" sortKeyValue="predicted" />
                </th>
                <th className="px-3 py-2 text-right">
                  <SortHeader label="Edge" sortKeyValue="edge" />
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {predictions.map((pred) => (
                <tr key={pred.team_id} className={cn('hover:bg-gray-50', getEdgeColor(pred.edge_percent))}>
                  <td className="px-3 py-2 text-sm font-medium text-gray-900">{pred.school_name}</td>
                  <td className="px-3 py-2 text-sm text-gray-700 text-center">{pred.seed}</td>
                  <td className="px-3 py-2 text-sm text-gray-500">{pred.region}</td>
                  <td className="px-3 py-2 text-sm text-gray-900 text-right font-medium tabular-nums">{pred.naive_points}</td>
                  <td className="px-3 py-2 text-sm text-gray-900 text-right font-medium tabular-nums">{pred.predicted_bid_points}</td>
                  <td className={cn('px-3 py-2 text-sm text-right font-medium tabular-nums', getEdgeTextColor(pred.edge_percent))}>
                    {formatEdge(pred.edge_percent)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </Card>
    </div>
  );
}

export type { PredSortKey };

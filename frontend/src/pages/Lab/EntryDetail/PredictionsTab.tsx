import React from 'react';

import { Card } from '../../../components/ui/Card';
import { cn } from '../../../lib/cn';
import type { EnrichedPrediction, SortDir } from '../../../schemas/lab';

type PredSortKey = 'seed' | 'team' | 'rational' | 'predicted' | 'edge';

interface PredictionsTabProps {
  predictions: EnrichedPrediction[];
  sortKey: PredSortKey;
  sortDir: SortDir;
  onSort: (key: PredSortKey) => void;
  optimizerParams?: Record<string, unknown>;
}

function getEdgeColor(edge: number): string {
  if (edge >= 25) return 'bg-success/10';
  if (edge >= 10) return 'bg-success/10';
  if (edge <= -25) return 'bg-destructive/10';
  if (edge <= -10) return 'bg-destructive/10';
  return '';
}

function getEdgeTextColor(edge: number): string {
  if (edge >= 10) return 'text-success';
  if (edge <= -10) return 'text-destructive';
  return 'text-muted-foreground';
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
        sortKey === sortKeyValue ? 'text-primary' : 'text-muted-foreground hover:text-foreground',
      )}
    >
      {label}
      {sortKey === sortKeyValue ? (sortDir === 'desc' ? ' \u25BC' : ' \u25B2') : ''}
    </button>
  );

  // Extract prediction inputs from optimizer params
  const estimatedParticipants = optimizerParams?.estimated_participants as number | undefined;
  const excludedEntryName = optimizerParams?.excluded_entry_name as string | undefined;

  // Summary stats
  const totalRational = predictions.reduce((sum, p) => sum + p.rationalPoints, 0);
  const totalPredicted = predictions.reduce((sum, p) => sum + p.predictedBidPoints, 0);

  // Find biggest edge opportunities (where market might undervalue)
  const sortedByEdge = [...predictions].sort((a, b) => b.edgePercent - a.edgePercent);
  const topUndervalued = sortedByEdge.filter((p) => p.edgePercent > 0).slice(0, 3);
  const topOvervalued = sortedByEdge
    .filter((p) => p.edgePercent < 0)
    .slice(-3)
    .reverse();

  return (
    <div className="space-y-4">
      {/* Prediction inputs */}
      {(estimatedParticipants != null || excludedEntryName) && (
        <div className="bg-accent rounded-lg border border-border p-3">
          <div className="text-xs text-muted-foreground uppercase mb-2">Prediction Inputs</div>
          <div className="flex gap-6 text-sm">
            {estimatedParticipants != null && (
              <div>
                <span className="text-muted-foreground">Estimated Participants:</span>{' '}
                <span className="font-medium">{estimatedParticipants}</span>
              </div>
            )}
            {excludedEntryName && (
              <div>
                <span className="text-muted-foreground">Excluded Entry:</span>{' '}
                <span className="font-medium">{excludedEntryName}</span>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Summary stats */}
      <div className="grid grid-cols-2 gap-4">
        <div className="bg-card rounded-lg border border-border p-3">
          <div className="text-xs text-muted-foreground uppercase">Total Rational Budget</div>
          <div className="text-lg font-semibold">{totalRational.toLocaleString()} credits</div>
        </div>
        <div className="bg-card rounded-lg border border-border p-3">
          <div className="text-xs text-muted-foreground uppercase">Total Predicted Market</div>
          <div className="text-lg font-semibold">{totalPredicted.toLocaleString()} credits</div>
        </div>
      </div>

      {/* Top opportunities */}
      {(topUndervalued.length > 0 || topOvervalued.length > 0) && (
        <div className="grid grid-cols-2 gap-4">
          {topUndervalued.length > 0 && (
            <div className="bg-success/10 rounded-lg border border-green-200 p-3">
              <div className="text-xs text-success uppercase mb-2">Potentially Undervalued</div>
              {topUndervalued.map((p) => (
                <div key={p.teamId} className="text-sm text-success">
                  {p.schoolName} ({p.seed}) {formatEdge(p.edgePercent)}
                </div>
              ))}
            </div>
          )}
          {topOvervalued.length > 0 && (
            <div className="bg-destructive/10 rounded-lg border border-red-200 p-3">
              <div className="text-xs text-destructive uppercase mb-2">Potentially Overvalued</div>
              {topOvervalued.map((p) => (
                <div key={p.teamId} className="text-sm text-destructive">
                  {p.schoolName} ({p.seed}) {formatEdge(p.edgePercent)}
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      <Card>
        <h2 className="text-lg font-semibold mb-3">Market Predictions</h2>
        <p className="text-sm text-muted-foreground mb-3">
          Comparing rational vs predicted market behavior.
          <strong className="ml-2">Rational</strong> = what perfect-info bidders would bid (proportional to expected
          value).
          <strong className="ml-2">Predicted</strong> = model's prediction of actual market bids (blind auction).
          <strong className="ml-2">Edge</strong> = (rational - predicted) / rational — where market may be
          over/undervaluing.
        </p>
        <div className="overflow-x-auto">
          <table className="min-w-full">
            <thead className="bg-accent border-b border-border">
              <tr>
                <th className="px-3 py-2 text-left">
                  <SortHeader label="Team" sortKeyValue="team" />
                </th>
                <th className="px-3 py-2 text-center">
                  <SortHeader label="Seed" sortKeyValue="seed" />
                </th>
                <th className="px-3 py-2 text-left text-xs font-medium text-muted-foreground uppercase">Region</th>
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
                <tr key={pred.teamId} className={cn('hover:bg-accent', getEdgeColor(pred.edgePercent))}>
                  <td className="px-3 py-2 text-sm font-medium text-foreground">{pred.schoolName}</td>
                  <td className="px-3 py-2 text-sm text-foreground text-center">{pred.seed}</td>
                  <td className="px-3 py-2 text-sm text-muted-foreground">{pred.region}</td>
                  <td className="px-3 py-2 text-sm text-foreground text-right font-medium tabular-nums">
                    {pred.rationalPoints}
                  </td>
                  <td className="px-3 py-2 text-sm text-foreground text-right font-medium tabular-nums">
                    {pred.predictedBidPoints}
                  </td>
                  <td
                    className={cn(
                      'px-3 py-2 text-sm text-right font-medium tabular-nums',
                      getEdgeTextColor(pred.edgePercent),
                    )}
                  >
                    {formatEdge(pred.edgePercent)}
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

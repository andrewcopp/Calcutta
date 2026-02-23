import React from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';

import { Alert } from '../../../components/ui/Alert';
import { Card } from '../../../components/ui/Card';
import { LoadingState } from '../../../components/ui/LoadingState';
import { labService } from '../../../services/labService';
import type { EvaluationEntryResult } from '../../../schemas/lab';
import { cn } from '../../../lib/cn';
import { queryKeys } from '../../../queryKeys';
import { formatPayoutX, formatPct, getPayoutColor } from '../../../utils/labFormatters';
import { formatDate } from '../../../utils/format';

interface Evaluation {
  id: string;
  nSims: number;
  meanNormalizedPayout?: number | null;
  pTop1?: number | null;
  pInMoney?: number | null;
  createdAt: string;
}

interface EvaluationsTabProps {
  evaluation: Evaluation | null;
  isLoading: boolean;
  modelName: string;
  calcuttaId: string;
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
      `/lab/models/${encodeURIComponent(modelName)}/calcutta/${encodeURIComponent(calcuttaId)}/entry-results/${encodeURIComponent(entryResultId)}`,
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
      <div className="bg-card rounded-lg border border-border p-4">
        <div className="flex flex-wrap items-center gap-6">
          <div>
            <span className="text-xs text-muted-foreground uppercase mr-2">Sims:</span>
            <span className="font-medium">{evaluation.nSims.toLocaleString()}</span>
          </div>
          <div>
            <span className="text-xs text-muted-foreground uppercase mr-2">Mean Payout:</span>
            <span className={cn('font-semibold', getPayoutColor(evaluation.meanNormalizedPayout))}>
              {formatPayoutX(evaluation.meanNormalizedPayout)}
            </span>
          </div>
          <div>
            <span className="text-xs text-muted-foreground uppercase mr-2">P(Top 1):</span>
            <span className="font-medium">{formatPct(evaluation.pTop1)}</span>
          </div>
          <div>
            <span className="text-xs text-muted-foreground uppercase mr-2">P(In Money):</span>
            <span className="font-medium">{formatPct(evaluation.pInMoney)}</span>
          </div>
          <div className="text-muted-foreground/60 text-sm">Created: {formatDate(evaluation.createdAt)}</div>
        </div>
      </div>

      {/* All Entries Ranked by Mean Payout */}
      <Card>
        <h2 className="text-lg font-semibold mb-3">All Entries Ranked by Mean Payout</h2>
        {entryResultsQuery.isLoading ? (
          <LoadingState label="Loading entry results..." layout="inline" />
        ) : entryResults.length === 0 ? (
          <p className="text-muted-foreground text-sm">No entry results available for this evaluation.</p>
        ) : (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-border">
              <thead className="bg-accent">
                <tr>
                  <th className="px-3 py-2 text-left text-xs font-medium text-muted-foreground uppercase">Rank</th>
                  <th className="px-3 py-2 text-left text-xs font-medium text-muted-foreground uppercase">
                    Entry Name
                  </th>
                  <th className="px-3 py-2 text-right text-xs font-medium text-muted-foreground uppercase">
                    Mean Payout
                  </th>
                  <th className="px-3 py-2 text-right text-xs font-medium text-muted-foreground uppercase">P(Top 1)</th>
                  <th className="px-3 py-2 text-right text-xs font-medium text-muted-foreground uppercase">
                    P(In Money)
                  </th>
                </tr>
              </thead>
              <tbody className="bg-card divide-y divide-border">
                {entryResults.map((entry) => {
                  const isOurStrategy = entry.entryName === 'Our Strategy';
                  return (
                    <tr
                      key={entry.entryName}
                      className={cn(
                        'hover:bg-accent cursor-pointer',
                        isOurStrategy && 'bg-primary/10 font-semibold hover:bg-blue-100',
                      )}
                      onClick={() => handleEntryClick(entry.id)}
                    >
                      <td className="px-3 py-2 text-sm text-foreground">#{entry.rank}</td>
                      <td className={cn('px-3 py-2 text-sm', isOurStrategy ? 'text-blue-900' : 'text-foreground')}>
                        {entry.entryName}
                      </td>
                      <td className="px-3 py-2 text-sm text-foreground text-right">
                        {formatPayoutX(entry.meanNormalizedPayout)}
                      </td>
                      <td className="px-3 py-2 text-sm text-foreground text-right">{formatPct(entry.pTop1)}</td>
                      <td className="px-3 py-2 text-sm text-foreground text-right">{formatPct(entry.pInMoney)}</td>
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

import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { useNavigate, useParams } from 'react-router-dom';

import { ErrorState } from '../../components/ui/ErrorState';
import { Breadcrumb } from '../../components/ui/Breadcrumb';
import { Card } from '../../components/ui/Card';
import { LoadingState } from '../../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../../components/ui/Page';
import { labService } from '../../services/labService';
import type { EvaluationDetail, EvaluationEntryResult } from '../../schemas/lab';
import { cn } from '../../lib/cn';
import { queryKeys } from '../../queryKeys';
import { formatPayoutX, formatPct } from '../../utils/labFormatters';
import { formatDate } from '../../utils/format';

export function EvaluationDetailPage() {
  const { evaluationId } = useParams<{
    evaluationId: string;
  }>();
  const navigate = useNavigate();

  const evaluationQuery = useQuery<EvaluationDetail | null>({
    queryKey: queryKeys.lab.evaluations.detail(evaluationId),
    queryFn: () => (evaluationId ? labService.getEvaluation(evaluationId) : Promise.resolve(null)),
    enabled: Boolean(evaluationId),
  });

  const entryResultsQuery = useQuery<EvaluationEntryResult[]>({
    queryKey: queryKeys.lab.evaluations.entries(evaluationId),
    queryFn: () => (evaluationId ? labService.getEvaluationEntryResults(evaluationId) : Promise.resolve([])),
    enabled: Boolean(evaluationId),
  });

  const evaluation = evaluationQuery.data;
  const entryResults = entryResultsQuery.data ?? [];

  if (evaluationQuery.isLoading) {
    return (
      <PageContainer>
        <LoadingState label="Loading evaluation..." />
      </PageContainer>
    );
  }

  if (evaluationQuery.isError || !evaluation) {
    return (
      <PageContainer>
        <ErrorState
          error={evaluationQuery.error ?? 'Failed to load evaluation.'}
          onRetry={() => evaluationQuery.refetch()}
        />
      </PageContainer>
    );
  }

  return (
    <PageContainer>
      <Breadcrumb
        items={[
          { label: 'Lab', href: '/lab' },
          { label: 'Evaluations', href: '/lab?tab=evaluations' },
          { label: `${evaluation.modelName} / ${evaluation.calcuttaName}` },
        ]}
      />

      <PageHeader title="Evaluation Results" subtitle={`${evaluation.modelName} on ${evaluation.calcuttaName}`} />

      <Card className="mb-6">
        <h2 className="text-lg font-semibold mb-3">Simulation Details</h2>
        <dl className="grid grid-cols-2 md:grid-cols-3 gap-4 text-sm">
          <div>
            <dt className="text-muted-foreground">Model</dt>
            <dd className="font-medium">
              <button
                type="button"
                className="text-primary hover:underline"
                onClick={() =>
                  navigate(
                    `/lab/models/${encodeURIComponent(evaluation.modelName)}/calcutta/${encodeURIComponent(evaluation.calcuttaId)}`,
                  )
                }
              >
                {evaluation.modelName}
              </button>
              <span className="text-muted-foreground ml-1">({evaluation.modelKind})</span>
            </dd>
          </div>
          <div>
            <dt className="text-muted-foreground">Calcutta</dt>
            <dd className="font-medium">{evaluation.calcuttaName}</dd>
          </div>
          <div>
            <dt className="text-muted-foreground">Starting State</dt>
            <dd className="font-medium">{evaluation.startingStateKey}</dd>
          </div>
          <div>
            <dt className="text-muted-foreground">Simulations</dt>
            <dd className="font-medium">{evaluation.nSims.toLocaleString()}</dd>
          </div>
          <div>
            <dt className="text-muted-foreground">Seed</dt>
            <dd className="font-medium">{evaluation.seed}</dd>
          </div>
          <div>
            <dt className="text-muted-foreground">Created</dt>
            <dd className="font-medium">{formatDate(evaluation.createdAt, true)}</dd>
          </div>
        </dl>
      </Card>

      <Card className="mb-6">
        <h2 className="text-lg font-semibold mb-3">Results</h2>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-6">
          <div className="text-center p-4 bg-accent rounded-lg">
            <div className="text-sm text-muted-foreground mb-1">Mean Normalized Payout</div>
            <div className="text-2xl font-bold text-foreground">
              {formatPayoutX(evaluation.meanNormalizedPayout, 4)}
            </div>
            <div className="text-xs text-muted-foreground mt-1">The key metric</div>
          </div>
          <div className="text-center p-4 bg-accent rounded-lg">
            <div className="text-sm text-muted-foreground mb-1">Median Normalized Payout</div>
            <div className="text-2xl font-bold text-foreground">
              {formatPayoutX(evaluation.medianNormalizedPayout, 4)}
            </div>
          </div>
          <div className="text-center p-4 bg-accent rounded-lg">
            <div className="text-sm text-muted-foreground mb-1">P(Top 1)</div>
            <div className="text-2xl font-bold text-foreground">{formatPct(evaluation.pTop1, 2)}</div>
            <div className="text-xs text-muted-foreground mt-1">Probability of winning</div>
          </div>
          <div className="text-center p-4 bg-accent rounded-lg">
            <div className="text-sm text-muted-foreground mb-1">P(In Money)</div>
            <div className="text-2xl font-bold text-foreground">{formatPct(evaluation.pInMoney, 2)}</div>
            <div className="text-xs text-muted-foreground mt-1">Probability of payout</div>
          </div>
        </div>

        {evaluation.ourRank != null ? (
          <div className="mt-4 p-4 bg-primary/10 rounded-lg text-center">
            <div className="text-sm text-primary mb-1">Our Rank (median)</div>
            <div className="text-2xl font-bold text-blue-900">#{evaluation.ourRank}</div>
          </div>
        ) : null}
      </Card>

      <Card className="mb-6">
        <h2 className="text-lg font-semibold mb-3">All Entries Ranked by Mean Payout</h2>
        {entryResultsQuery.isLoading ? (
          <LoadingState label="Loading entry results..." />
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
                    <tr key={entry.entryName} className={cn(isOurStrategy && 'bg-primary/10 font-semibold')}>
                      <td className="px-3 py-2 text-sm text-foreground">#{entry.rank}</td>
                      <td className={cn('px-3 py-2 text-sm', isOurStrategy ? 'text-blue-900' : 'text-foreground')}>
                        {entry.entryName}
                      </td>
                      <td className="px-3 py-2 text-sm text-foreground text-right">
                        {formatPayoutX(entry.meanNormalizedPayout, 4)}
                      </td>
                      <td className="px-3 py-2 text-sm text-foreground text-right">{formatPct(entry.pTop1, 2)}</td>
                      <td className="px-3 py-2 text-sm text-foreground text-right">{formatPct(entry.pInMoney, 2)}</td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </Card>
    </PageContainer>
  );
}

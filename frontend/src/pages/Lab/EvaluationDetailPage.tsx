import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { useNavigate, useParams } from 'react-router-dom';

import { Alert } from '../../components/ui/Alert';
import { Breadcrumb } from '../../components/ui/Breadcrumb';
import { Card } from '../../components/ui/Card';
import { LoadingState } from '../../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../../components/ui/Page';
import { labService } from '../../services/labService';
import type { EvaluationDetail, EvaluationEntryResult } from '../../types/lab';
import { cn } from '../../lib/cn';
import { queryKeys } from '../../queryKeys';
import { formatDate, formatPayoutX, formatPct } from '../../utils/labFormatters';

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
        <Alert variant="error">Failed to load evaluation.</Alert>
      </PageContainer>
    );
  }

  return (
    <PageContainer>
      <Breadcrumb
        items={[
          { label: 'Lab', href: '/lab' },
          { label: 'Evaluations', href: '/lab?tab=evaluations' },
          { label: `${evaluation.model_name} / ${evaluation.calcutta_name}` },
        ]}
      />

      <PageHeader
        title="Evaluation Results"
        subtitle={`${evaluation.model_name} on ${evaluation.calcutta_name}`}
      />

      <Card className="mb-6">
        <h2 className="text-lg font-semibold mb-3">Simulation Details</h2>
        <dl className="grid grid-cols-2 md:grid-cols-3 gap-4 text-sm">
          <div>
            <dt className="text-gray-500">Model</dt>
            <dd className="font-medium">
              <button
                type="button"
                className="text-blue-600 hover:underline"
                onClick={() => navigate(`/lab/models/${encodeURIComponent(evaluation.model_name)}/calcutta/${encodeURIComponent(evaluation.calcutta_id)}`)}
              >
                {evaluation.model_name}
              </button>
              <span className="text-gray-500 ml-1">({evaluation.model_kind})</span>
            </dd>
          </div>
          <div>
            <dt className="text-gray-500">Calcutta</dt>
            <dd className="font-medium">{evaluation.calcutta_name}</dd>
          </div>
          <div>
            <dt className="text-gray-500">Starting State</dt>
            <dd className="font-medium">{evaluation.starting_state_key}</dd>
          </div>
          <div>
            <dt className="text-gray-500">Simulations</dt>
            <dd className="font-medium">{evaluation.n_sims.toLocaleString()}</dd>
          </div>
          <div>
            <dt className="text-gray-500">Seed</dt>
            <dd className="font-medium">{evaluation.seed}</dd>
          </div>
          <div>
            <dt className="text-gray-500">Created</dt>
            <dd className="font-medium">{formatDate(evaluation.created_at, true)}</dd>
          </div>
        </dl>
      </Card>

      <Card className="mb-6">
        <h2 className="text-lg font-semibold mb-3">Results</h2>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-6">
          <div className="text-center p-4 bg-gray-50 rounded-lg">
            <div className="text-sm text-gray-500 mb-1">Mean Normalized Payout</div>
            <div className="text-2xl font-bold text-gray-900">
              {formatPayoutX(evaluation.mean_normalized_payout, 4)}
            </div>
            <div className="text-xs text-gray-500 mt-1">The key metric</div>
          </div>
          <div className="text-center p-4 bg-gray-50 rounded-lg">
            <div className="text-sm text-gray-500 mb-1">Median Normalized Payout</div>
            <div className="text-2xl font-bold text-gray-700">
              {formatPayoutX(evaluation.median_normalized_payout, 4)}
            </div>
          </div>
          <div className="text-center p-4 bg-gray-50 rounded-lg">
            <div className="text-sm text-gray-500 mb-1">P(Top 1)</div>
            <div className="text-2xl font-bold text-gray-700">
              {formatPct(evaluation.p_top1, 2)}
            </div>
            <div className="text-xs text-gray-500 mt-1">Probability of winning</div>
          </div>
          <div className="text-center p-4 bg-gray-50 rounded-lg">
            <div className="text-sm text-gray-500 mb-1">P(In Money)</div>
            <div className="text-2xl font-bold text-gray-700">
              {formatPct(evaluation.p_in_money, 2)}
            </div>
            <div className="text-xs text-gray-500 mt-1">Probability of payout</div>
          </div>
        </div>

        {evaluation.our_rank != null ? (
          <div className="mt-4 p-4 bg-blue-50 rounded-lg text-center">
            <div className="text-sm text-blue-600 mb-1">Our Rank (median)</div>
            <div className="text-2xl font-bold text-blue-900">#{evaluation.our_rank}</div>
          </div>
        ) : null}
      </Card>

      <Card className="mb-6">
        <h2 className="text-lg font-semibold mb-3">All Entries Ranked by Mean Payout</h2>
        {entryResultsQuery.isLoading ? (
          <LoadingState label="Loading entry results..." />
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
                        isOurStrategy && 'bg-blue-50 font-semibold'
                      )}
                    >
                      <td className="px-3 py-2 text-sm text-gray-700">#{entry.rank}</td>
                      <td className={cn('px-3 py-2 text-sm', isOurStrategy ? 'text-blue-900' : 'text-gray-900')}>
                        {entry.entry_name}
                      </td>
                      <td className="px-3 py-2 text-sm text-gray-700 text-right">{formatPayoutX(entry.mean_normalized_payout, 4)}</td>
                      <td className="px-3 py-2 text-sm text-gray-700 text-right">{formatPct(entry.p_top1, 2)}</td>
                      <td className="px-3 py-2 text-sm text-gray-700 text-right">{formatPct(entry.p_in_money, 2)}</td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </Card>

      <Card>
        <h2 className="text-lg font-semibold mb-3">Navigation</h2>
        <div className="flex gap-4">
          <button
            type="button"
            className="text-blue-600 hover:underline text-sm"
            onClick={() => navigate(`/lab/models/${encodeURIComponent(evaluation.model_name)}/calcutta/${encodeURIComponent(evaluation.calcutta_id)}`)}
          >
            View Entry Details
          </button>
        </div>
      </Card>
    </PageContainer>
  );
}

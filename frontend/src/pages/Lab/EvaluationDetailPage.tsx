import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { useNavigate, useParams } from 'react-router-dom';

import { Alert } from '../../components/ui/Alert';
import { Breadcrumb } from '../../components/ui/Breadcrumb';
import { Card } from '../../components/ui/Card';
import { LoadingState } from '../../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../../components/ui/Page';
import { labService, EvaluationDetail } from '../../services/labService';

export function EvaluationDetailPage() {
  const { evaluationId } = useParams<{ evaluationId: string }>();
  const navigate = useNavigate();

  const evaluationQuery = useQuery<EvaluationDetail | null>({
    queryKey: ['lab', 'evaluations', evaluationId],
    queryFn: () => (evaluationId ? labService.getEvaluation(evaluationId) : Promise.resolve(null)),
    enabled: Boolean(evaluationId),
  });

  const evaluation = evaluationQuery.data;

  const formatDate = (dateStr: string) => {
    const d = new Date(dateStr);
    return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric', hour: '2-digit', minute: '2-digit' });
  };

  const formatPayoutX = (val?: number | null) => {
    if (val == null) return '-';
    return `${val.toFixed(4)}x`;
  };

  const formatPct = (val?: number | null) => {
    if (val == null) return '-';
    return `${(val * 100).toFixed(2)}%`;
  };

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
                onClick={() => navigate(`/lab/entries/${encodeURIComponent(evaluation.entry_id)}`)}
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
            <dd className="font-medium">{formatDate(evaluation.created_at)}</dd>
          </div>
        </dl>
      </Card>

      <Card className="mb-6">
        <h2 className="text-lg font-semibold mb-3">Results</h2>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-6">
          <div className="text-center p-4 bg-gray-50 rounded-lg">
            <div className="text-sm text-gray-500 mb-1">Mean Normalized Payout</div>
            <div className="text-2xl font-bold text-gray-900">
              {formatPayoutX(evaluation.mean_normalized_payout)}
            </div>
            <div className="text-xs text-gray-500 mt-1">The key metric</div>
          </div>
          <div className="text-center p-4 bg-gray-50 rounded-lg">
            <div className="text-sm text-gray-500 mb-1">Median Normalized Payout</div>
            <div className="text-2xl font-bold text-gray-700">
              {formatPayoutX(evaluation.median_normalized_payout)}
            </div>
          </div>
          <div className="text-center p-4 bg-gray-50 rounded-lg">
            <div className="text-sm text-gray-500 mb-1">P(Top 1)</div>
            <div className="text-2xl font-bold text-gray-700">
              {formatPct(evaluation.p_top1)}
            </div>
            <div className="text-xs text-gray-500 mt-1">Probability of winning</div>
          </div>
          <div className="text-center p-4 bg-gray-50 rounded-lg">
            <div className="text-sm text-gray-500 mb-1">P(In Money)</div>
            <div className="text-2xl font-bold text-gray-700">
              {formatPct(evaluation.p_in_money)}
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

      <Card>
        <h2 className="text-lg font-semibold mb-3">Navigation</h2>
        <div className="flex gap-4">
          <button
            type="button"
            className="text-blue-600 hover:underline text-sm"
            onClick={() => navigate(`/lab/entries/${encodeURIComponent(evaluation.entry_id)}`)}
          >
            View Entry Details
          </button>
          {evaluation.simulated_calcutta_id ? (
            <button
              type="button"
              className="text-blue-600 hover:underline text-sm"
              onClick={() => navigate(`/sandbox/calcuttas/${encodeURIComponent(evaluation.simulated_calcutta_id!)}`)}
            >
              View Simulated Calcutta
            </button>
          ) : null}
        </div>
      </Card>
    </PageContainer>
  );
}

import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { useNavigate, useParams } from 'react-router-dom';

import { Alert } from '../../components/ui/Alert';
import { Breadcrumb } from '../../components/ui/Breadcrumb';
import { Card } from '../../components/ui/Card';
import { LoadingState } from '../../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../../components/ui/Page';
import { labService, InvestmentModel, ListEntriesResponse, ListEvaluationsResponse } from '../../services/labService';

export function ModelDetailPage() {
  const { modelId } = useParams<{ modelId: string }>();
  const navigate = useNavigate();

  const modelQuery = useQuery<InvestmentModel | null>({
    queryKey: ['lab', 'models', modelId],
    queryFn: () => (modelId ? labService.getModel(modelId) : Promise.resolve(null)),
    enabled: Boolean(modelId),
  });

  const entriesQuery = useQuery<ListEntriesResponse | null>({
    queryKey: ['lab', 'entries', { investment_model_id: modelId }],
    queryFn: () => (modelId ? labService.listEntries({ investment_model_id: modelId, limit: 50 }) : Promise.resolve(null)),
    enabled: Boolean(modelId),
  });

  const evaluationsQuery = useQuery<ListEvaluationsResponse | null>({
    queryKey: ['lab', 'evaluations', { investment_model_id: modelId }],
    queryFn: () => (modelId ? labService.listEvaluations({ investment_model_id: modelId, limit: 50 }) : Promise.resolve(null)),
    enabled: Boolean(modelId),
  });

  const model = modelQuery.data;
  const entries = entriesQuery.data?.items ?? [];
  const evaluations = evaluationsQuery.data?.items ?? [];

  const formatDate = (dateStr: string) => {
    const d = new Date(dateStr);
    return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
  };

  const formatPayoutX = (val?: number | null) => {
    if (val == null) return '-';
    return `${val.toFixed(3)}x`;
  };

  const formatPct = (val?: number | null) => {
    if (val == null) return '-';
    return `${(val * 100).toFixed(1)}%`;
  };

  if (modelQuery.isLoading) {
    return (
      <PageContainer>
        <LoadingState label="Loading model..." />
      </PageContainer>
    );
  }

  if (modelQuery.isError || !model) {
    return (
      <PageContainer>
        <Alert variant="error">Failed to load model.</Alert>
      </PageContainer>
    );
  }

  return (
    <PageContainer>
      <Breadcrumb
        items={[
          { label: 'Lab', href: '/lab' },
          { label: 'Models', href: '/lab?tab=models' },
          { label: model.name },
        ]}
      />

      <PageHeader title={model.name} subtitle={`Kind: ${model.kind}`} />

      <Card className="mb-6">
        <h2 className="text-lg font-semibold mb-3">Model Details</h2>
        <dl className="grid grid-cols-2 gap-4 text-sm">
          <div>
            <dt className="text-gray-500">Kind</dt>
            <dd className="font-medium">{model.kind}</dd>
          </div>
          <div>
            <dt className="text-gray-500">Created</dt>
            <dd className="font-medium">{formatDate(model.created_at)}</dd>
          </div>
          <div>
            <dt className="text-gray-500">Entries</dt>
            <dd className="font-medium">{model.n_entries}</dd>
          </div>
          <div>
            <dt className="text-gray-500">Evaluations</dt>
            <dd className="font-medium">{model.n_evaluations}</dd>
          </div>
          {model.notes ? (
            <div className="col-span-2">
              <dt className="text-gray-500">Notes</dt>
              <dd className="font-medium">{model.notes}</dd>
            </div>
          ) : null}
        </dl>
      </Card>

      <Card className="mb-6">
        <h2 className="text-lg font-semibold mb-3">Entries ({entries.length})</h2>
        {entriesQuery.isLoading ? <LoadingState label="Loading entries..." layout="inline" /> : null}
        {!entriesQuery.isLoading && entries.length === 0 ? (
          <Alert variant="info">No entries for this model.</Alert>
        ) : null}
        {!entriesQuery.isLoading && entries.length > 0 ? (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase">Calcutta</th>
                  <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase">State</th>
                  <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase">Optimizer</th>
                  <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase">Evaluations</th>
                  <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase">Created</th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {entries.map((e) => (
                  <tr
                    key={e.id}
                    className="hover:bg-gray-50 cursor-pointer"
                    onClick={() => navigate(`/lab/entries/${encodeURIComponent(e.id)}`)}
                  >
                    <td className="px-3 py-2 text-sm text-gray-900">{e.calcutta_name}</td>
                    <td className="px-3 py-2 text-sm text-gray-600">{e.starting_state_key}</td>
                    <td className="px-3 py-2 text-sm text-gray-600">{e.optimizer_kind}</td>
                    <td className="px-3 py-2 text-sm text-gray-700 text-right">{e.n_evaluations}</td>
                    <td className="px-3 py-2 text-sm text-gray-500">{formatDate(e.created_at)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : null}
      </Card>

      <Card>
        <h2 className="text-lg font-semibold mb-3">Evaluations ({evaluations.length})</h2>
        {evaluationsQuery.isLoading ? <LoadingState label="Loading evaluations..." layout="inline" /> : null}
        {!evaluationsQuery.isLoading && evaluations.length === 0 ? (
          <Alert variant="info">No evaluations for this model.</Alert>
        ) : null}
        {!evaluationsQuery.isLoading && evaluations.length > 0 ? (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase">Calcutta</th>
                  <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase">Sims</th>
                  <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase">Mean Payout</th>
                  <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase">P(Top 1)</th>
                  <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase">P(In Money)</th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {evaluations.map((ev) => (
                  <tr
                    key={ev.id}
                    className="hover:bg-gray-50 cursor-pointer"
                    onClick={() => navigate(`/lab/evaluations/${encodeURIComponent(ev.id)}`)}
                  >
                    <td className="px-3 py-2 text-sm text-gray-900">{ev.calcutta_name}</td>
                    <td className="px-3 py-2 text-sm text-gray-600 text-right">{ev.n_sims.toLocaleString()}</td>
                    <td className="px-3 py-2 text-sm text-gray-900 text-right font-medium">
                      {formatPayoutX(ev.mean_normalized_payout)}
                    </td>
                    <td className="px-3 py-2 text-sm text-gray-700 text-right">{formatPct(ev.p_top1)}</td>
                    <td className="px-3 py-2 text-sm text-gray-700 text-right">{formatPct(ev.p_in_money)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : null}
      </Card>
    </PageContainer>
  );
}

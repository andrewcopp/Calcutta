import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { useNavigate, useParams } from 'react-router-dom';

import { Alert } from '../../components/ui/Alert';
import { Breadcrumb } from '../../components/ui/Breadcrumb';
import { Card } from '../../components/ui/Card';
import { LoadingState } from '../../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../../components/ui/Page';
import { labService, EntryDetail, ListEvaluationsResponse } from '../../services/labService';

export function EntryDetailPage() {
  const { entryId } = useParams<{ entryId: string }>();
  const navigate = useNavigate();

  const entryQuery = useQuery<EntryDetail | null>({
    queryKey: ['lab', 'entries', entryId],
    queryFn: () => (entryId ? labService.getEntry(entryId) : Promise.resolve(null)),
    enabled: Boolean(entryId),
  });

  const evaluationsQuery = useQuery<ListEvaluationsResponse | null>({
    queryKey: ['lab', 'evaluations', { entry_id: entryId }],
    queryFn: () => (entryId ? labService.listEvaluations({ entry_id: entryId, limit: 50 }) : Promise.resolve(null)),
    enabled: Boolean(entryId),
  });

  const entry = entryQuery.data;
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

  if (entryQuery.isLoading) {
    return (
      <PageContainer>
        <LoadingState label="Loading entry..." />
      </PageContainer>
    );
  }

  if (entryQuery.isError || !entry) {
    return (
      <PageContainer>
        <Alert variant="error">Failed to load entry.</Alert>
      </PageContainer>
    );
  }

  const bids = entry.bids_json ?? [];
  const sortedBids = [...bids].sort((a, b) => b.bid_points - a.bid_points);

  return (
    <PageContainer>
      <Breadcrumb
        items={[
          { label: 'Lab', href: '/lab' },
          { label: 'Entries', href: '/lab?tab=entries' },
          { label: `${entry.model_name} / ${entry.calcutta_name}` },
        ]}
      />

      <PageHeader
        title={`${entry.model_name}`}
        subtitle={`Entry for ${entry.calcutta_name}`}
      />

      <Card className="mb-6">
        <h2 className="text-lg font-semibold mb-3">Entry Details</h2>
        <dl className="grid grid-cols-2 gap-4 text-sm">
          <div>
            <dt className="text-gray-500">Model</dt>
            <dd className="font-medium">
              <button
                type="button"
                className="text-blue-600 hover:underline"
                onClick={() => navigate(`/lab/models/${encodeURIComponent(entry.investment_model_id)}`)}
              >
                {entry.model_name}
              </button>
              <span className="text-gray-500 ml-1">({entry.model_kind})</span>
            </dd>
          </div>
          <div>
            <dt className="text-gray-500">Calcutta</dt>
            <dd className="font-medium">{entry.calcutta_name}</dd>
          </div>
          <div>
            <dt className="text-gray-500">Starting State</dt>
            <dd className="font-medium">{entry.starting_state_key}</dd>
          </div>
          <div>
            <dt className="text-gray-500">Optimizer</dt>
            <dd className="font-medium">{entry.optimizer_kind}</dd>
          </div>
          <div>
            <dt className="text-gray-500">Game Outcome</dt>
            <dd className="font-medium">{entry.game_outcome_kind}</dd>
          </div>
          <div>
            <dt className="text-gray-500">Created</dt>
            <dd className="font-medium">{formatDate(entry.created_at)}</dd>
          </div>
        </dl>
      </Card>

      <Card className="mb-6">
        <h2 className="text-lg font-semibold mb-3">Bids ({bids.length} teams)</h2>
        {bids.length === 0 ? (
          <Alert variant="info">No bids in this entry.</Alert>
        ) : (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase">Team ID</th>
                  <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase">Bid Points</th>
                  <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase">Expected ROI</th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {sortedBids.map((bid) => (
                  <tr key={bid.team_id}>
                    <td className="px-3 py-2 text-sm text-gray-900 font-mono">{bid.team_id}</td>
                    <td className="px-3 py-2 text-sm text-gray-900 text-right font-medium">{bid.bid_points}</td>
                    <td className="px-3 py-2 text-sm text-gray-600 text-right">
                      {bid.expected_roi != null ? `${(bid.expected_roi * 100).toFixed(1)}%` : '-'}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </Card>

      <Card>
        <h2 className="text-lg font-semibold mb-3">Evaluations ({evaluations.length})</h2>
        {evaluationsQuery.isLoading ? <LoadingState label="Loading evaluations..." layout="inline" /> : null}
        {!evaluationsQuery.isLoading && evaluations.length === 0 ? (
          <Alert variant="info">No evaluations for this entry yet.</Alert>
        ) : null}
        {!evaluationsQuery.isLoading && evaluations.length > 0 ? (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase">Sims</th>
                  <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase">Seed</th>
                  <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase">Mean Payout</th>
                  <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase">Median Payout</th>
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
                    <td className="px-3 py-2 text-sm text-gray-600 text-right">{ev.seed}</td>
                    <td className="px-3 py-2 text-sm text-gray-900 text-right font-medium">
                      {formatPayoutX(ev.mean_normalized_payout)}
                    </td>
                    <td className="px-3 py-2 text-sm text-gray-700 text-right">
                      {formatPayoutX(ev.median_normalized_payout)}
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
    </PageContainer>
  );
}

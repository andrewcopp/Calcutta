import React, { useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useNavigate, useParams } from 'react-router-dom';

import { Alert } from '../../components/ui/Alert';
import { Breadcrumb } from '../../components/ui/Breadcrumb';
import { Card } from '../../components/ui/Card';
import { LoadingState } from '../../components/ui/LoadingState';
import { labService, EntryDetail, EnrichedBid, ListEvaluationsResponse } from '../../services/labService';
import { cn } from '../../lib/cn';

type SortKey = 'edge' | 'seed' | 'model' | 'naive' | 'team';
type SortDir = 'asc' | 'desc';

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

export function EntryDetailPage() {
  // Support both URL patterns:
  // - /lab/models/:modelName/calcutta/:calcuttaId (new)
  // - /lab/entries/:entryId (legacy)
  const { entryId, modelName, calcuttaId } = useParams<{
    entryId?: string;
    modelName?: string;
    calcuttaId?: string;
  }>();
  const navigate = useNavigate();
  const [sortKey, setSortKey] = useState<SortKey>('edge');
  const [sortDir, setSortDir] = useState<SortDir>('desc');

  // Determine which API to call based on URL params
  const useNewEndpoint = Boolean(modelName && calcuttaId);
  const queryKey = useNewEndpoint
    ? ['lab', 'entries', 'by-model-calcutta', modelName, calcuttaId]
    : ['lab', 'entries', entryId];

  const entryQuery = useQuery<EntryDetail | null>({
    queryKey,
    queryFn: () => {
      if (useNewEndpoint) {
        return labService.getEntryByModelAndCalcutta(modelName!, calcuttaId!);
      }
      return entryId ? labService.getEntry(entryId) : Promise.resolve(null);
    },
    enabled: Boolean(useNewEndpoint || entryId),
  });

  // For evaluations, we need the entry ID from the loaded entry
  const loadedEntryId = entryQuery.data?.id;
  const evaluationsQuery = useQuery<ListEvaluationsResponse | null>({
    queryKey: ['lab', 'evaluations', { entry_id: loadedEntryId }],
    queryFn: () => (loadedEntryId ? labService.listEvaluations({ entry_id: loadedEntryId, limit: 50 }) : Promise.resolve(null)),
    enabled: Boolean(loadedEntryId),
  });

  const entry = entryQuery.data;
  const evaluations = evaluationsQuery.data?.items ?? [];

  // Sort bids based on current sort settings
  const sortedBids = useMemo(() => {
    const bids = entry?.bids ?? [];
    return [...bids].sort((a, b) => {
      let cmp = 0;
      switch (sortKey) {
        case 'edge':
          cmp = Math.abs(b.edge_percent) - Math.abs(a.edge_percent);
          // Secondary sort by direction (positive first)
          if (cmp === 0) cmp = b.edge_percent - a.edge_percent;
          break;
        case 'seed':
          cmp = a.seed - b.seed;
          break;
        case 'model':
          cmp = b.bid_points - a.bid_points;
          break;
        case 'naive':
          cmp = b.naive_points - a.naive_points;
          break;
        case 'team':
          cmp = a.school_name.localeCompare(b.school_name);
          break;
      }
      return sortDir === 'asc' ? -cmp : cmp;
    });
  }, [entry?.bids, sortKey, sortDir]);

  const handleSort = (key: SortKey) => {
    if (sortKey === key) {
      setSortDir(sortDir === 'asc' ? 'desc' : 'asc');
    } else {
      setSortKey(key);
      setSortDir('desc');
    }
  };

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

  const formatEdge = (edge: number) => {
    const sign = edge >= 0 ? '+' : '';
    return `${sign}${edge.toFixed(1)}%`;
  };

  if (entryQuery.isLoading) {
    return (
      <div className="container mx-auto px-4 py-4">
        <LoadingState label="Loading predictions..." />
      </div>
    );
  }

  if (entryQuery.isError || !entry) {
    return (
      <div className="container mx-auto px-4 py-4">
        <Alert variant="error">Failed to load predictions.</Alert>
      </div>
    );
  }

  const bids = entry.bids ?? [];
  const totalBudget = bids.reduce((sum, b) => sum + b.bid_points, 0);

  // Calculate summary stats
  const topOpportunity = [...bids].sort((a, b) => b.edge_percent - a.edge_percent)[0];
  const topAvoid = [...bids].sort((a, b) => a.edge_percent - b.edge_percent)[0];

  const SortHeader = ({ label, sortKeyValue }: { label: string; sortKeyValue: SortKey }) => (
    <button
      type="button"
      onClick={() => handleSort(sortKeyValue)}
      className={cn(
        'flex items-center gap-1 text-xs font-medium uppercase tracking-wider',
        sortKey === sortKeyValue ? 'text-blue-600' : 'text-gray-500 hover:text-gray-700'
      )}
    >
      {label}
      {sortKey === sortKeyValue ? (sortDir === 'desc' ? ' ▼' : ' ▲') : ''}
    </button>
  );

  return (
    <div className="container mx-auto px-4 py-4">
      <Breadcrumb
        items={[
          { label: 'Lab', href: '/lab' },
          { label: entry.model_name, href: `/lab/models/${entry.investment_model_id}` },
          { label: entry.calcutta_name },
        ]}
      />

      {/* Compact header */}
      <div className="flex items-baseline gap-3 mb-4">
        <h1 className="text-xl font-bold text-gray-900">Model Predictions</h1>
        <span className="text-gray-500">
          {entry.model_name} ({entry.model_kind}) → {entry.calcutta_name}
        </span>
      </div>

      {/* Summary stats */}
      <div className="grid grid-cols-4 gap-4 mb-4">
        <div className="bg-white rounded-lg border border-gray-200 p-3">
          <div className="text-xs text-gray-500 uppercase">Teams</div>
          <div className="text-lg font-semibold">{bids.length}</div>
        </div>
        <div className="bg-white rounded-lg border border-gray-200 p-3">
          <div className="text-xs text-gray-500 uppercase">Budget</div>
          <div className="text-lg font-semibold">{totalBudget.toLocaleString()} pts</div>
        </div>
        {topOpportunity && topOpportunity.edge_percent > 0 ? (
          <div className="bg-green-50 rounded-lg border border-green-200 p-3">
            <div className="text-xs text-green-700 uppercase">Top Opportunity</div>
            <div className="text-sm font-semibold text-green-800">
              {topOpportunity.school_name} ({topOpportunity.seed}) {formatEdge(topOpportunity.edge_percent)}
            </div>
          </div>
        ) : (
          <div className="bg-white rounded-lg border border-gray-200 p-3">
            <div className="text-xs text-gray-500 uppercase">Top Opportunity</div>
            <div className="text-sm text-gray-400">None</div>
          </div>
        )}
        {topAvoid && topAvoid.edge_percent < 0 ? (
          <div className="bg-red-50 rounded-lg border border-red-200 p-3">
            <div className="text-xs text-red-700 uppercase">Top Avoid</div>
            <div className="text-sm font-semibold text-red-800">
              {topAvoid.school_name} ({topAvoid.seed}) {formatEdge(topAvoid.edge_percent)}
            </div>
          </div>
        ) : (
          <div className="bg-white rounded-lg border border-gray-200 p-3">
            <div className="text-xs text-gray-500 uppercase">Top Avoid</div>
            <div className="text-sm text-gray-400">None</div>
          </div>
        )}
      </div>

      {/* Predictions table */}
      <Card className="mb-4">
        <h2 className="text-lg font-semibold mb-3">Predicted Bids</h2>
        <p className="text-sm text-gray-600 mb-3">
          <strong>Naive</strong> = expected allocation if everyone bid proportional to seed value.
          <strong className="ml-2">Model</strong> = what this model predicts the market will bid.
          <strong className="ml-2">Edge</strong> = naive minus model (positive = undervalued opportunity, negative = overvalued avoid).
        </p>
        {bids.length === 0 ? (
          <Alert variant="info">No bids in this entry.</Alert>
        ) : (
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
                    <SortHeader label="Naive" sortKeyValue="naive" />
                  </th>
                  <th className="px-3 py-2 text-right">
                    <SortHeader label="Model" sortKeyValue="model" />
                  </th>
                  <th className="px-3 py-2 text-right">
                    <SortHeader label="Edge" sortKeyValue="edge" />
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {sortedBids.map((bid) => (
                  <tr key={bid.team_id} className={cn('hover:bg-gray-50', getEdgeColor(bid.edge_percent))}>
                    <td className="px-3 py-2 text-sm font-medium text-gray-900">{bid.school_name}</td>
                    <td className="px-3 py-2 text-sm text-gray-700 text-center">{bid.seed}</td>
                    <td className="px-3 py-2 text-sm text-gray-500">{bid.region}</td>
                    <td className="px-3 py-2 text-sm text-gray-600 text-right tabular-nums">{bid.naive_points}</td>
                    <td className="px-3 py-2 text-sm text-gray-900 text-right font-medium tabular-nums">{bid.bid_points}</td>
                    <td className={cn('px-3 py-2 text-sm text-right font-medium tabular-nums', getEdgeTextColor(bid.edge_percent))}>
                      {formatEdge(bid.edge_percent)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </Card>

      {/* Evaluations */}
      <Card>
        <h2 className="text-lg font-semibold mb-3">Evaluations ({evaluations.length})</h2>
        {evaluationsQuery.isLoading ? <LoadingState label="Loading evaluations..." layout="inline" /> : null}
        {!evaluationsQuery.isLoading && evaluations.length === 0 ? (
          <Alert variant="info">No evaluations yet. Run simulations to see how this prediction would perform.</Alert>
        ) : null}
        {!evaluationsQuery.isLoading && evaluations.length > 0 ? (
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
                    <td className="px-3 py-2 text-sm text-gray-900 text-right font-medium">
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

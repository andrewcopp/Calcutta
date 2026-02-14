import React, { useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useParams, useSearchParams } from 'react-router-dom';

import { Alert } from '../../components/ui/Alert';
import { Breadcrumb } from '../../components/ui/Breadcrumb';
import { LoadingState } from '../../components/ui/LoadingState';
import { TabsNav } from '../../components/TabsNav';
import { labService, EntryDetail, ListEvaluationsResponse } from '../../services/labService';
import { cn } from '../../lib/cn';

import { PredictionsTab, PredSortKey } from './EntryDetail/PredictionsTab';
import { EntryTab, BidSortKey } from './EntryDetail/EntryTab';
import { EvaluationsTab } from './EntryDetail/EvaluationsTab';

type EntryDetailTabId = 'predictions' | 'entry' | 'evaluations';
type SortDir = 'asc' | 'desc';

const TABS = [
  { id: 'predictions' as const, label: 'Market Predictions' },
  { id: 'entry' as const, label: 'Optimized Entry' },
  { id: 'evaluations' as const, label: 'Evaluations' },
] as const;

export function EntryDetailPage() {
  // Support both URL patterns:
  // - /lab/models/:modelName/calcutta/:calcuttaId (new)
  // - /lab/entries/:entryId (legacy)
  const { entryId, modelName, calcuttaId } = useParams<{
    entryId?: string;
    modelName?: string;
    calcuttaId?: string;
  }>();
  const [searchParams, setSearchParams] = useSearchParams();

  // Tab state with URL sync
  const [activeTab, setActiveTab] = useState<EntryDetailTabId>(() => {
    const tab = searchParams.get('tab');
    if (tab === 'predictions' || tab === 'entry' || tab === 'evaluations') return tab;
    return 'predictions';
  });

  // Sync tab changes to URL
  const handleTabChange = (tab: EntryDetailTabId) => {
    setActiveTab(tab);
    setSearchParams({ tab }, { replace: true });
  };

  // Sort state for predictions tab
  const [predSortKey, setPredSortKey] = useState<PredSortKey>('edge');
  const [predSortDir, setPredSortDir] = useState<SortDir>('desc');

  // Sort state for entry tab
  const [bidSortKey, setBidSortKey] = useState<BidSortKey>('adj_roi');
  const [bidSortDir, setBidSortDir] = useState<SortDir>('desc');
  const [showOnlyInvested, setShowOnlyInvested] = useState(false);

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
    queryFn: () => (loadedEntryId ? labService.listEvaluations({ entry_id: loadedEntryId, limit: 1 }) : Promise.resolve(null)),
    enabled: Boolean(loadedEntryId),
  });

  const entry = entryQuery.data;
  const evaluation = evaluationsQuery.data?.items[0] ?? null;
  const predictions = entry?.predictions ?? [];
  const bids = entry?.bids ?? [];

  // Sort predictions based on current sort settings
  const sortedPredictions = useMemo(() => {
    return [...predictions].sort((a, b) => {
      let cmp = 0;
      switch (predSortKey) {
        case 'seed':
          cmp = a.seed - b.seed;
          break;
        case 'team':
          cmp = a.school_name.localeCompare(b.school_name);
          break;
        case 'rational':
          cmp = b.naive_points - a.naive_points;
          break;
        case 'predicted':
          cmp = b.predicted_bid_points - a.predicted_bid_points;
          break;
        case 'edge':
          cmp = b.edge_percent - a.edge_percent;
          break;
      }
      return predSortDir === 'asc' ? -cmp : cmp;
    });
  }, [predictions, predSortKey, predSortDir]);

  const handleBidSort = (key: BidSortKey) => {
    if (bidSortKey === key) {
      setBidSortDir(bidSortDir === 'asc' ? 'desc' : 'asc');
    } else {
      setBidSortKey(key);
      setBidSortDir('desc');
    }
  };

  const handlePredSort = (key: PredSortKey) => {
    if (predSortKey === key) {
      setPredSortDir(predSortDir === 'asc' ? 'desc' : 'asc');
    } else {
      setPredSortKey(key);
      setPredSortDir('desc');
    }
  };

  if (entryQuery.isLoading) {
    return (
      <div className="container mx-auto px-4 py-4">
        <LoadingState label="Loading entry..." />
      </div>
    );
  }

  if (entryQuery.isError || !entry) {
    return (
      <div className="container mx-auto px-4 py-4">
        <Alert variant="error">Failed to load entry.</Alert>
      </div>
    );
  }

  // Compute stage completion status
  const hasPredictions = entry.has_predictions && predictions.length > 0;
  const hasBids = bids.length > 0;
  const hasEvaluations = evaluation !== null;

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
        <h1 className="text-xl font-bold text-gray-900">Entry Detail</h1>
        <span className="text-gray-500">
          {entry.model_name} ({entry.model_kind}) → {entry.calcutta_name}
        </span>
      </div>

      {/* Interactive pipeline stage indicator */}
      <div className="flex items-center gap-2 mb-4 text-sm">
        <span className="px-2 py-1 rounded bg-green-100 text-green-800">✓ Registered</span>
        <span className="text-gray-400">→</span>
        <button
          type="button"
          onClick={() => handleTabChange('predictions')}
          disabled={!hasPredictions}
          className={cn(
            'px-2 py-1 rounded transition-colors',
            hasPredictions ? 'bg-green-100 text-green-800 hover:bg-green-200' : 'bg-gray-100 text-gray-500',
            activeTab === 'predictions' && hasPredictions && 'ring-2 ring-green-500 ring-offset-1'
          )}
        >
          {hasPredictions ? '✓' : '○'} Predicted
        </button>
        <span className="text-gray-400">→</span>
        <button
          type="button"
          onClick={() => handleTabChange('entry')}
          disabled={!hasBids}
          className={cn(
            'px-2 py-1 rounded transition-colors',
            hasBids ? 'bg-green-100 text-green-800 hover:bg-green-200' : 'bg-gray-100 text-gray-500',
            activeTab === 'entry' && hasBids && 'ring-2 ring-green-500 ring-offset-1'
          )}
        >
          {hasBids ? '✓' : '○'} Optimized
        </button>
        <span className="text-gray-400">→</span>
        <button
          type="button"
          onClick={() => handleTabChange('evaluations')}
          disabled={!hasEvaluations}
          className={cn(
            'px-2 py-1 rounded transition-colors',
            hasEvaluations ? 'bg-green-100 text-green-800 hover:bg-green-200' : 'bg-gray-100 text-gray-500',
            activeTab === 'evaluations' && hasEvaluations && 'ring-2 ring-green-500 ring-offset-1'
          )}
        >
          {hasEvaluations ? '✓' : '○'} Evaluated
        </button>
      </div>

      {/* Tabs navigation */}
      <TabsNav tabs={TABS} activeTab={activeTab} onTabChange={handleTabChange} />

      {/* Tab content */}
      {activeTab === 'predictions' && (
        hasPredictions ? (
          <PredictionsTab
            predictions={sortedPredictions}
            sortKey={predSortKey}
            sortDir={predSortDir}
            onSort={handlePredSort}
            optimizerParams={entry.optimizer_params_json}
          />
        ) : (
          <Alert variant="info">No predictions available for this entry.</Alert>
        )
      )}

      {activeTab === 'entry' && (
        <EntryTab
          bids={bids}
          predictions={predictions}
          sortKey={bidSortKey}
          sortDir={bidSortDir}
          onSort={handleBidSort}
          showOnlyInvested={showOnlyInvested}
          onShowOnlyInvestedChange={setShowOnlyInvested}
          optimizerKind={entry.optimizer_kind}
          optimizerParams={entry.optimizer_params_json}
        />
      )}

      {activeTab === 'evaluations' && (
        <EvaluationsTab
          evaluation={evaluation}
          isLoading={evaluationsQuery.isLoading}
        />
      )}
    </div>
  );
}

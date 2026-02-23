import React, { useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useParams, useSearchParams } from 'react-router-dom';

import { Alert } from '../../components/ui/Alert';
import { ErrorState } from '../../components/ui/ErrorState';
import { Breadcrumb } from '../../components/ui/Breadcrumb';
import { LoadingState } from '../../components/ui/LoadingState';
import { PageContainer } from '../../components/ui/Page';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '../../components/ui/Tabs';
import { labService } from '../../services/labService';
import type { EntryDetail, ListEvaluationsResponse, SortDir } from '../../schemas/lab';
import { cn } from '../../lib/cn';
import { queryKeys } from '../../queryKeys';

import { PredictionsTab, PredSortKey } from './EntryDetail/PredictionsTab';
import { EntryTab, BidSortKey } from './EntryDetail/EntryTab';
import { EvaluationsTab } from './EntryDetail/EvaluationsTab';

type EntryDetailTabId = 'predictions' | 'entry' | 'evaluations';

export function EntryDetailPage() {
  const { modelName, calcuttaId } = useParams<{
    modelName: string;
    calcuttaId: string;
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

  const entryQuery = useQuery<EntryDetail | null>({
    queryKey: queryKeys.lab.entries.byModelAndCalcutta(modelName, calcuttaId),
    queryFn: () => labService.getEntryByModelAndCalcutta(modelName!, calcuttaId!),
    enabled: Boolean(modelName && calcuttaId),
  });

  // For evaluations, we need the entry ID from the loaded entry
  const loadedEntryId = entryQuery.data?.id;
  const evaluationsQuery = useQuery<ListEvaluationsResponse | null>({
    queryKey: queryKeys.lab.evaluations.byEntry(loadedEntryId),
    queryFn: () =>
      loadedEntryId ? labService.listEvaluations({ entryId: loadedEntryId, limit: 1 }) : Promise.resolve(null),
    enabled: Boolean(loadedEntryId),
  });

  const entry = entryQuery.data;
  const evaluation = evaluationsQuery.data?.items[0] ?? null;
  const predictions = useMemo(() => entry?.predictions ?? [], [entry?.predictions]);
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
          cmp = a.schoolName.localeCompare(b.schoolName);
          break;
        case 'rational':
          cmp = b.naivePoints - a.naivePoints;
          break;
        case 'predicted':
          cmp = b.predictedBidPoints - a.predictedBidPoints;
          break;
        case 'edge':
          cmp = b.edgePercent - a.edgePercent;
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
      <PageContainer>
        <LoadingState label="Loading entry..." />
      </PageContainer>
    );
  }

  if (entryQuery.isError || !entry) {
    return (
      <PageContainer>
        <ErrorState error={entryQuery.error ?? 'Failed to load entry.'} onRetry={() => entryQuery.refetch()} />
      </PageContainer>
    );
  }

  // Compute stage completion status
  const hasPredictions = entry.hasPredictions && predictions.length > 0;
  const hasBids = bids.length > 0;
  const hasEvaluations = evaluation !== null;

  return (
    <PageContainer>
      <Breadcrumb
        items={[
          { label: 'Lab', href: '/lab' },
          { label: entry.modelName, href: `/lab/models/${entry.investmentModelId}` },
          { label: entry.calcuttaName },
        ]}
      />

      {/* Compact header */}
      <div className="flex items-baseline gap-3 mb-4">
        <h1 className="text-xl font-bold text-foreground">Entry Detail</h1>
        <span className="text-muted-foreground">
          {entry.modelName} ({entry.modelKind}) → {entry.calcuttaName}
        </span>
      </div>

      {/* Interactive pipeline stage indicator */}
      <div className="flex items-center gap-2 mb-4 text-sm">
        <span className="px-2 py-1 rounded bg-success/10 text-success">✓ Registered</span>
        <span className="text-muted-foreground/60">→</span>
        <button
          type="button"
          onClick={() => handleTabChange('predictions')}
          disabled={!hasPredictions}
          className={cn(
            'px-2 py-1 rounded transition-colors',
            hasPredictions ? 'bg-success/10 text-success hover:bg-green-200' : 'bg-muted text-muted-foreground',
            activeTab === 'predictions' && hasPredictions && 'ring-2 ring-green-500 ring-offset-1',
          )}
        >
          {hasPredictions ? '✓' : '○'} Predicted
        </button>
        <span className="text-muted-foreground/60">→</span>
        <button
          type="button"
          onClick={() => handleTabChange('entry')}
          disabled={!hasBids}
          className={cn(
            'px-2 py-1 rounded transition-colors',
            hasBids ? 'bg-success/10 text-success hover:bg-green-200' : 'bg-muted text-muted-foreground',
            activeTab === 'entry' && hasBids && 'ring-2 ring-green-500 ring-offset-1',
          )}
        >
          {hasBids ? '✓' : '○'} Optimized
        </button>
        <span className="text-muted-foreground/60">→</span>
        <button
          type="button"
          onClick={() => handleTabChange('evaluations')}
          disabled={!hasEvaluations}
          className={cn(
            'px-2 py-1 rounded transition-colors',
            hasEvaluations ? 'bg-success/10 text-success hover:bg-green-200' : 'bg-muted text-muted-foreground',
            activeTab === 'evaluations' && hasEvaluations && 'ring-2 ring-green-500 ring-offset-1',
          )}
        >
          {hasEvaluations ? '✓' : '○'} Evaluated
        </button>
      </div>

      <Tabs value={activeTab} onValueChange={(v) => handleTabChange(v as EntryDetailTabId)}>
        <TabsList>
          <TabsTrigger value="predictions">Market Predictions</TabsTrigger>
          <TabsTrigger value="entry">Optimized Entry</TabsTrigger>
          <TabsTrigger value="evaluations">Evaluations</TabsTrigger>
        </TabsList>

        <TabsContent value="predictions">
          {hasPredictions ? (
            <PredictionsTab
              predictions={sortedPredictions}
              sortKey={predSortKey}
              sortDir={predSortDir}
              onSort={handlePredSort}
              optimizerParams={entry.optimizerParamsJson}
            />
          ) : (
            <Alert variant="info">No predictions available for this entry.</Alert>
          )}
        </TabsContent>

        <TabsContent value="entry">
          <EntryTab
            bids={bids}
            predictions={predictions}
            sortKey={bidSortKey}
            sortDir={bidSortDir}
            onSort={handleBidSort}
            showOnlyInvested={showOnlyInvested}
            onShowOnlyInvestedChange={setShowOnlyInvested}
            optimizerKind={entry.optimizerKind}
            optimizerParams={entry.optimizerParamsJson}
          />
        </TabsContent>

        <TabsContent value="evaluations">
          <EvaluationsTab
            evaluation={evaluation}
            isLoading={evaluationsQuery.isLoading}
            modelName={entry.modelName}
            calcuttaId={entry.calcuttaId}
          />
        </TabsContent>
      </Tabs>
    </PageContainer>
  );
}

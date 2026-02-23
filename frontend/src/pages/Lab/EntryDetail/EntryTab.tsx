import { useMemo } from 'react';

import { Alert } from '../../../components/ui/Alert';
import { Card } from '../../../components/ui/Card';
import { cn } from '../../../lib/cn';
import type { EnrichedBid, EnrichedPrediction, SortDir } from '../../../types/lab';
import { getRoiColor, formatRoi } from '../../../utils/labFormatters';
import { OptimizerConfigCard } from './OptimizerConfigCard';
import { EntryStatsCards } from './EntryStatsCards';
import { SeedAllocationChart } from './SeedAllocationChart';

type BidSortKey = 'seed' | 'team' | 'pred_perf' | 'pred_inv' | 'our_inv' | 'pred_roi' | 'adj_roi';

interface EntryTabProps {
  bids: EnrichedBid[];
  predictions: EnrichedPrediction[];
  sortKey: BidSortKey;
  sortDir: SortDir;
  onSort: (key: BidSortKey) => void;
  showOnlyInvested: boolean;
  onShowOnlyInvestedChange: (value: boolean) => void;
  optimizerKind?: string;
  optimizerParams?: Record<string, unknown>;
}

// Combined row for display - joins bids with prediction data
interface EntryRow {
  teamId: string;
  schoolName: string;
  seed: number;
  region: string;
  predPerformance: number; // naivePoints (budget-normalized expected value, same as Rational)
  predInvestment: number; // predictedBidPoints from predictions
  ourInvestment: number; // bidPoints from bids
  predRoi: number; // predPerformance / predInvestment
  adjRoi: number; // predPerformance / (predInvestment + ourInvestment)
}

export function EntryTab({
  bids,
  predictions,
  sortKey,
  sortDir,
  onSort,
  showOnlyInvested,
  onShowOnlyInvestedChange,
  optimizerKind,
  optimizerParams,
}: EntryTabProps) {
  // Join bids with predictions by teamId
  const rows = useMemo((): EntryRow[] => {
    const predByTeam = new Map(predictions.map((p) => [p.teamId, p]));

    return bids.map((bid) => {
      const pred = predByTeam.get(bid.teamId);
      // Use expectedPoints from predictions as predPerformance (actual expected tournament points)
      const predPerf = pred?.expectedPoints ?? 0;
      const predInv = pred?.predictedBidPoints ?? 0;
      const ourInv = bid.bidPoints;

      const predRoi = predInv > 0 ? predPerf / predInv : 0;
      const adjRoi = predInv + ourInv > 0 ? predPerf / (predInv + ourInv) : 0;

      return {
        teamId: bid.teamId,
        schoolName: bid.schoolName,
        seed: bid.seed,
        region: bid.region,
        predPerformance: predPerf,
        predInvestment: predInv,
        ourInvestment: ourInv,
        predRoi: predRoi,
        adjRoi: adjRoi,
      };
    });
  }, [bids, predictions]);

  // Sort rows
  const sortedRows = useMemo(() => {
    return [...rows].sort((a, b) => {
      let cmp = 0;
      switch (sortKey) {
        case 'seed':
          cmp = a.seed - b.seed;
          break;
        case 'team':
          cmp = a.schoolName.localeCompare(b.schoolName);
          break;
        case 'pred_perf':
          cmp = b.predPerformance - a.predPerformance;
          break;
        case 'pred_inv':
          cmp = b.predInvestment - a.predInvestment;
          break;
        case 'our_inv':
          cmp = b.ourInvestment - a.ourInvestment;
          break;
        case 'pred_roi':
          cmp = b.predRoi - a.predRoi;
          break;
        case 'adj_roi':
          cmp = b.adjRoi - a.adjRoi;
          break;
      }
      return sortDir === 'asc' ? -cmp : cmp;
    });
  }, [rows, sortKey, sortDir]);

  // Filter if needed
  const displayRows = showOnlyInvested ? sortedRows.filter((r) => r.ourInvestment > 0) : sortedRows;

  const SortHeader = ({ label, sortKeyValue }: { label: string; sortKeyValue: BidSortKey }) => (
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

  // Summary stats
  const investedRows = rows.filter((r) => r.ourInvestment > 0);
  const totalOurInvestment = rows.reduce((sum, r) => sum + r.ourInvestment, 0);

  // Weighted average ROI (by our investment)
  const weightedPredRoi =
    investedRows.length > 0
      ? investedRows.reduce((sum, r) => sum + r.predRoi * r.ourInvestment, 0) / totalOurInvestment
      : 0;
  const weightedAdjRoi =
    investedRows.length > 0
      ? investedRows.reduce((sum, r) => sum + r.adjRoi * r.ourInvestment, 0) / totalOurInvestment
      : 0;

  // Top ROI teams we invested in
  const topRoiInvested = [...investedRows].sort((a, b) => b.adjRoi - a.adjRoi).slice(0, 3);

  return (
    <div className="space-y-4">
      <OptimizerConfigCard optimizerKind={optimizerKind} optimizerParams={optimizerParams} />

      <EntryStatsCards
        investedCount={investedRows.length}
        totalCount={rows.length}
        totalOurInvestment={totalOurInvestment}
        weightedPredRoi={weightedPredRoi}
        weightedAdjRoi={weightedAdjRoi}
        topRoiInvested={topRoiInvested}
      />

      <SeedAllocationChart investedRows={investedRows} totalOurInvestment={totalOurInvestment} />

      <Card>
        <div className="flex items-center justify-between mb-3">
          <h2 className="text-lg font-semibold">Optimized Entry</h2>
          <label className="flex items-center gap-2 text-sm text-muted-foreground">
            <input
              type="checkbox"
              checked={showOnlyInvested}
              onChange={(e) => onShowOnlyInvestedChange(e.target.checked)}
              className="rounded border-border"
            />
            Show only invested ({investedRows.length} of {rows.length})
          </label>
        </div>
        <p className="text-sm text-muted-foreground mb-3">
          Our optimized bid allocation based on predicted performance and market behavior.
          <strong className="ml-2">Pred Perf</strong> = expected tournament points (same as Rational).
          <strong className="ml-2">Pred Inv</strong> = predicted market bid.
          <strong className="ml-2">Our Inv</strong> = our bid.
          <strong className="ml-2">Pred ROI</strong> = perf / pred_inv.
          <strong className="ml-2">Adj ROI</strong> = perf / (pred_inv + our_inv).
        </p>
        {bids.length === 0 ? (
          <Alert variant="info">No bids in this entry.</Alert>
        ) : (
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
                    <SortHeader label="Pred Perf" sortKeyValue="pred_perf" />
                  </th>
                  <th className="px-3 py-2 text-right">
                    <SortHeader label="Pred Inv" sortKeyValue="pred_inv" />
                  </th>
                  <th className="px-3 py-2 text-right">
                    <SortHeader label="Our Inv" sortKeyValue="our_inv" />
                  </th>
                  <th className="px-3 py-2 text-right">
                    <SortHeader label="Pred ROI" sortKeyValue="pred_roi" />
                  </th>
                  <th className="px-3 py-2 text-right">
                    <SortHeader label="Adj ROI" sortKeyValue="adj_roi" />
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {displayRows.map((row) => (
                  <tr key={row.teamId} className={cn('hover:bg-accent', row.ourInvestment > 0 ? 'bg-primary/10' : '')}>
                    <td className="px-3 py-2 text-sm font-medium text-foreground">{row.schoolName}</td>
                    <td className="px-3 py-2 text-sm text-foreground text-center">{row.seed}</td>
                    <td className="px-3 py-2 text-sm text-muted-foreground">{row.region}</td>
                    <td className="px-3 py-2 text-sm text-foreground text-right tabular-nums">
                      {row.predPerformance.toFixed(0)}
                    </td>
                    <td className="px-3 py-2 text-sm text-foreground text-right tabular-nums">{row.predInvestment}</td>
                    <td className="px-3 py-2 text-sm text-right font-medium tabular-nums">
                      {row.ourInvestment > 0 ? (
                        <span className="text-primary">{row.ourInvestment}</span>
                      ) : (
                        <span className="text-muted-foreground/60">0</span>
                      )}
                    </td>
                    <td className={cn('px-3 py-2 text-sm text-right tabular-nums', getRoiColor(row.predRoi))}>
                      {formatRoi(row.predRoi)}
                    </td>
                    <td
                      className={cn('px-3 py-2 text-sm text-right font-medium tabular-nums', getRoiColor(row.adjRoi))}
                    >
                      {formatRoi(row.adjRoi)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </Card>
    </div>
  );
}

export type { BidSortKey };

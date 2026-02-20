import React, { useMemo } from 'react';
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer, Cell } from 'recharts';

import { Alert } from '../../../components/ui/Alert';
import { Card } from '../../../components/ui/Card';
import { cn } from '../../../lib/cn';
import type { EnrichedBid, EnrichedPrediction, SortDir } from '../../../types/lab';

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
  team_id: string;
  school_name: string;
  seed: number;
  region: string;
  pred_performance: number; // naive_points (budget-normalized expected value, same as Rational)
  pred_investment: number;  // predicted_bid_points from predictions
  our_investment: number;   // bid_points from bids
  pred_roi: number;         // pred_performance / pred_investment
  adj_roi: number;          // pred_performance / (pred_investment + our_investment)
}

export function getRoiColor(roi: number): string {
  if (roi >= 2.0) return 'text-green-700 font-bold';
  if (roi >= 1.5) return 'text-green-600';
  if (roi >= 1.0) return 'text-gray-900';
  if (roi >= 0.5) return 'text-yellow-600';
  return 'text-red-600';
}

export function formatRoi(roi: number): string {
  if (!isFinite(roi) || isNaN(roi)) return '—';
  return `${roi.toFixed(2)}x`;
}

export function EntryTab({ bids, predictions, sortKey, sortDir, onSort, showOnlyInvested, onShowOnlyInvestedChange, optimizerKind, optimizerParams }: EntryTabProps) {
  // Join bids with predictions by team_id
  const rows = useMemo((): EntryRow[] => {
    const predByTeam = new Map(predictions.map(p => [p.team_id, p]));

    return bids.map(bid => {
      const pred = predByTeam.get(bid.team_id);
      // Use expected_points from predictions as pred_performance (actual expected tournament points)
      const predPerf = pred?.expected_points ?? 0;
      const predInv = pred?.predicted_bid_points ?? 0;
      const ourInv = bid.bid_points;

      const predRoi = predInv > 0 ? predPerf / predInv : 0;
      const adjRoi = (predInv + ourInv) > 0 ? predPerf / (predInv + ourInv) : 0;

      return {
        team_id: bid.team_id,
        school_name: bid.school_name,
        seed: bid.seed,
        region: bid.region,
        pred_performance: predPerf,
        pred_investment: predInv,
        our_investment: ourInv,
        pred_roi: predRoi,
        adj_roi: adjRoi,
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
          cmp = a.school_name.localeCompare(b.school_name);
          break;
        case 'pred_perf':
          cmp = b.pred_performance - a.pred_performance;
          break;
        case 'pred_inv':
          cmp = b.pred_investment - a.pred_investment;
          break;
        case 'our_inv':
          cmp = b.our_investment - a.our_investment;
          break;
        case 'pred_roi':
          cmp = b.pred_roi - a.pred_roi;
          break;
        case 'adj_roi':
          cmp = b.adj_roi - a.adj_roi;
          break;
      }
      return sortDir === 'asc' ? -cmp : cmp;
    });
  }, [rows, sortKey, sortDir]);

  // Filter if needed
  const displayRows = showOnlyInvested ? sortedRows.filter(r => r.our_investment > 0) : sortedRows;

  const SortHeader = ({ label, sortKeyValue }: { label: string; sortKeyValue: BidSortKey }) => (
    <button
      type="button"
      onClick={() => onSort(sortKeyValue)}
      className={cn(
        'flex items-center gap-1 text-xs font-medium uppercase tracking-wider',
        sortKey === sortKeyValue ? 'text-blue-600' : 'text-gray-500 hover:text-gray-700'
      )}
    >
      {label}
      {sortKey === sortKeyValue ? (sortDir === 'desc' ? ' ▼' : ' ▲') : ''}
    </button>
  );

  // Summary stats
  const investedRows = rows.filter(r => r.our_investment > 0);
  const totalOurInvestment = rows.reduce((sum, r) => sum + r.our_investment, 0);

  // Weighted average ROI (by our investment)
  const weightedPredRoi = investedRows.length > 0
    ? investedRows.reduce((sum, r) => sum + r.pred_roi * r.our_investment, 0) / totalOurInvestment
    : 0;
  const weightedAdjRoi = investedRows.length > 0
    ? investedRows.reduce((sum, r) => sum + r.adj_roi * r.our_investment, 0) / totalOurInvestment
    : 0;

  // Top ROI teams we invested in
  const topRoiInvested = [...investedRows].sort((a, b) => b.adj_roi - a.adj_roi).slice(0, 3);

  // Seed allocation data for chart
  const seedAllocationData = useMemo(() => {
    const groups = [
      { name: '1-4', seeds: [1, 2, 3, 4], color: '#1d4ed8' },
      { name: '5-8', seeds: [5, 6, 7, 8], color: '#3b82f6' },
      { name: '9-12', seeds: [9, 10, 11, 12], color: '#93c5fd' },
      { name: '13-16', seeds: [13, 14, 15, 16], color: '#dbeafe' },
    ];
    return groups.map(({ name, seeds, color }) => {
      const amount = investedRows
        .filter((r) => seeds.includes(r.seed))
        .reduce((sum, r) => sum + r.our_investment, 0);
      return { name, amount, color, pct: totalOurInvestment > 0 ? (amount / totalOurInvestment * 100) : 0 };
    });
  }, [investedRows, totalOurInvestment]);

  // Extract optimizer constraints
  const budgetPoints = optimizerParams?.budget_points as number | undefined;
  const maxPerTeam = optimizerParams?.max_per_team as number | undefined;
  const minTeams = optimizerParams?.min_teams as number | undefined;
  const maxTeams = optimizerParams?.max_teams as number | undefined;
  const minBid = optimizerParams?.min_bid as number | undefined;
  const edgeMultiplier = optimizerParams?.edge_multiplier as number | undefined;

  return (
    <div className="space-y-4">
      {/* Optimizer configuration */}
      {(optimizerKind || budgetPoints != null) && (
        <div className="bg-gray-50 rounded-lg border border-gray-200 p-3">
          <div className="text-xs text-gray-500 uppercase mb-2">Optimizer Configuration</div>
          <div className="flex flex-wrap gap-x-6 gap-y-1 text-sm">
            {optimizerKind && (
              <div>
                <span className="text-gray-500">Optimizer:</span>{' '}
                <span className="font-medium">{optimizerKind}</span>
              </div>
            )}
            {budgetPoints != null && (
              <div>
                <span className="text-gray-500">Budget:</span>{' '}
                <span className="font-medium">{budgetPoints} pts</span>
              </div>
            )}
            {maxPerTeam != null && (
              <div>
                <span className="text-gray-500">Max Per Team:</span>{' '}
                <span className="font-medium">{maxPerTeam} pts</span>
              </div>
            )}
            {minTeams != null && (
              <div>
                <span className="text-gray-500">Min Teams:</span>{' '}
                <span className="font-medium">{minTeams}</span>
              </div>
            )}
            {maxTeams != null && (
              <div>
                <span className="text-gray-500">Max Teams:</span>{' '}
                <span className="font-medium">{maxTeams}</span>
              </div>
            )}
            {minBid != null && (
              <div>
                <span className="text-gray-500">Min Bid:</span>{' '}
                <span className="font-medium">{minBid} pts</span>
              </div>
            )}
            {edgeMultiplier != null && (
              <div>
                <span className="text-gray-500">Edge Multiplier:</span>{' '}
                <span className="font-medium">{edgeMultiplier}x</span>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Summary stats */}
      <div className="grid grid-cols-4 gap-4">
        <div className="bg-white rounded-lg border border-gray-200 p-3">
          <div className="text-xs text-gray-500 uppercase">Teams Invested</div>
          <div className="text-lg font-semibold">{investedRows.length} of {rows.length}</div>
        </div>
        <div className="bg-white rounded-lg border border-gray-200 p-3">
          <div className="text-xs text-gray-500 uppercase">Our Investment</div>
          <div className="text-lg font-semibold">{totalOurInvestment.toLocaleString()} pts</div>
        </div>
        <div className="bg-white rounded-lg border border-gray-200 p-3">
          <div className="text-xs text-gray-500 uppercase">Wtd Pred ROI</div>
          <div className={cn('text-lg font-semibold', getRoiColor(weightedPredRoi))}>
            {formatRoi(weightedPredRoi)}
          </div>
        </div>
        <div className="bg-white rounded-lg border border-gray-200 p-3">
          <div className="text-xs text-gray-500 uppercase">Wtd Adj ROI</div>
          <div className={cn('text-lg font-semibold', getRoiColor(weightedAdjRoi))}>
            {formatRoi(weightedAdjRoi)}
          </div>
        </div>
      </div>

      {/* Top ROI investments */}
      {topRoiInvested.length > 0 && (
        <div className="bg-blue-50 rounded-lg border border-blue-200 p-3">
          <div className="text-xs text-blue-700 uppercase mb-2">Top Adj ROI Investments</div>
          {topRoiInvested.map(r => (
            <div key={r.team_id} className="text-sm text-blue-800">
              {r.school_name} ({r.seed}): {r.our_investment} pts → {formatRoi(r.adj_roi)} adj ROI
            </div>
          ))}
        </div>
      )}

      {investedRows.length > 0 && totalOurInvestment > 0 && (
        <Card>
          <h2 className="text-lg font-semibold mb-3">Seed Allocation</h2>
          <p className="text-sm text-gray-500 mb-3">Budget distribution across seed groups.</p>
          <ResponsiveContainer width="100%" height={160}>
            <BarChart data={seedAllocationData} layout="vertical">
              <XAxis type="number" tickFormatter={(v: number) => `${v} pts`} fontSize={12} />
              <YAxis type="category" dataKey="name" width={50} fontSize={12} />
              <Tooltip
                formatter={(value: number, _name: string, props: { payload?: { pct?: number } }) =>
                  [`${value} pts (${props.payload?.pct?.toFixed(0) ?? 0}%)`, 'Investment']
                }
              />
              <Bar dataKey="amount" radius={[0, 4, 4, 0]}>
                {seedAllocationData.map((entry) => (
                  <Cell key={entry.name} fill={entry.color} />
                ))}
              </Bar>
            </BarChart>
          </ResponsiveContainer>
        </Card>
      )}

      <Card>
        <div className="flex items-center justify-between mb-3">
          <h2 className="text-lg font-semibold">Optimized Entry</h2>
          <label className="flex items-center gap-2 text-sm text-gray-600">
            <input
              type="checkbox"
              checked={showOnlyInvested}
              onChange={(e) => onShowOnlyInvestedChange(e.target.checked)}
              className="rounded border-gray-300"
            />
            Show only invested ({investedRows.length} of {rows.length})
          </label>
        </div>
        <p className="text-sm text-gray-600 mb-3">
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
                  <tr
                    key={row.team_id}
                    className={cn(
                      'hover:bg-gray-50',
                      row.our_investment > 0 ? 'bg-blue-50' : ''
                    )}
                  >
                    <td className="px-3 py-2 text-sm font-medium text-gray-900">{row.school_name}</td>
                    <td className="px-3 py-2 text-sm text-gray-700 text-center">{row.seed}</td>
                    <td className="px-3 py-2 text-sm text-gray-500">{row.region}</td>
                    <td className="px-3 py-2 text-sm text-gray-900 text-right tabular-nums">
                      {row.pred_performance.toFixed(0)}
                    </td>
                    <td className="px-3 py-2 text-sm text-gray-900 text-right tabular-nums">
                      {row.pred_investment}
                    </td>
                    <td className="px-3 py-2 text-sm text-right font-medium tabular-nums">
                      {row.our_investment > 0 ? (
                        <span className="text-blue-700">{row.our_investment}</span>
                      ) : (
                        <span className="text-gray-400">0</span>
                      )}
                    </td>
                    <td className={cn('px-3 py-2 text-sm text-right tabular-nums', getRoiColor(row.pred_roi))}>
                      {formatRoi(row.pred_roi)}
                    </td>
                    <td className={cn('px-3 py-2 text-sm text-right font-medium tabular-nums', getRoiColor(row.adj_roi))}>
                      {formatRoi(row.adj_roi)}
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

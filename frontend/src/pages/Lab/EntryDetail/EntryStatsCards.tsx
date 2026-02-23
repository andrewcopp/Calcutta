import { cn } from '../../../lib/cn';
import { getRoiColor, formatRoi } from '../../../utils/labFormatters';

interface EntryStatsCardsProps {
  investedCount: number;
  totalCount: number;
  totalOurInvestment: number;
  weightedPredRoi: number;
  weightedAdjRoi: number;
  topRoiInvested: Array<{
    teamId: string;
    schoolName: string;
    seed: number;
    ourInvestment: number;
    adjRoi: number;
  }>;
}

export function EntryStatsCards({
  investedCount,
  totalCount,
  totalOurInvestment,
  weightedPredRoi,
  weightedAdjRoi,
  topRoiInvested,
}: EntryStatsCardsProps) {
  return (
    <>
      <div className="grid grid-cols-4 gap-4">
        <div className="bg-card rounded-lg border border-border p-3">
          <div className="text-xs text-muted-foreground uppercase">Teams Invested</div>
          <div className="text-lg font-semibold">
            {investedCount} of {totalCount}
          </div>
        </div>
        <div className="bg-card rounded-lg border border-border p-3">
          <div className="text-xs text-muted-foreground uppercase">Our Investment</div>
          <div className="text-lg font-semibold">{totalOurInvestment.toLocaleString()} credits</div>
        </div>
        <div className="bg-card rounded-lg border border-border p-3">
          <div className="text-xs text-muted-foreground uppercase">Wtd Pred ROI</div>
          <div className={cn('text-lg font-semibold', getRoiColor(weightedPredRoi))}>{formatRoi(weightedPredRoi)}</div>
        </div>
        <div className="bg-card rounded-lg border border-border p-3">
          <div className="text-xs text-muted-foreground uppercase">Wtd Adj ROI</div>
          <div className={cn('text-lg font-semibold', getRoiColor(weightedAdjRoi))}>{formatRoi(weightedAdjRoi)}</div>
        </div>
      </div>

      {topRoiInvested.length > 0 && (
        <div className="bg-primary/10 rounded-lg border border-blue-200 p-3">
          <div className="text-xs text-primary uppercase mb-2">Top Adj ROI Investments</div>
          {topRoiInvested.map((r) => (
            <div key={r.teamId} className="text-sm text-primary">
              {r.schoolName} ({r.seed}): {r.ourInvestment} credits {'\u2192'} {formatRoi(r.adjRoi)} adj ROI
            </div>
          ))}
        </div>
      )}
    </>
  );
}

import type { BumpPoint } from '@nivo/bump';
import type { RaceDatum } from '../../utils/buildBumpChartData';

interface RacePointTooltipProps {
  point: BumpPoint<RaceDatum, Record<string, unknown>>;
}

function formatRankDelta(delta: number): string {
  if (delta > 0) return `\u2191 ${delta}`;
  if (delta < 0) return `\u2193 ${Math.abs(delta)}`;
  return '';
}

function formatPointsDelta(delta: number): string {
  if (delta > 0) return `(+${delta.toFixed(2)})`;
  if (delta < 0) return `(${delta.toFixed(2)})`;
  return '';
}

export function RacePointTooltip({ point }: RacePointTooltipProps) {
  const { data, color } = point;
  const name = point.serie.id;
  const rankDeltaStr = formatRankDelta(data.rankDelta);
  const pointsDeltaStr = formatPointsDelta(data.pointsDelta);

  return (
    <div className="bg-white rounded-lg shadow-md px-3 py-2 text-sm min-w-[160px]">
      <div className="flex items-center gap-2 font-medium">
        <span className="inline-block w-3 h-3 rounded-full flex-shrink-0" style={{ backgroundColor: color }} />
        <span className="truncate">{name}</span>
        <span className="text-muted-foreground ml-auto whitespace-nowrap">
          #{data.y}
          {rankDeltaStr && <span className="ml-1 text-xs">{rankDeltaStr}</span>}
        </span>
      </div>
      <div className="mt-1 text-muted-foreground flex justify-between gap-3">
        <span>Points: {data.totalPoints.toFixed(2)}</span>
        {pointsDeltaStr && <span className="text-xs">{pointsDeltaStr}</span>}
      </div>
      {data.projectedEv != null && (
        <div className="text-muted-foreground">EV: {data.projectedEv.toFixed(2)}</div>
      )}
    </div>
  );
}

import { useMemo, useState } from 'react';
import { ResponsiveBump } from '@nivo/bump';
import type { CalcuttaDashboard, CalcuttaEntry } from '../../schemas/calcutta';
import { buildBumpChartData } from '../../utils/buildBumpChartData';
import { getEntryColorById } from '../../utils/entryColors';
import { Select } from '../../components/ui/Select';

type SortMode = 'actual' | 'projected' | 'favorites';

interface RaceTabProps {
  entries: CalcuttaEntry[];
  dashboard: CalcuttaDashboard;
}

export function RaceTab({ entries, dashboard }: RaceTabProps) {
  const [sortMode, setSortMode] = useState<SortMode>('actual');

  const hasProjections = dashboard.roundStandings.some((g) => g.entries.some((e) => e.projectedEv != null));
  const hasFavorites = dashboard.roundStandings.some((g) => g.entries.some((e) => e.projectedFavorites != null));

  const data = useMemo(
    () => buildBumpChartData(entries, dashboard.roundStandings, sortMode),
    [entries, dashboard.roundStandings, sortMode],
  );

  const colorMap = useMemo(() => getEntryColorById(entries), [entries]);
  const nameToColor = useMemo(() => {
    const m = new Map<string, string>();
    for (const entry of entries) {
      m.set(entry.name, colorMap.get(entry.id) ?? '#6B7280');
    }
    return m;
  }, [entries, colorMap]);

  if (data.length === 0) {
    return <p className="text-muted-foreground py-8 text-center">Not enough rounds to display the race chart.</p>;
  }

  const chartHeight = Math.max(400, entries.length * 30);

  return (
    <div className="grid gap-4">
      {(hasProjections || hasFavorites) && (
        <div className="flex gap-3 items-center">
          <Select value={sortMode} onChange={(e) => setSortMode(e.target.value as SortMode)} className="w-auto">
            <option value="actual">Actual Points</option>
            {hasProjections && <option value="projected">Projected EV</option>}
            {hasFavorites && <option value="favorites">Projected Favorites</option>}
          </Select>
        </div>
      )}

      <div style={{ height: chartHeight }}>
        <ResponsiveBump
          data={data}
          colors={(series) => nameToColor.get(series.id as string) ?? '#6B7280'}
          lineWidth={3}
          activeLineWidth={5}
          inactiveLineWidth={2}
          inactiveOpacity={0.3}
          pointSize={8}
          activePointSize={12}
          inactivePointSize={6}
          pointBorderWidth={2}
          pointBorderColor={{ from: 'serie.color' }}
          startLabel={true}
          startLabelPadding={16}
          endLabel={true}
          endLabelPadding={16}
          margin={{ top: 20, right: 120, bottom: 40, left: 120 }}
          axisBottom={{
            tickSize: 5,
            tickPadding: 5,
            tickRotation: 0,
          }}
          axisLeft={{
            tickSize: 5,
            tickPadding: 5,
          }}
        />
      </div>
    </div>
  );
}

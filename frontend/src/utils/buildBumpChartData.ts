import type { CalcuttaEntry, RoundStandingGroup } from '../schemas/calcutta';
import { getRoundName } from './roundLabels';

export interface BumpDatum {
  x: string;
  y: number;
}

export interface BumpSeries {
  id: string;
  data: BumpDatum[];
  [key: string]: unknown;
}

type Mode = 'actual' | 'projected' | 'favorites';

function metricForEntry(
  entry: { totalPoints: number; projectedEv?: number; projectedFavorites?: number },
  mode: Mode,
): number {
  if (mode === 'projected') return entry.projectedEv ?? entry.totalPoints;
  if (mode === 'favorites') return entry.projectedFavorites ?? entry.totalPoints;
  return entry.totalPoints;
}

function rankEntries(values: { id: string; value: number }[]): Map<string, number> {
  const sorted = [...values].sort((a, b) => b.value - a.value);
  const ranks = new Map<string, number>();
  let rank = 1;
  for (let i = 0; i < sorted.length; i++) {
    if (i > 0 && sorted[i].value < sorted[i - 1].value) {
      rank = i + 1;
    }
    ranks.set(sorted[i].id, rank);
  }
  return ranks;
}

export function buildBumpChartData(
  entries: CalcuttaEntry[],
  roundStandings: RoundStandingGroup[],
  mode: Mode,
): BumpSeries[] {
  if (roundStandings.length < 2) return [];

  const sorted = [...roundStandings].sort((a, b) => a.round - b.round);
  const maxRound = sorted[sorted.length - 1].round;
  const nameById = new Map(entries.map((e) => [e.id, e.name]));

  const seriesMap = new Map<string, BumpDatum[]>();
  for (const entry of entries) {
    seriesMap.set(entry.id, []);
  }

  for (const group of sorted) {
    const label = getRoundName(group.round, maxRound);
    const values = group.entries.map((e) => ({
      id: e.entryId,
      value: metricForEntry(e, mode),
    }));
    const ranks = rankEntries(values);

    for (const entry of entries) {
      const data = seriesMap.get(entry.id);
      if (data) {
        data.push({ x: label, y: ranks.get(entry.id) ?? entries.length });
      }
    }
  }

  return entries.map((e) => ({
    id: nameById.get(e.id) ?? e.id,
    data: seriesMap.get(e.id) ?? [],
  }));
}

import type { Portfolio, RoundStandingGroup } from '../schemas/pool';
import { getRoundName } from './roundLabels';

export interface RaceDatum {
  x: string;
  y: number;
  totalPoints: number;
  pointsDelta: number;
  rankDelta: number;
  expectedValue?: number;
}

export interface BumpSeries {
  id: string;
  data: RaceDatum[];
  [key: string]: unknown;
}

type Mode = 'actual' | 'projected' | 'favorites';

function metricForPortfolio(
  standing: { totalReturns: number; expectedValue?: number; projectedFavorites?: number },
  mode: Mode,
): number {
  if (mode === 'projected') return standing.expectedValue ?? standing.totalReturns;
  if (mode === 'favorites') return standing.projectedFavorites ?? standing.totalReturns;
  return standing.totalReturns;
}

function rankEntries(
  values: { id: string; value: number }[],
  tiebreaker?: Map<string, number>,
): Map<string, number> {
  const sorted = [...values].sort((a, b) => {
    if (b.value !== a.value) return b.value - a.value;
    if (tiebreaker) {
      return (tiebreaker.get(b.id) ?? 0) - (tiebreaker.get(a.id) ?? 0);
    }
    return 0;
  });
  const ranks = new Map<string, number>();
  let rank = 1;
  for (let i = 0; i < sorted.length; i++) {
    if (i > 0) {
      const prev = sorted[i - 1];
      const curr = sorted[i];
      const prevTb = tiebreaker?.get(prev.id) ?? 0;
      const currTb = tiebreaker?.get(curr.id) ?? 0;
      if (curr.value < prev.value || (tiebreaker && currTb < prevTb)) {
        rank = i + 1;
      }
    }
    ranks.set(sorted[i].id, rank);
  }
  return ranks;
}

export function buildBumpChartData(
  entries: Portfolio[],
  roundStandings: RoundStandingGroup[],
  mode: Mode,
): BumpSeries[] {
  if (roundStandings.length < 2) return [];

  const sorted = [...roundStandings].sort((a, b) => a.round - b.round);
  const maxRound = sorted[sorted.length - 1].round;
  const nameById = new Map(entries.map((e) => [e.id, e.name]));

  const seriesMap = new Map<string, RaceDatum[]>();
  for (const entry of entries) {
    seriesMap.set(entry.id, []);
  }

  const prevState = new Map<string, { rank: number; points: number }>();

  for (const group of sorted) {
    const label = getRoundName(group.round, maxRound);

    const portfolioLookup = new Map(group.entries.map((e) => [e.portfolioId, e]));

    const values = group.entries.map((e) => ({
      id: e.portfolioId,
      value: metricForPortfolio(e, mode),
    }));

    let tiebreaker: Map<string, number> | undefined;
    if (mode === 'actual') {
      const evMap = new Map<string, number>();
      for (const e of group.entries) {
        if (e.expectedValue != null) {
          evMap.set(e.portfolioId, e.expectedValue);
        }
      }
      if (evMap.size > 0) {
        tiebreaker = evMap;
      }
    }

    const ranks = rankEntries(values, tiebreaker);

    for (const entry of entries) {
      const data = seriesMap.get(entry.id);
      if (!data) continue;

      const standing = portfolioLookup.get(entry.id);
      const currentPoints = standing ? metricForPortfolio(standing, mode) : 0;
      const currentRank = ranks.get(entry.id) ?? entries.length;
      const prev = prevState.get(entry.id);

      data.push({
        x: label,
        y: currentRank,
        totalPoints: standing?.totalReturns ?? 0,
        pointsDelta: prev ? currentPoints - prev.points : 0,
        rankDelta: prev ? prev.rank - currentRank : 0,
        expectedValue: standing?.expectedValue,
      });

      prevState.set(entry.id, { rank: currentRank, points: currentPoints });
    }
  }

  return entries.map((e) => ({
    id: nameById.get(e.id) ?? e.id,
    data: seriesMap.get(e.id) ?? [],
  }));
}

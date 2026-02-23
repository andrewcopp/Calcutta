import type { ScoringRule, DashboardPayout, CalcuttaEntry } from '../schemas/calcutta';

export function pointsForProgress(rules: ScoringRule[], wins: number, byes: number): number {
  const p = wins + byes;
  if (p <= 0) return 0;
  let pts = 0;
  for (const r of rules) {
    if (r.winIndex <= p) {
      pts += r.pointsAwarded;
    }
  }
  return pts;
}

export interface Standing {
  entryId: string;
  totalPoints: number;
  finishPosition: number;
  isTied: boolean;
  payoutCents: number;
  inTheMoney: boolean;
}

export function computeStandings(
  entries: CalcuttaEntry[],
  pointsByEntry: Record<string, number>,
  payouts: DashboardPayout[],
): Standing[] {
  if (entries.length === 0) return [];

  const sorted = entries
    .map((e) => ({ entry: e, points: pointsByEntry[e.id] ?? 0 }))
    .sort((a, b) => {
      if (a.points !== b.points) return b.points - a.points;
      // Tie-break: later createdAt first (descending)
      return b.entry.createdAt.localeCompare(a.entry.createdAt);
    });

  const payoutByPosition: Record<number, number> = {};
  for (const p of payouts) {
    payoutByPosition[p.position] = p.amountCents;
  }

  const epsilon = 0.0001;
  const standings: Standing[] = new Array(sorted.length);

  let position = 1;
  let i = 0;
  while (i < sorted.length) {
    let j = i + 1;
    while (j < sorted.length && Math.abs(sorted[j].points - sorted[i].points) < epsilon) {
      j++;
    }

    const groupSize = j - i;
    const isTied = groupSize > 1;

    let totalGroupPayout = 0;
    for (let pos = position; pos < position + groupSize; pos++) {
      totalGroupPayout += payoutByPosition[pos] ?? 0;
    }

    const base = Math.floor(totalGroupPayout / groupSize);
    let remainder = totalGroupPayout % groupSize;

    for (let k = 0; k < groupSize; k++) {
      let payoutCents = base;
      if (remainder > 0) {
        payoutCents++;
        remainder--;
      }
      standings[i + k] = {
        entryId: sorted[i + k].entry.id,
        totalPoints: sorted[i + k].points,
        finishPosition: position,
        isTied,
        payoutCents,
        inTheMoney: payoutCents > 0,
      };
    }

    position += groupSize;
    i = j;
  }

  return standings;
}

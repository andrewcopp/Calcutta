import { CalcuttaWithRanking } from '../types/calcutta';

export function getEffectiveDate(calcutta: CalcuttaWithRanking): string {
  return calcutta.tournamentStartingAt ?? calcutta.createdAt;
}

export function groupCalcuttas(
  calcuttas: CalcuttaWithRanking[],
  now: Date = new Date(),
): { current: CalcuttaWithRanking[]; historical: CalcuttaWithRanking[] } {
  const oneYearAgo = new Date(now);
  oneYearAgo.setFullYear(oneYearAgo.getFullYear() - 1);

  const sorted = [...calcuttas].sort((a, b) => {
    const aHasStart = a.tournamentStartingAt != null;
    const bHasStart = b.tournamentStartingAt != null;

    if (aHasStart && !bHasStart) return -1;
    if (!aHasStart && bHasStart) return 1;

    const aDate = getEffectiveDate(a);
    const bDate = getEffectiveDate(b);
    return new Date(bDate).getTime() - new Date(aDate).getTime();
  });

  const current: CalcuttaWithRanking[] = [];
  const historical: CalcuttaWithRanking[] = [];

  for (const calcutta of sorted) {
    const effectiveDate = new Date(getEffectiveDate(calcutta));
    if (effectiveDate < oneYearAgo) {
      historical.push(calcutta);
    } else {
      current.push(calcutta);
    }
  }

  return { current, historical };
}

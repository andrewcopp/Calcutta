import { PoolWithRanking } from '../schemas/pool';

export function getEffectiveDate(pool: PoolWithRanking): string {
  return pool.tournamentStartingAt ?? pool.createdAt;
}

export function groupPools(
  pools: PoolWithRanking[],
  now: Date = new Date(),
): { current: PoolWithRanking[]; historical: PoolWithRanking[] } {
  const oneYearAgo = new Date(now);
  oneYearAgo.setFullYear(oneYearAgo.getFullYear() - 1);

  const sorted = [...pools].sort((a, b) => {
    const aHasStart = a.tournamentStartingAt != null;
    const bHasStart = b.tournamentStartingAt != null;

    if (aHasStart && !bHasStart) return -1;
    if (!aHasStart && bHasStart) return 1;

    const aDate = getEffectiveDate(a);
    const bDate = getEffectiveDate(b);
    return new Date(bDate).getTime() - new Date(aDate).getTime();
  });

  const current: PoolWithRanking[] = [];
  const historical: PoolWithRanking[] = [];

  for (const pool of sorted) {
    const effectiveDate = new Date(getEffectiveDate(pool));
    if (effectiveDate < oneYearAgo) {
      historical.push(pool);
    } else {
      current.push(pool);
    }
  }

  return { current, historical };
}

/**
 * Shared multi-level sort comparator for team rows.
 *
 * Every CalcuttaEntries tab (Investment, Ownership, Returns) and the
 * EntryTeams InvestmentsTab share identical seed / region / team name
 * fallback logic. This module extracts that logic into a single
 * reusable comparator.
 */

/** Minimal shape a row must satisfy to be sortable by this utility. */
export interface TeamSortRow {
  seed: number | undefined;
  region: string;
  teamName: string;
}

/** Sort keys supported by the shared comparator. */
export type TeamSortKey = 'seed' | 'region' | 'team';

/**
 * Creates a comparator function for sorting team rows by the given key.
 *
 * The comparator applies multi-level tiebreakers matching the pattern
 * used across all tab components:
 *
 * - **seed**:   seed asc -> region asc -> teamName asc
 * - **region**: region asc -> seed asc  -> teamName asc
 * - **team**:   teamName asc -> seed asc
 *
 * Undefined seeds sort last (treated as 999).
 */
export function createTeamSortComparator<T extends TeamSortRow>(sortKey: TeamSortKey): (a: T, b: T) => number {
  const seedVal = (seed: number | undefined) => seed ?? 999;
  const regionVal = (region: string) => region || '';
  const teamVal = (name: string) => name || '';

  return (a: T, b: T): number => {
    if (sortKey === 'seed') {
      const seedDiff = seedVal(a.seed) - seedVal(b.seed);
      if (seedDiff !== 0) return seedDiff;
      const regionDiff = regionVal(a.region).localeCompare(regionVal(b.region));
      if (regionDiff !== 0) return regionDiff;
      return teamVal(a.teamName).localeCompare(teamVal(b.teamName));
    }

    if (sortKey === 'region') {
      const regionDiff = regionVal(a.region).localeCompare(regionVal(b.region));
      if (regionDiff !== 0) return regionDiff;
      const seedDiff = seedVal(a.seed) - seedVal(b.seed);
      if (seedDiff !== 0) return seedDiff;
      return teamVal(a.teamName).localeCompare(teamVal(b.teamName));
    }

    // sortKey === 'team'
    const nameDiff = teamVal(a.teamName).localeCompare(teamVal(b.teamName));
    if (nameDiff !== 0) return nameDiff;
    return seedVal(a.seed) - seedVal(b.seed);
  };
}

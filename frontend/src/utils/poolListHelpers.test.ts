import { describe, it, expect } from 'vitest';
import { PoolWithRanking } from '../schemas/pool';
import { groupPools } from './poolListHelpers';

function makePool(overrides: Partial<PoolWithRanking> & { id: string }): PoolWithRanking {
  return {
    name: 'Test',
    tournamentId: 't1',
    ownerId: 'o1',
    minTeams: 1,
    maxTeams: 64,
    maxInvestmentCredits: 100,
    budgetCredits: 1000,
    createdAt: '2025-01-01T00:00:00Z',
    updatedAt: '2025-01-01T00:00:00Z',
    hasPortfolio: false,
    ...overrides,
  };
}

const NOW = new Date('2026-02-20T00:00:00Z');

describe('groupPools', () => {
  it('sorts by tournamentStartingAt descending', () => {
    // GIVEN two pools with different tournament start dates
    const older = makePool({ id: 'a', tournamentStartingAt: '2026-01-01T00:00:00Z' });
    const newer = makePool({ id: 'b', tournamentStartingAt: '2026-02-01T00:00:00Z' });

    // WHEN grouping with older first in input
    const { current } = groupPools([older, newer], NOW);

    // THEN newer tournament appears first
    expect(current.map((c) => c.id)).toEqual(['b', 'a']);
  });

  it('places pools without tournamentStartingAt after dated ones', () => {
    // GIVEN one pool with a start date and one without
    const withStart = makePool({ id: 'a', tournamentStartingAt: '2026-01-01T00:00:00Z' });
    const withoutStart = makePool({ id: 'b', createdAt: '2026-02-15T00:00:00Z' });

    // WHEN grouping
    const { current } = groupPools([withoutStart, withStart], NOW);

    // THEN the one with tournamentStartingAt comes first
    expect(current.map((c) => c.id)).toEqual(['a', 'b']);
  });

  it('groups pools older than one year into historical', () => {
    // GIVEN a pool with a tournament start date over one year ago
    const old = makePool({ id: 'a', tournamentStartingAt: '2025-01-01T00:00:00Z' });

    // WHEN grouping
    const { historical } = groupPools([old], NOW);

    // THEN it appears in historical
    expect(historical.map((c) => c.id)).toEqual(['a']);
  });

  it('uses created date as fallback for grouping', () => {
    // GIVEN a pool without tournamentStartingAt and created over one year ago
    const old = makePool({ id: 'a', createdAt: '2024-12-01T00:00:00Z' });

    // WHEN grouping
    const { historical } = groupPools([old], NOW);

    // THEN it appears in historical based on created date
    expect(historical.map((c) => c.id)).toEqual(['a']);
  });

  it('returns empty arrays when given no pools', () => {
    // GIVEN no pools
    // WHEN grouping
    const result = groupPools([], NOW);

    // THEN both arrays are empty
    expect(result).toEqual({ current: [], historical: [] });
  });

  it('treats exactly one year old as current', () => {
    // GIVEN a pool with tournamentStartingAt exactly one year ago
    const exact = makePool({ id: 'a', tournamentStartingAt: '2025-02-20T00:00:00Z' });

    // WHEN grouping
    const { current } = groupPools([exact], NOW);

    // THEN it appears in current (not historical)
    expect(current.map((c) => c.id)).toEqual(['a']);
  });

  it('treats one year plus one day as historical', () => {
    // GIVEN a pool with tournamentStartingAt one year and one day ago
    const old = makePool({ id: 'a', tournamentStartingAt: '2025-02-19T00:00:00Z' });

    // WHEN grouping
    const { historical } = groupPools([old], NOW);

    // THEN it appears in historical
    expect(historical.map((c) => c.id)).toEqual(['a']);
  });
});

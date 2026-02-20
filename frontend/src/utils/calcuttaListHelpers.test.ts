import { describe, it, expect } from 'vitest';
import { CalcuttaWithRanking } from '../types/calcutta';
import { groupCalcuttas } from './calcuttaListHelpers';

function makeCalcutta(
  overrides: Partial<CalcuttaWithRanking> & { id: string },
): CalcuttaWithRanking {
  return {
    name: 'Test',
    tournamentId: 't1',
    ownerId: 'o1',
    minTeams: 1,
    maxTeams: 64,
    maxBid: 100,
    budgetPoints: 1000,
    biddingOpen: false,
    created: '2025-01-01T00:00:00Z',
    updated: '2025-01-01T00:00:00Z',
    hasEntry: false,
    ...overrides,
  };
}

const NOW = new Date('2026-02-20T00:00:00Z');

describe('groupCalcuttas', () => {
  it('sorts by tournamentStartingAt descending', () => {
    // GIVEN two calcuttas with different tournament start dates
    const older = makeCalcutta({ id: 'a', tournamentStartingAt: '2026-01-01T00:00:00Z' });
    const newer = makeCalcutta({ id: 'b', tournamentStartingAt: '2026-02-01T00:00:00Z' });

    // WHEN grouping with older first in input
    const { current } = groupCalcuttas([older, newer], NOW);

    // THEN newer tournament appears first
    expect(current.map((c) => c.id)).toEqual(['b', 'a']);
  });

  it('places calcuttas without tournamentStartingAt after dated ones', () => {
    // GIVEN one calcutta with a start date and one without
    const withStart = makeCalcutta({ id: 'a', tournamentStartingAt: '2026-01-01T00:00:00Z' });
    const withoutStart = makeCalcutta({ id: 'b', created: '2026-02-15T00:00:00Z' });

    // WHEN grouping
    const { current } = groupCalcuttas([withoutStart, withStart], NOW);

    // THEN the one with tournamentStartingAt comes first
    expect(current.map((c) => c.id)).toEqual(['a', 'b']);
  });

  it('groups calcuttas older than one year into historical', () => {
    // GIVEN a calcutta with a tournament start date over one year ago
    const old = makeCalcutta({ id: 'a', tournamentStartingAt: '2025-01-01T00:00:00Z' });

    // WHEN grouping
    const { historical } = groupCalcuttas([old], NOW);

    // THEN it appears in historical
    expect(historical.map((c) => c.id)).toEqual(['a']);
  });

  it('uses created date as fallback for grouping', () => {
    // GIVEN a calcutta without tournamentStartingAt and created over one year ago
    const old = makeCalcutta({ id: 'a', created: '2024-12-01T00:00:00Z' });

    // WHEN grouping
    const { historical } = groupCalcuttas([old], NOW);

    // THEN it appears in historical based on created date
    expect(historical.map((c) => c.id)).toEqual(['a']);
  });

  it('returns empty arrays when given no calcuttas', () => {
    // GIVEN no calcuttas
    // WHEN grouping
    const result = groupCalcuttas([], NOW);

    // THEN both arrays are empty
    expect(result).toEqual({ current: [], historical: [] });
  });

  it('treats exactly one year old as current', () => {
    // GIVEN a calcutta with tournamentStartingAt exactly one year ago
    const exact = makeCalcutta({ id: 'a', tournamentStartingAt: '2025-02-20T00:00:00Z' });

    // WHEN grouping
    const { current } = groupCalcuttas([exact], NOW);

    // THEN it appears in current (not historical)
    expect(current.map((c) => c.id)).toEqual(['a']);
  });

  it('treats one year plus one day as historical', () => {
    // GIVEN a calcutta with tournamentStartingAt one year and one day ago
    const old = makeCalcutta({ id: 'a', tournamentStartingAt: '2025-02-19T00:00:00Z' });

    // WHEN grouping
    const { historical } = groupCalcuttas([old], NOW);

    // THEN it appears in historical
    expect(historical.map((c) => c.id)).toEqual(['a']);
  });
});

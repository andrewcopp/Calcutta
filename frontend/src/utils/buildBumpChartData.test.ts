import { describe, it, expect } from 'vitest';
import { buildBumpChartData } from './buildBumpChartData';
import { makePortfolio } from '../test/factories';
import type { RoundStandingGroup } from '../schemas/pool';

function makeStandingGroup(
  round: number,
  entries: { entryId: string; totalPoints: number; expectedValue?: number; projectedFavorites?: number }[],
): RoundStandingGroup {
  return {
    round,
    entries: entries.map((e) => ({
      portfolioId: e.entryId,
      totalReturns: e.totalPoints,
      finishPosition: 0,
      isTied: false,
      payoutCents: 0,
      inTheMoney: false,
      expectedValue: e.expectedValue,
      projectedFavorites: e.projectedFavorites,
    })),
  };
}

describe('buildBumpChartData', () => {
  it('returns empty when fewer than 2 rounds', () => {
    const entries = [makePortfolio({ id: 'e1' })];
    const standings = [makeStandingGroup(0, [{ entryId: 'e1', totalPoints: 10 }])];
    expect(buildBumpChartData(entries, standings, 'actual')).toEqual([]);
  });

  it('uses entry name as series id', () => {
    const entries = [makePortfolio({ id: 'e1', name: 'Alice' }), makePortfolio({ id: 'e2', name: 'Bob' })];
    const standings = [
      makeStandingGroup(0, [
        { entryId: 'e1', totalPoints: 0 },
        { entryId: 'e2', totalPoints: 0 },
      ]),
      makeStandingGroup(1, [
        { entryId: 'e1', totalPoints: 5 },
        { entryId: 'e2', totalPoints: 3 },
      ]),
    ];
    const result = buildBumpChartData(entries, standings, 'actual');
    expect(result.map((s) => s.id)).toEqual(['Alice', 'Bob']);
  });

  it('ranks higher-scoring entries with lower rank number', () => {
    const entries = [makePortfolio({ id: 'e1', name: 'Alice' }), makePortfolio({ id: 'e2', name: 'Bob' })];
    const standings = [
      makeStandingGroup(0, [
        { entryId: 'e1', totalPoints: 0 },
        { entryId: 'e2', totalPoints: 0 },
      ]),
      makeStandingGroup(1, [
        { entryId: 'e1', totalPoints: 10 },
        { entryId: 'e2', totalPoints: 5 },
      ]),
    ];
    const result = buildBumpChartData(entries, standings, 'actual');
    const aliceRound1 = result.find((s) => s.id === 'Alice')!.data[1].y;
    expect(aliceRound1).toBe(1);
  });

  it('breaks ties by name when no EV available', () => {
    const entries = [
      makePortfolio({ id: 'e1', name: 'Alice' }),
      makePortfolio({ id: 'e2', name: 'Bob' }),
      makePortfolio({ id: 'e3', name: 'Carol' }),
    ];
    const standings = [
      makeStandingGroup(0, [
        { entryId: 'e1', totalPoints: 0 },
        { entryId: 'e2', totalPoints: 0 },
        { entryId: 'e3', totalPoints: 0 },
      ]),
      makeStandingGroup(1, [
        { entryId: 'e1', totalPoints: 10 },
        { entryId: 'e2', totalPoints: 10 },
        { entryId: 'e3', totalPoints: 5 },
      ]),
    ];
    const result = buildBumpChartData(entries, standings, 'actual');
    const aliceRank = result.find((s) => s.id === 'Alice')!.data[1].y;
    const bobRank = result.find((s) => s.id === 'Bob')!.data[1].y;
    expect(aliceRank).toBe(1);
    expect(bobRank).toBe(2);
  });

  it('produces one data point per round per entry', () => {
    const entries = [makePortfolio({ id: 'e1', name: 'Alice' }), makePortfolio({ id: 'e2', name: 'Bob' })];
    const standings = [
      makeStandingGroup(0, [
        { entryId: 'e1', totalPoints: 0 },
        { entryId: 'e2', totalPoints: 0 },
      ]),
      makeStandingGroup(1, [
        { entryId: 'e1', totalPoints: 5 },
        { entryId: 'e2', totalPoints: 3 },
      ]),
      makeStandingGroup(2, [
        { entryId: 'e1', totalPoints: 8 },
        { entryId: 'e2', totalPoints: 12 },
      ]),
    ];
    const result = buildBumpChartData(entries, standings, 'actual');
    expect(result[0].data).toHaveLength(3);
  });

  it('uses projected EV when mode is projected', () => {
    const entries = [makePortfolio({ id: 'e1', name: 'Alice' }), makePortfolio({ id: 'e2', name: 'Bob' })];
    const standings = [
      makeStandingGroup(0, [
        { entryId: 'e1', totalPoints: 0, expectedValue: 50 },
        { entryId: 'e2', totalPoints: 0, expectedValue: 40 },
      ]),
      makeStandingGroup(1, [
        { entryId: 'e1', totalPoints: 5, expectedValue: 30 },
        { entryId: 'e2', totalPoints: 3, expectedValue: 45 },
      ]),
    ];
    const result = buildBumpChartData(entries, standings, 'projected');
    const bobRound1 = result.find((s) => s.id === 'Bob')!.data[1].y;
    expect(bobRound1).toBe(1);
  });

  it('uses round names as x-axis labels', () => {
    const entries = [makePortfolio({ id: 'e1', name: 'Alice' })];
    const standings = [
      makeStandingGroup(0, [{ entryId: 'e1', totalPoints: 0 }]),
      makeStandingGroup(1, [{ entryId: 'e1', totalPoints: 5 }]),
      makeStandingGroup(2, [{ entryId: 'e1', totalPoints: 10 }]),
    ];
    const result = buildBumpChartData(entries, standings, 'actual');
    expect(result[0].data.map((d) => d.x)).toEqual(['Start', 'Round of 64', 'Round of 32']);
  });
});

describe('tiebreaker', () => {
  it('breaks ties using projected EV in actual mode', () => {
    // GIVEN two entries with same actual points but different EV
    const entries = [makePortfolio({ id: 'e1', name: 'Alice' }), makePortfolio({ id: 'e2', name: 'Bob' })];
    const standings = [
      makeStandingGroup(0, [
        { entryId: 'e1', totalPoints: 0, expectedValue: 80 },
        { entryId: 'e2', totalPoints: 0, expectedValue: 120 },
      ]),
      makeStandingGroup(1, [
        { entryId: 'e1', totalPoints: 10, expectedValue: 80 },
        { entryId: 'e2', totalPoints: 10, expectedValue: 120 },
      ]),
    ];

    // WHEN building chart data in actual mode
    const result = buildBumpChartData(entries, standings, 'actual');

    // THEN higher EV entry ranks first
    expect(result.find((s) => s.id === 'Bob')!.data[1].y).toBe(1);
  });

  it('breaks ties by name when no projected EV available', () => {
    // GIVEN two entries with same points and no EV
    const entries = [makePortfolio({ id: 'e1', name: 'Alice' }), makePortfolio({ id: 'e2', name: 'Bob' })];
    const standings = [
      makeStandingGroup(0, [
        { entryId: 'e1', totalPoints: 0 },
        { entryId: 'e2', totalPoints: 0 },
      ]),
      makeStandingGroup(1, [
        { entryId: 'e1', totalPoints: 10 },
        { entryId: 'e2', totalPoints: 10 },
      ]),
    ];

    // WHEN building chart data in actual mode
    const result = buildBumpChartData(entries, standings, 'actual');

    // THEN Alice ranks first alphabetically
    expect(result.find((s) => s.id === 'Alice')!.data[1].y).toBe(1);
    expect(result.find((s) => s.id === 'Bob')!.data[1].y).toBe(2);
  });

  it('breaks projected mode ties by name', () => {
    // GIVEN two entries with same projected EV
    const entries = [makePortfolio({ id: 'e1', name: 'Alice' }), makePortfolio({ id: 'e2', name: 'Bob' })];
    const standings = [
      makeStandingGroup(0, [
        { entryId: 'e1', totalPoints: 0, expectedValue: 50 },
        { entryId: 'e2', totalPoints: 0, expectedValue: 50 },
      ]),
      makeStandingGroup(1, [
        { entryId: 'e1', totalPoints: 5, expectedValue: 50 },
        { entryId: 'e2', totalPoints: 3, expectedValue: 50 },
      ]),
    ];

    // WHEN building chart data in projected mode
    const result = buildBumpChartData(entries, standings, 'projected');

    // THEN Alice ranks first alphabetically at round 0 (same EV, name breaks tie)
    expect(result.find((s) => s.id === 'Alice')!.data[0].y).toBe(1);
    expect(result.find((s) => s.id === 'Bob')!.data[0].y).toBe(2);
  });
});

describe('datum metadata', () => {
  it('includes totalPoints in datum', () => {
    // GIVEN an entry with 15 total points
    const entries = [makePortfolio({ id: 'e1', name: 'Alice' })];
    const standings = [
      makeStandingGroup(0, [{ entryId: 'e1', totalPoints: 0 }]),
      makeStandingGroup(1, [{ entryId: 'e1', totalPoints: 15 }]),
    ];

    // WHEN building chart data
    const result = buildBumpChartData(entries, standings, 'actual');

    // THEN datum includes totalPoints
    expect(result[0].data[1].totalPoints).toBe(15);
  });

  it('computes pointsDelta as difference from previous round', () => {
    // GIVEN an entry gaining points across rounds
    const entries = [makePortfolio({ id: 'e1', name: 'Alice' })];
    const standings = [
      makeStandingGroup(0, [{ entryId: 'e1', totalPoints: 0 }]),
      makeStandingGroup(1, [{ entryId: 'e1', totalPoints: 10 }]),
      makeStandingGroup(2, [{ entryId: 'e1', totalPoints: 22.5 }]),
    ];

    // WHEN building chart data
    const result = buildBumpChartData(entries, standings, 'actual');

    // THEN pointsDelta reflects round-over-round change
    expect(result[0].data[2].pointsDelta).toBe(12.5);
  });

  it('computes rankDelta as improvement since previous round', () => {
    // GIVEN Alice overtakes Bob in round 2
    const entries = [makePortfolio({ id: 'e1', name: 'Alice' }), makePortfolio({ id: 'e2', name: 'Bob' })];
    const standings = [
      makeStandingGroup(0, [
        { entryId: 'e1', totalPoints: 0 },
        { entryId: 'e2', totalPoints: 0 },
      ]),
      makeStandingGroup(1, [
        { entryId: 'e1', totalPoints: 5 },
        { entryId: 'e2', totalPoints: 10 },
      ]),
      makeStandingGroup(2, [
        { entryId: 'e1', totalPoints: 20 },
        { entryId: 'e2', totalPoints: 12 },
      ]),
    ];

    // WHEN building chart data
    const result = buildBumpChartData(entries, standings, 'actual');

    // THEN Alice's rankDelta is positive (moved from 2 to 1 = +1)
    expect(result.find((s) => s.id === 'Alice')!.data[2].rankDelta).toBe(1);
  });

  it('sets deltas to zero for first round', () => {
    // GIVEN an entry at round 0
    const entries = [makePortfolio({ id: 'e1', name: 'Alice' })];
    const standings = [
      makeStandingGroup(0, [{ entryId: 'e1', totalPoints: 5 }]),
      makeStandingGroup(1, [{ entryId: 'e1', totalPoints: 10 }]),
    ];

    // WHEN building chart data
    const result = buildBumpChartData(entries, standings, 'actual');

    // THEN first round deltas are zero
    expect(result[0].data[0].pointsDelta).toBe(0);
  });

  it('sets rankDelta to zero for first round', () => {
    // GIVEN an entry at round 0
    const entries = [makePortfolio({ id: 'e1', name: 'Alice' })];
    const standings = [
      makeStandingGroup(0, [{ entryId: 'e1', totalPoints: 5 }]),
      makeStandingGroup(1, [{ entryId: 'e1', totalPoints: 10 }]),
    ];

    // WHEN building chart data
    const result = buildBumpChartData(entries, standings, 'actual');

    // THEN first round rank delta is zero
    expect(result[0].data[0].rankDelta).toBe(0);
  });

  it('includes expectedValue in datum when available', () => {
    // GIVEN an entry with projected EV
    const entries = [makePortfolio({ id: 'e1', name: 'Alice' })];
    const standings = [
      makeStandingGroup(0, [{ entryId: 'e1', totalPoints: 0, expectedValue: 95.5 }]),
      makeStandingGroup(1, [{ entryId: 'e1', totalPoints: 10, expectedValue: 88.3 }]),
    ];

    // WHEN building chart data
    const result = buildBumpChartData(entries, standings, 'actual');

    // THEN datum includes expectedValue
    expect(result[0].data[1].expectedValue).toBe(88.3);
  });
});

import { describe, it, expect } from 'vitest';
import { buildBumpChartData } from './buildBumpChartData';
import { makeEntry } from '../test/factories';
import type { RoundStandingGroup } from '../schemas/calcutta';

function makeStandingGroup(
  round: number,
  entries: { entryId: string; totalPoints: number; projectedEv?: number; projectedFavorites?: number }[],
): RoundStandingGroup {
  return {
    round,
    entries: entries.map((e) => ({
      entryId: e.entryId,
      totalPoints: e.totalPoints,
      finishPosition: 0,
      isTied: false,
      payoutCents: 0,
      inTheMoney: false,
      projectedEv: e.projectedEv,
      projectedFavorites: e.projectedFavorites,
    })),
  };
}

describe('buildBumpChartData', () => {
  it('returns empty when fewer than 2 rounds', () => {
    const entries = [makeEntry({ id: 'e1' })];
    const standings = [makeStandingGroup(0, [{ entryId: 'e1', totalPoints: 10 }])];
    expect(buildBumpChartData(entries, standings, 'actual')).toEqual([]);
  });

  it('uses entry name as series id', () => {
    const entries = [makeEntry({ id: 'e1', name: 'Alice' }), makeEntry({ id: 'e2', name: 'Bob' })];
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
    const entries = [makeEntry({ id: 'e1', name: 'Alice' }), makeEntry({ id: 'e2', name: 'Bob' })];
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

  it('assigns same rank to tied entries', () => {
    const entries = [
      makeEntry({ id: 'e1', name: 'Alice' }),
      makeEntry({ id: 'e2', name: 'Bob' }),
      makeEntry({ id: 'e3', name: 'Carol' }),
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
    expect(aliceRank).toBe(bobRank);
  });

  it('produces one data point per round per entry', () => {
    const entries = [makeEntry({ id: 'e1', name: 'Alice' }), makeEntry({ id: 'e2', name: 'Bob' })];
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
    const entries = [makeEntry({ id: 'e1', name: 'Alice' }), makeEntry({ id: 'e2', name: 'Bob' })];
    const standings = [
      makeStandingGroup(0, [
        { entryId: 'e1', totalPoints: 0, projectedEv: 50 },
        { entryId: 'e2', totalPoints: 0, projectedEv: 40 },
      ]),
      makeStandingGroup(1, [
        { entryId: 'e1', totalPoints: 5, projectedEv: 30 },
        { entryId: 'e2', totalPoints: 3, projectedEv: 45 },
      ]),
    ];
    const result = buildBumpChartData(entries, standings, 'projected');
    const bobRound1 = result.find((s) => s.id === 'Bob')!.data[1].y;
    expect(bobRound1).toBe(1);
  });

  it('uses round names as x-axis labels', () => {
    const entries = [makeEntry({ id: 'e1', name: 'Alice' })];
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
    const entries = [makeEntry({ id: 'e1', name: 'Alice' }), makeEntry({ id: 'e2', name: 'Bob' })];
    const standings = [
      makeStandingGroup(0, [
        { entryId: 'e1', totalPoints: 0, projectedEv: 80 },
        { entryId: 'e2', totalPoints: 0, projectedEv: 120 },
      ]),
      makeStandingGroup(1, [
        { entryId: 'e1', totalPoints: 10, projectedEv: 80 },
        { entryId: 'e2', totalPoints: 10, projectedEv: 120 },
      ]),
    ];

    // WHEN building chart data in actual mode
    const result = buildBumpChartData(entries, standings, 'actual');

    // THEN higher EV entry ranks first
    expect(result.find((s) => s.id === 'Bob')!.data[1].y).toBe(1);
  });

  it('preserves tied ranks when no projected EV available', () => {
    // GIVEN two entries with same points and no EV
    const entries = [makeEntry({ id: 'e1', name: 'Alice' }), makeEntry({ id: 'e2', name: 'Bob' })];
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

    // THEN both entries share rank 1
    expect(result.find((s) => s.id === 'Alice')!.data[1].y).toBe(
      result.find((s) => s.id === 'Bob')!.data[1].y,
    );
  });

  it('does not use EV tiebreaker in projected mode', () => {
    // GIVEN two entries with same projected EV
    const entries = [makeEntry({ id: 'e1', name: 'Alice' }), makeEntry({ id: 'e2', name: 'Bob' })];
    const standings = [
      makeStandingGroup(0, [
        { entryId: 'e1', totalPoints: 0, projectedEv: 50 },
        { entryId: 'e2', totalPoints: 0, projectedEv: 50 },
      ]),
      makeStandingGroup(1, [
        { entryId: 'e1', totalPoints: 5, projectedEv: 50 },
        { entryId: 'e2', totalPoints: 3, projectedEv: 50 },
      ]),
    ];

    // WHEN building chart data in projected mode
    const result = buildBumpChartData(entries, standings, 'projected');

    // THEN both share rank 1 at round 0 (same EV is the primary metric, no tiebreaker)
    expect(result.find((s) => s.id === 'Alice')!.data[0].y).toBe(
      result.find((s) => s.id === 'Bob')!.data[0].y,
    );
  });
});

describe('datum metadata', () => {
  it('includes totalPoints in datum', () => {
    // GIVEN an entry with 15 total points
    const entries = [makeEntry({ id: 'e1', name: 'Alice' })];
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
    const entries = [makeEntry({ id: 'e1', name: 'Alice' })];
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
    const entries = [makeEntry({ id: 'e1', name: 'Alice' }), makeEntry({ id: 'e2', name: 'Bob' })];
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
    const entries = [makeEntry({ id: 'e1', name: 'Alice' })];
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
    const entries = [makeEntry({ id: 'e1', name: 'Alice' })];
    const standings = [
      makeStandingGroup(0, [{ entryId: 'e1', totalPoints: 5 }]),
      makeStandingGroup(1, [{ entryId: 'e1', totalPoints: 10 }]),
    ];

    // WHEN building chart data
    const result = buildBumpChartData(entries, standings, 'actual');

    // THEN first round rank delta is zero
    expect(result[0].data[0].rankDelta).toBe(0);
  });

  it('includes projectedEv in datum when available', () => {
    // GIVEN an entry with projected EV
    const entries = [makeEntry({ id: 'e1', name: 'Alice' })];
    const standings = [
      makeStandingGroup(0, [{ entryId: 'e1', totalPoints: 0, projectedEv: 95.5 }]),
      makeStandingGroup(1, [{ entryId: 'e1', totalPoints: 10, projectedEv: 88.3 }]),
    ];

    // WHEN building chart data
    const result = buildBumpChartData(entries, standings, 'actual');

    // THEN datum includes projectedEv
    expect(result[0].data[1].projectedEv).toBe(88.3);
  });
});

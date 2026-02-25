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
    const bobRound1 = result.find((s) => s.id === 'Bob')!.data[1].y;
    expect(aliceRound1).toBe(1);
    expect(bobRound1).toBe(2);
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
    const carolRank = result.find((s) => s.id === 'Carol')!.data[1].y;
    expect(aliceRank).toBe(1);
    expect(bobRank).toBe(1);
    expect(carolRank).toBe(3);
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
    expect(result[1].data).toHaveLength(3);
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
    const aliceRound1 = result.find((s) => s.id === 'Alice')!.data[1].y;
    const bobRound1 = result.find((s) => s.id === 'Bob')!.data[1].y;
    expect(bobRound1).toBe(1);
    expect(aliceRound1).toBe(2);
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

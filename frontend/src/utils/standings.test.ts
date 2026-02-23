import { describe, it, expect } from 'vitest';
import { makeEntry } from '../test/factories';
import { pointsForProgress, computeStandings, type Standing } from './standings';
import type { ScoringRule, DashboardPayout, CalcuttaEntry } from '../schemas/calcutta';

function standingsByID(standings: Standing[]): Record<string, Standing> {
  const m: Record<string, Standing> = {};
  for (const s of standings) {
    m[s.entryId] = s;
  }
  return m;
}

describe('pointsForProgress', () => {
  const rules: ScoringRule[] = [
    { winIndex: 1, pointsAwarded: 1 },
    { winIndex: 2, pointsAwarded: 2 },
    { winIndex: 3, pointsAwarded: 4 },
  ];

  it('returns 0 when wins + byes <= 0', () => {
    expect(pointsForProgress(rules, 0, 0)).toBe(0);
  });

  it('sums points for all rules at or below progress', () => {
    expect(pointsForProgress(rules, 2, 0)).toBe(3);
  });

  it('includes byes in progress', () => {
    expect(pointsForProgress(rules, 1, 1)).toBe(3);
  });

  it('sums all rules when progress exceeds max win index', () => {
    expect(pointsForProgress(rules, 5, 0)).toBe(7);
  });

  it('returns 0 with empty rules', () => {
    expect(pointsForProgress([], 3, 0)).toBe(0);
  });
});

describe('computeStandings', () => {
  it('returns empty array for empty entries', () => {
    expect(computeStandings([], {}, [])).toEqual([]);
  });

  it('sorts by total points descending', () => {
    const e1 = makeEntry({ id: 'e1', createdAt: '2026-01-01T00:00:01Z' });
    const e2 = makeEntry({ id: 'e2', createdAt: '2026-01-01T00:00:02Z' });
    const points = { e1: 10, e2: 20 };
    const standings = computeStandings([e1, e2], points, []);
    expect(standings[0].entryId).toBe('e2');
  });

  it('sorts ties by createdAt descending', () => {
    const eOld = makeEntry({ id: 'old', createdAt: '2026-01-01T00:00:01Z' });
    const eNew = makeEntry({ id: 'new', createdAt: '2026-01-01T00:00:02Z' });
    const points = { old: 10, new: 10 };
    const standings = computeStandings([eOld, eNew], points, []);
    expect(standings[0].entryId).toBe('new');
  });

  it('marks ties within epsilon', () => {
    const e1 = makeEntry({ id: 'e1', createdAt: '2026-01-01T00:00:02Z' });
    const e2 = makeEntry({ id: 'e2', createdAt: '2026-01-01T00:00:01Z' });
    const points = { e1: 10.0, e2: 10.00001 };
    const byID = standingsByID(computeStandings([e1, e2], points, []));
    expect(byID['e1'].isTied).toBe(true);
  });

  it('sets finish position 1 for top entry', () => {
    const e1 = makeEntry({ id: 'e1', createdAt: '2026-01-01T00:00:01Z' });
    const e2 = makeEntry({ id: 'e2', createdAt: '2026-01-01T00:00:01Z' });
    const points = { e1: 20, e2: 10 };
    const byID = standingsByID(computeStandings([e2, e1], points, []));
    expect(byID['e1'].finishPosition).toBe(1);
  });

  it('splits payout across tie group', () => {
    const e1 = makeEntry({ id: 'e1', createdAt: '2026-01-01T00:00:02Z' });
    const e2 = makeEntry({ id: 'e2', createdAt: '2026-01-01T00:00:01Z' });
    const points = { e1: 10, e2: 10 };
    const payouts: DashboardPayout[] = [
      { position: 1, amountCents: 100 },
      { position: 2, amountCents: 50 },
    ];
    const byID = standingsByID(computeStandings([e1, e2], points, payouts));
    expect(byID['e1'].payoutCents).toBe(75);
  });

  it('distributes remainder to earlier entry in sorted order', () => {
    const e1 = makeEntry({ id: 'e1', createdAt: '2026-01-01T00:00:02Z' });
    const e2 = makeEntry({ id: 'e2', createdAt: '2026-01-01T00:00:01Z' });
    const points = { e1: 10, e2: 10 };
    const payouts: DashboardPayout[] = [
      { position: 1, amountCents: 100 },
      { position: 2, amountCents: 99 },
    ];
    const byID = standingsByID(computeStandings([e1, e2], points, payouts));
    expect(byID['e1'].payoutCents).toBe(100);
  });

  it('sets inTheMoney when payout is positive', () => {
    const e1 = makeEntry({ id: 'e1', createdAt: '2026-01-01T00:00:01Z' });
    const points = { e1: 10 };
    const payouts: DashboardPayout[] = [{ position: 1, amountCents: 1 }];
    const byID = standingsByID(computeStandings([e1], points, payouts));
    expect(byID['e1'].inTheMoney).toBe(true);
  });

  it('three-way tie pools and splits payouts evenly', () => {
    const e1 = makeEntry({ id: 'e1', createdAt: '2026-01-01T00:00:03Z' });
    const e2 = makeEntry({ id: 'e2', createdAt: '2026-01-01T00:00:02Z' });
    const e3 = makeEntry({ id: 'e3', createdAt: '2026-01-01T00:00:01Z' });
    const points = { e1: 10, e2: 10, e3: 10 };
    const payouts: DashboardPayout[] = [
      { position: 1, amountCents: 300 },
      { position: 2, amountCents: 150 },
      { position: 3, amountCents: 150 },
    ];
    const byID = standingsByID(computeStandings([e1, e2, e3], points, payouts));
    expect(byID['e1'].payoutCents).toBe(200);
  });

  it('three-way tie distributes remainder to earliest entries in sorted order', () => {
    const e1 = makeEntry({ id: 'e1', createdAt: '2026-01-01T00:00:03Z' });
    const e2 = makeEntry({ id: 'e2', createdAt: '2026-01-01T00:00:02Z' });
    const e3 = makeEntry({ id: 'e3', createdAt: '2026-01-01T00:00:01Z' });
    const points = { e1: 10, e2: 10, e3: 10 };
    const payouts: DashboardPayout[] = [
      { position: 1, amountCents: 100 },
      { position: 2, amountCents: 100 },
      { position: 3, amountCents: 101 },
    ];
    const byID = standingsByID(computeStandings([e1, e2, e3], points, payouts));
    expect(byID['e1'].payoutCents).toBe(101);
    expect(byID['e2'].payoutCents).toBe(100);
  });

  it('tie outside payout positions results in zero payout', () => {
    const e1 = makeEntry({ id: 'e1', createdAt: '2026-01-01T00:00:04Z' });
    const e2 = makeEntry({ id: 'e2', createdAt: '2026-01-01T00:00:03Z' });
    const e3 = makeEntry({ id: 'e3', createdAt: '2026-01-01T00:00:02Z' });
    const e4 = makeEntry({ id: 'e4', createdAt: '2026-01-01T00:00:01Z' });
    const points = { e1: 20, e2: 15, e3: 10, e4: 10 };
    const payouts: DashboardPayout[] = [
      { position: 1, amountCents: 200 },
      { position: 2, amountCents: 100 },
    ];
    const byID = standingsByID(computeStandings([e1, e2, e3, e4], points, payouts));
    expect(byID['e3'].isTied).toBe(true);
    expect(byID['e3'].payoutCents).toBe(0);
  });

  it('all entries tied results in even payout split', () => {
    const e1 = makeEntry({ id: 'e1', createdAt: '2026-01-01T00:00:04Z' });
    const e2 = makeEntry({ id: 'e2', createdAt: '2026-01-01T00:00:03Z' });
    const e3 = makeEntry({ id: 'e3', createdAt: '2026-01-01T00:00:02Z' });
    const e4 = makeEntry({ id: 'e4', createdAt: '2026-01-01T00:00:01Z' });
    const points = { e1: 10, e2: 10, e3: 10, e4: 10 };
    const payouts: DashboardPayout[] = [
      { position: 1, amountCents: 400 },
      { position: 2, amountCents: 200 },
      { position: 3, amountCents: 200 },
      { position: 4, amountCents: 200 },
    ];
    const byID = standingsByID(computeStandings([e1, e2, e3, e4], points, payouts));
    expect(byID['e1'].payoutCents).toBe(250);
  });

  it('single entry gets full payout with no tie', () => {
    const e1 = makeEntry({ id: 'e1', createdAt: '2026-01-01T00:00:01Z' });
    const points = { e1: 10 };
    const payouts: DashboardPayout[] = [{ position: 1, amountCents: 500 }];
    const byID = standingsByID(computeStandings([e1], points, payouts));
    expect(byID['e1'].isTied).toBe(false);
    expect(byID['e1'].payoutCents).toBe(500);
  });

  it('tie group spanning payout boundary pools only defined payouts', () => {
    const e1 = makeEntry({ id: 'e1', createdAt: '2026-01-01T00:00:04Z' });
    const e2 = makeEntry({ id: 'e2', createdAt: '2026-01-01T00:00:03Z' });
    const e3 = makeEntry({ id: 'e3', createdAt: '2026-01-01T00:00:02Z' });
    const e4 = makeEntry({ id: 'e4', createdAt: '2026-01-01T00:00:01Z' });
    const points = { e1: 20, e2: 10, e3: 10, e4: 10 };
    const payouts: DashboardPayout[] = [
      { position: 1, amountCents: 300 },
      { position: 2, amountCents: 150 },
      { position: 3, amountCents: 50 },
    ];
    const byID = standingsByID(computeStandings([e1, e2, e3, e4], points, payouts));
    expect(byID['e2'].payoutCents).toBe(67);
    expect(byID['e4'].payoutCents).toBe(66);
  });

  it('non-tied entry after tie group gets correct finish position', () => {
    const e1 = makeEntry({ id: 'e1', createdAt: '2026-01-01T00:00:03Z' });
    const e2 = makeEntry({ id: 'e2', createdAt: '2026-01-01T00:00:02Z' });
    const e3 = makeEntry({ id: 'e3', createdAt: '2026-01-01T00:00:01Z' });
    const points = { e1: 20, e2: 20, e3: 10 };
    const byID = standingsByID(computeStandings([e1, e2, e3], points, []));
    expect(byID['e3'].finishPosition).toBe(3);
  });
});

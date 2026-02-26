import { describe, it, expect } from 'vitest';
import type { InvestmentSlot, TeamWithSchool } from './useInvesting';
import {
  createEmptySlot,
  initializeSlotsFromInvestments,
  deriveInvestmentsByTeamId,
  compactSlots,
  deriveTeamOptions,
  deriveUsedTeamIds,
} from './useInvesting';

// ---------------------------------------------------------------------------
// createEmptySlot
// ---------------------------------------------------------------------------
describe('createEmptySlot', () => {
  it('returns empty teamId', () => {
    const slot = createEmptySlot();
    expect(slot.teamId).toBe('');
  });

  it('returns empty searchText', () => {
    const slot = createEmptySlot();
    expect(slot.searchText).toBe('');
  });

  it('returns zero investmentAmount', () => {
    const slot = createEmptySlot();
    expect(slot.investmentAmount).toBe(0);
  });
});

// ---------------------------------------------------------------------------
// initializeSlotsFromInvestments
// ---------------------------------------------------------------------------
describe('initializeSlotsFromInvestments', () => {
  const teams: TeamWithSchool[] = [
    { id: 't1', schoolId: 's1', tournamentId: 'tourn-1', seed: 1, region: 'East', byes: 0, wins: 0, isEliminated: false, createdAt: '', updatedAt: '', school: { id: 's1', name: 'Duke' } },
    { id: 't2', schoolId: 's2', tournamentId: 'tourn-1', seed: 2, region: 'East', byes: 0, wins: 0, isEliminated: false, createdAt: '', updatedAt: '', school: { id: 's2', name: 'UNC' } },
  ];

  it('creates filled slots from existing investments', () => {
    const investments = { t1: 20, t2: 15 };
    const slots = initializeSlotsFromInvestments(investments, teams, 5);
    expect(slots.filter((s) => s.teamId !== '').length).toBe(2);
  });

  it('sets searchText to school name for filled slots', () => {
    const investments = { t1: 20 };
    const slots = initializeSlotsFromInvestments(investments, teams, 5);
    expect(slots[0].searchText).toBe('Duke');
  });

  it('sets investmentAmount for filled slots', () => {
    const investments = { t1: 20 };
    const slots = initializeSlotsFromInvestments(investments, teams, 5);
    expect(slots[0].investmentAmount).toBe(20);
  });

  it('pads remaining slots with empty slots up to maxTeams', () => {
    const investments = { t1: 20 };
    const slots = initializeSlotsFromInvestments(investments, teams, 5);
    expect(slots).toHaveLength(5);
  });

  it('returns only filled slots when filled count equals maxTeams', () => {
    const investments = { t1: 20, t2: 15 };
    const slots = initializeSlotsFromInvestments(investments, teams, 2);
    expect(slots).toHaveLength(2);
  });

  it('handles empty initial investments', () => {
    const slots = initializeSlotsFromInvestments({}, teams, 3);
    expect(slots.every((s) => s.teamId === '')).toBe(true);
  });

  it('uses empty searchText when team has no school', () => {
    const teamsNoSchool: TeamWithSchool[] = [
      { id: 't1', schoolId: 's1', tournamentId: 'tourn-1', seed: 1, region: 'East', byes: 0, wins: 0, isEliminated: false, createdAt: '', updatedAt: '' },
    ];
    const slots = initializeSlotsFromInvestments({ t1: 10 }, teamsNoSchool, 3);
    expect(slots[0].searchText).toBe('');
  });
});

// ---------------------------------------------------------------------------
// deriveInvestmentsByTeamId
// ---------------------------------------------------------------------------
describe('deriveInvestmentsByTeamId', () => {
  it('returns mapping of teamId to investmentAmount for filled slots', () => {
    const slots: InvestmentSlot[] = [
      { teamId: 't1', searchText: 'Duke', investmentAmount: 20 },
      { teamId: 't2', searchText: 'UNC', investmentAmount: 15 },
    ];
    const result = deriveInvestmentsByTeamId(slots);
    expect(result).toEqual({ t1: 20, t2: 15 });
  });

  it('skips slots with empty teamId', () => {
    const slots: InvestmentSlot[] = [
      { teamId: '', searchText: '', investmentAmount: 0 },
      { teamId: 't1', searchText: 'Duke', investmentAmount: 10 },
    ];
    const result = deriveInvestmentsByTeamId(slots);
    expect(Object.keys(result)).toHaveLength(1);
  });

  it('skips slots with zero investmentAmount', () => {
    const slots: InvestmentSlot[] = [
      { teamId: 't1', searchText: 'Duke', investmentAmount: 0 },
    ];
    const result = deriveInvestmentsByTeamId(slots);
    expect(Object.keys(result)).toHaveLength(0);
  });

  it('returns empty object for all-empty slots', () => {
    const slots: InvestmentSlot[] = [createEmptySlot(), createEmptySlot()];
    const result = deriveInvestmentsByTeamId(slots);
    expect(result).toEqual({});
  });
});

// ---------------------------------------------------------------------------
// compactSlots
// ---------------------------------------------------------------------------
describe('compactSlots', () => {
  it('moves filled slots before empty slots', () => {
    const slots: InvestmentSlot[] = [
      createEmptySlot(),
      { teamId: 't1', searchText: 'Duke', investmentAmount: 20 },
      createEmptySlot(),
    ];
    const result = compactSlots(slots);
    expect(result[0].teamId).toBe('t1');
  });

  it('preserves total slot count', () => {
    const slots: InvestmentSlot[] = [
      createEmptySlot(),
      { teamId: 't1', searchText: 'Duke', investmentAmount: 20 },
      createEmptySlot(),
    ];
    const result = compactSlots(slots);
    expect(result).toHaveLength(3);
  });

  it('returns all-empty array unchanged in length', () => {
    const slots = [createEmptySlot(), createEmptySlot()];
    const result = compactSlots(slots);
    expect(result.every((s) => s.teamId === '')).toBe(true);
  });

  it('maintains relative order of filled slots', () => {
    const slots: InvestmentSlot[] = [
      { teamId: 't2', searchText: 'UNC', investmentAmount: 15 },
      createEmptySlot(),
      { teamId: 't1', searchText: 'Duke', investmentAmount: 20 },
    ];
    const result = compactSlots(slots);
    expect(result[0].teamId).toBe('t2');
    expect(result[1].teamId).toBe('t1');
  });
});

// ---------------------------------------------------------------------------
// deriveTeamOptions
// ---------------------------------------------------------------------------
describe('deriveTeamOptions', () => {
  it('maps teams to options with label from school name', () => {
    const teams: TeamWithSchool[] = [
      { id: 't1', schoolId: 's1', tournamentId: 'tourn-1', seed: 1, region: 'East', byes: 0, wins: 0, isEliminated: false, createdAt: '', updatedAt: '', school: { id: 's1', name: 'Duke' } },
    ];
    const options = deriveTeamOptions(teams);
    expect(options[0].label).toBe('Duke');
  });

  it('uses "Unknown" label when school is missing', () => {
    const teams: TeamWithSchool[] = [
      { id: 't1', schoolId: 's1', tournamentId: 'tourn-1', seed: 1, region: 'East', byes: 0, wins: 0, isEliminated: false, createdAt: '', updatedAt: '' },
    ];
    const options = deriveTeamOptions(teams);
    expect(options[0].label).toBe('Unknown');
  });

  it('sorts by seed ascending', () => {
    const teams: TeamWithSchool[] = [
      { id: 't2', schoolId: 's2', tournamentId: 'tourn-1', seed: 3, region: 'East', byes: 0, wins: 0, isEliminated: false, createdAt: '', updatedAt: '', school: { id: 's2', name: 'UNC' } },
      { id: 't1', schoolId: 's1', tournamentId: 'tourn-1', seed: 1, region: 'East', byes: 0, wins: 0, isEliminated: false, createdAt: '', updatedAt: '', school: { id: 's1', name: 'Duke' } },
    ];
    const options = deriveTeamOptions(teams);
    expect(options[0].seed).toBe(1);
    expect(options[1].seed).toBe(3);
  });

  it('sorts alphabetically by label for same seed', () => {
    const teams: TeamWithSchool[] = [
      { id: 't2', schoolId: 's2', tournamentId: 'tourn-1', seed: 1, region: 'East', byes: 0, wins: 0, isEliminated: false, createdAt: '', updatedAt: '', school: { id: 's2', name: 'UNC' } },
      { id: 't1', schoolId: 's1', tournamentId: 'tourn-1', seed: 1, region: 'West', byes: 0, wins: 0, isEliminated: false, createdAt: '', updatedAt: '', school: { id: 's1', name: 'Duke' } },
    ];
    const options = deriveTeamOptions(teams);
    expect(options[0].label).toBe('Duke');
    expect(options[1].label).toBe('UNC');
  });

  it('returns empty array for no teams', () => {
    const options = deriveTeamOptions([]);
    expect(options).toEqual([]);
  });

  it('includes region from team data', () => {
    const teams: TeamWithSchool[] = [
      { id: 't1', schoolId: 's1', tournamentId: 'tourn-1', seed: 1, region: 'West', byes: 0, wins: 0, isEliminated: false, createdAt: '', updatedAt: '', school: { id: 's1', name: 'Duke' } },
    ];
    const options = deriveTeamOptions(teams);
    expect(options[0].region).toBe('West');
  });
});

// ---------------------------------------------------------------------------
// deriveUsedTeamIds
// ---------------------------------------------------------------------------
describe('deriveUsedTeamIds', () => {
  it('returns set of team IDs from filled slots', () => {
    const slots: InvestmentSlot[] = [
      { teamId: 't1', searchText: 'Duke', investmentAmount: 20 },
      { teamId: 't2', searchText: 'UNC', investmentAmount: 15 },
    ];
    const used = deriveUsedTeamIds(slots);
    expect(used.size).toBe(2);
  });

  it('excludes empty slots', () => {
    const slots: InvestmentSlot[] = [
      { teamId: 't1', searchText: 'Duke', investmentAmount: 20 },
      createEmptySlot(),
    ];
    const used = deriveUsedTeamIds(slots);
    expect(used.has('t1')).toBe(true);
    expect(used.size).toBe(1);
  });

  it('returns empty set when all slots are empty', () => {
    const slots = [createEmptySlot(), createEmptySlot()];
    const used = deriveUsedTeamIds(slots);
    expect(used.size).toBe(0);
  });
});

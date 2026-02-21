import { describe, it, expect } from 'vitest';
import {
  createEmptySlot,
  initializeSlotsFromBids,
  deriveBidsByTeamId,
  compactSlots,
  deriveTeamOptions,
  deriveUsedTeamIds,
} from './useBidding';
import type { BidSlot, TeamWithSchool } from './useBidding';

// ---------------------------------------------------------------------------
// The hook orchestrates React state (useState, useMemo, useCallback) around
// pure slot-manipulation logic. Budget/validation is already tested in
// bidValidation.test.ts (18 tests). This file tests the pure functions that
// were extracted from useBidding: slot creation, initialization from existing
// bids, bid-to-slot derivation, compaction, team options, and used team IDs.
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// Test helpers -- factory functions for building fixture data
// ---------------------------------------------------------------------------

function makeSlot(overrides: Partial<BidSlot> = {}): BidSlot {
  return {
    teamId: '',
    searchText: '',
    bidAmount: 0,
    ...overrides,
  };
}

function makeTeam(overrides: Partial<TeamWithSchool> & { id: string }): TeamWithSchool {
  return {
    schoolId: 's1',
    tournamentId: 'tourn-1',
    seed: 1,
    region: 'East',
    byes: 0,
    wins: 0,
    eliminated: false,
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
    ...overrides,
  };
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe('createEmptySlot', () => {
  it('returns empty teamId', () => {
    // GIVEN no arguments
    // WHEN creating an empty slot
    const slot = createEmptySlot();

    // THEN teamId is an empty string
    expect(slot.teamId).toBe('');
  });

  it('returns empty searchText', () => {
    // GIVEN no arguments
    // WHEN creating an empty slot
    const slot = createEmptySlot();

    // THEN searchText is an empty string
    expect(slot.searchText).toBe('');
  });

  it('returns zero bidAmount', () => {
    // GIVEN no arguments
    // WHEN creating an empty slot
    const slot = createEmptySlot();

    // THEN bidAmount is 0
    expect(slot.bidAmount).toBe(0);
  });

  it('returns a new object each time', () => {
    // GIVEN two calls to createEmptySlot
    // WHEN creating two empty slots
    const slotA = createEmptySlot();
    const slotB = createEmptySlot();

    // THEN they are not the same reference
    expect(slotA).not.toBe(slotB);
  });
});

describe('initializeSlotsFromBids', () => {
  it('creates filled slots from existing bids', () => {
    // GIVEN one existing bid for team t1 at 25 points
    const teams = [makeTeam({ id: 't1', school: { id: 's1', name: 'Duke' } })];
    const initialBids = { t1: 25 };

    // WHEN initializing slots with maxTeams=3
    const slots = initializeSlotsFromBids(initialBids, teams, 3);

    // THEN the first slot has teamId t1 with bidAmount 25
    expect(slots[0].teamId).toBe('t1');
  });

  it('populates searchText from team school name', () => {
    // GIVEN a team with school name "Duke"
    const teams = [makeTeam({ id: 't1', school: { id: 's1', name: 'Duke' } })];
    const initialBids = { t1: 10 };

    // WHEN initializing slots
    const slots = initializeSlotsFromBids(initialBids, teams, 3);

    // THEN the filled slot searchText is the school name
    expect(slots[0].searchText).toBe('Duke');
  });

  it('preserves bid amount from initial bids', () => {
    // GIVEN an initial bid of 42 points
    const teams = [makeTeam({ id: 't1', school: { id: 's1', name: 'Duke' } })];
    const initialBids = { t1: 42 };

    // WHEN initializing slots
    const slots = initializeSlotsFromBids(initialBids, teams, 3);

    // THEN the bid amount is preserved
    expect(slots[0].bidAmount).toBe(42);
  });

  it('uses empty searchText when team is not found in teams list', () => {
    // GIVEN a bid for a team that does not exist in the teams array
    const teams: TeamWithSchool[] = [];
    const initialBids = { 'unknown-team': 15 };

    // WHEN initializing slots
    const slots = initializeSlotsFromBids(initialBids, teams, 3);

    // THEN searchText falls back to empty string
    expect(slots[0].searchText).toBe('');
  });

  it('uses empty searchText when team has no school', () => {
    // GIVEN a team with no school attached
    const teams = [makeTeam({ id: 't1' })];
    const initialBids = { t1: 10 };

    // WHEN initializing slots
    const slots = initializeSlotsFromBids(initialBids, teams, 3);

    // THEN searchText falls back to empty string
    expect(slots[0].searchText).toBe('');
  });

  it('pads remaining slots with empty slots to fill maxTeams', () => {
    // GIVEN 1 existing bid and maxTeams=4
    const teams = [makeTeam({ id: 't1', school: { id: 's1', name: 'Duke' } })];
    const initialBids = { t1: 20 };

    // WHEN initializing slots
    const slots = initializeSlotsFromBids(initialBids, teams, 4);

    // THEN total slot count equals maxTeams
    expect(slots).toHaveLength(4);
  });

  it('creates all empty slots when there are no initial bids', () => {
    // GIVEN no initial bids and maxTeams=3
    const initialBids: Record<string, number> = {};

    // WHEN initializing slots
    const slots = initializeSlotsFromBids(initialBids, [], 3);

    // THEN all 3 slots are empty
    expect(slots.every((s) => s.teamId === '')).toBe(true);
  });

  it('places filled slots before empty slots', () => {
    // GIVEN 2 initial bids and maxTeams=5
    const teams = [
      makeTeam({ id: 't1', school: { id: 's1', name: 'Duke' } }),
      makeTeam({ id: 't2', school: { id: 's2', name: 'UNC' } }),
    ];
    const initialBids = { t1: 10, t2: 20 };

    // WHEN initializing slots
    const slots = initializeSlotsFromBids(initialBids, teams, 5);

    // THEN the first two slots are filled, followed by three empty
    expect(slots[2].teamId).toBe('');
  });

  it('creates no empty slots when initial bids fill maxTeams', () => {
    // GIVEN 2 initial bids and maxTeams=2
    const teams = [
      makeTeam({ id: 't1', school: { id: 's1', name: 'Duke' } }),
      makeTeam({ id: 't2', school: { id: 's2', name: 'UNC' } }),
    ];
    const initialBids = { t1: 10, t2: 20 };

    // WHEN initializing slots
    const slots = initializeSlotsFromBids(initialBids, teams, 2);

    // THEN all slots are filled (no empty padding)
    expect(slots.every((s) => s.teamId !== '')).toBe(true);
  });

  it('handles maxTeams of zero', () => {
    // GIVEN maxTeams=0 and no bids
    // WHEN initializing slots
    const slots = initializeSlotsFromBids({}, [], 0);

    // THEN the result is an empty array
    expect(slots).toHaveLength(0);
  });
});

describe('deriveBidsByTeamId', () => {
  it('returns empty object when all slots are empty', () => {
    // GIVEN slots with no team selections
    const slots = [createEmptySlot(), createEmptySlot()];

    // WHEN deriving bids by team ID
    const result = deriveBidsByTeamId(slots);

    // THEN the result is an empty object
    expect(result).toEqual({});
  });

  it('extracts teamId and bidAmount for filled slots', () => {
    // GIVEN a slot with teamId and bidAmount > 0
    const slots = [makeSlot({ teamId: 't1', bidAmount: 25 })];

    // WHEN deriving bids by team ID
    const result = deriveBidsByTeamId(slots);

    // THEN the result maps teamId to bidAmount
    expect(result).toEqual({ t1: 25 });
  });

  it('excludes slots with empty teamId', () => {
    // GIVEN a slot with no teamId but a bidAmount
    const slots = [makeSlot({ teamId: '', bidAmount: 10 })];

    // WHEN deriving bids by team ID
    const result = deriveBidsByTeamId(slots);

    // THEN the slot is excluded
    expect(result).toEqual({});
  });

  it('excludes slots with zero bidAmount', () => {
    // GIVEN a slot with a teamId but bidAmount of 0
    const slots = [makeSlot({ teamId: 't1', bidAmount: 0 })];

    // WHEN deriving bids by team ID
    const result = deriveBidsByTeamId(slots);

    // THEN the slot is excluded
    expect(result).toEqual({});
  });

  it('collects multiple filled slots', () => {
    // GIVEN three slots: two filled, one empty
    const slots = [
      makeSlot({ teamId: 't1', bidAmount: 10 }),
      createEmptySlot(),
      makeSlot({ teamId: 't2', bidAmount: 30 }),
    ];

    // WHEN deriving bids by team ID
    const result = deriveBidsByTeamId(slots);

    // THEN both filled slots are included
    expect(Object.keys(result)).toHaveLength(2);
  });

  it('uses the last bid when the same teamId appears in multiple slots', () => {
    // GIVEN two slots with the same teamId but different bids
    const slots = [
      makeSlot({ teamId: 't1', bidAmount: 10 }),
      makeSlot({ teamId: 't1', bidAmount: 30 }),
    ];

    // WHEN deriving bids by team ID
    const result = deriveBidsByTeamId(slots);

    // THEN the last bid wins
    expect(result['t1']).toBe(30);
  });
});

describe('compactSlots', () => {
  it('returns same order when already compacted', () => {
    // GIVEN slots where filled come before empty
    const slots = [
      makeSlot({ teamId: 't1', bidAmount: 10 }),
      makeSlot({ teamId: 't2', bidAmount: 20 }),
      createEmptySlot(),
    ];

    // WHEN compacting
    const result = compactSlots(slots);

    // THEN order is unchanged
    expect(result.map((s) => s.teamId)).toEqual(['t1', 't2', '']);
  });

  it('moves filled slots to the front', () => {
    // GIVEN an empty slot between two filled slots
    const slots = [
      makeSlot({ teamId: 't1', bidAmount: 10 }),
      createEmptySlot(),
      makeSlot({ teamId: 't2', bidAmount: 20 }),
    ];

    // WHEN compacting
    const result = compactSlots(slots);

    // THEN filled slots are first, empty last
    expect(result[0].teamId).toBe('t1');
  });

  it('preserves total slot count', () => {
    // GIVEN 4 slots with a gap
    const slots = [
      createEmptySlot(),
      makeSlot({ teamId: 't1', bidAmount: 10 }),
      createEmptySlot(),
      makeSlot({ teamId: 't2', bidAmount: 20 }),
    ];

    // WHEN compacting
    const result = compactSlots(slots);

    // THEN total count is preserved
    expect(result).toHaveLength(4);
  });

  it('moves all empty slots to the end', () => {
    // GIVEN slots with empties scattered
    const slots = [
      createEmptySlot(),
      makeSlot({ teamId: 't1', bidAmount: 10 }),
      createEmptySlot(),
      makeSlot({ teamId: 't2', bidAmount: 20 }),
    ];

    // WHEN compacting
    const result = compactSlots(slots);

    // THEN the last two slots are empty
    expect(result[2].teamId).toBe('');
  });

  it('handles all empty slots', () => {
    // GIVEN only empty slots
    const slots = [createEmptySlot(), createEmptySlot(), createEmptySlot()];

    // WHEN compacting
    const result = compactSlots(slots);

    // THEN all slots remain empty
    expect(result.every((s) => s.teamId === '')).toBe(true);
  });

  it('handles all filled slots', () => {
    // GIVEN only filled slots
    const slots = [
      makeSlot({ teamId: 't1', bidAmount: 10 }),
      makeSlot({ teamId: 't2', bidAmount: 20 }),
    ];

    // WHEN compacting
    const result = compactSlots(slots);

    // THEN all slots remain filled
    expect(result.every((s) => s.teamId !== '')).toBe(true);
  });

  it('preserves bid amounts during compaction', () => {
    // GIVEN an empty slot between two filled slots with specific bids
    const slots = [
      createEmptySlot(),
      makeSlot({ teamId: 't1', searchText: 'Duke', bidAmount: 42 }),
    ];

    // WHEN compacting
    const result = compactSlots(slots);

    // THEN the filled slot retains its bidAmount
    expect(result[0].bidAmount).toBe(42);
  });

  it('preserves searchText during compaction', () => {
    // GIVEN a filled slot with searchText behind an empty slot
    const slots = [
      createEmptySlot(),
      makeSlot({ teamId: 't1', searchText: 'Duke', bidAmount: 10 }),
    ];

    // WHEN compacting
    const result = compactSlots(slots);

    // THEN the filled slot retains its searchText
    expect(result[0].searchText).toBe('Duke');
  });

  it('handles empty array', () => {
    // GIVEN no slots
    // WHEN compacting
    const result = compactSlots([]);

    // THEN result is empty
    expect(result).toHaveLength(0);
  });
});

describe('deriveTeamOptions', () => {
  it('returns empty array for no teams', () => {
    // GIVEN no teams
    // WHEN deriving team options
    const options = deriveTeamOptions([]);

    // THEN result is empty
    expect(options).toEqual([]);
  });

  it('maps team id to option id', () => {
    // GIVEN a team with id "t1"
    const teams = [makeTeam({ id: 't1', school: { id: 's1', name: 'Duke' } })];

    // WHEN deriving team options
    const options = deriveTeamOptions(teams);

    // THEN the option id matches the team id
    expect(options[0].id).toBe('t1');
  });

  it('maps school name to option label', () => {
    // GIVEN a team with school name "Duke"
    const teams = [makeTeam({ id: 't1', school: { id: 's1', name: 'Duke' } })];

    // WHEN deriving team options
    const options = deriveTeamOptions(teams);

    // THEN the option label is the school name
    expect(options[0].label).toBe('Duke');
  });

  it('uses "Unknown" when team has no school', () => {
    // GIVEN a team with no school
    const teams = [makeTeam({ id: 't1' })];

    // WHEN deriving team options
    const options = deriveTeamOptions(teams);

    // THEN label falls back to "Unknown"
    expect(options[0].label).toBe('Unknown');
  });

  it('includes seed in option', () => {
    // GIVEN a team with seed 3
    const teams = [makeTeam({ id: 't1', seed: 3, school: { id: 's1', name: 'Duke' } })];

    // WHEN deriving team options
    const options = deriveTeamOptions(teams);

    // THEN the option includes the seed
    expect(options[0].seed).toBe(3);
  });

  it('includes region in option', () => {
    // GIVEN a team in the "West" region
    const teams = [makeTeam({ id: 't1', region: 'West', school: { id: 's1', name: 'Duke' } })];

    // WHEN deriving team options
    const options = deriveTeamOptions(teams);

    // THEN the option includes the region
    expect(options[0].region).toBe('West');
  });

  it('sorts by seed ascending as primary sort', () => {
    // GIVEN teams with seeds 3 and 1
    const teams = [
      makeTeam({ id: 't1', seed: 3, school: { id: 's1', name: 'Auburn' } }),
      makeTeam({ id: 't2', seed: 1, school: { id: 's2', name: 'Duke' } }),
    ];

    // WHEN deriving team options
    const options = deriveTeamOptions(teams);

    // THEN lower seed comes first
    expect(options[0].seed).toBe(1);
  });

  it('sorts by name alphabetically when seeds are equal', () => {
    // GIVEN two teams with the same seed but different names
    const teams = [
      makeTeam({ id: 't1', seed: 1, school: { id: 's1', name: 'UNC' } }),
      makeTeam({ id: 't2', seed: 1, school: { id: 's2', name: 'Duke' } }),
    ];

    // WHEN deriving team options
    const options = deriveTeamOptions(teams);

    // THEN alphabetically earlier name comes first
    expect(options[0].label).toBe('Duke');
  });

  it('sorts teams with different seeds before considering name', () => {
    // GIVEN teams where a later-alphabet name has a lower seed
    const teams = [
      makeTeam({ id: 't1', seed: 8, school: { id: 's1', name: 'Alabama' } }),
      makeTeam({ id: 't2', seed: 2, school: { id: 's2', name: 'Villanova' } }),
    ];

    // WHEN deriving team options
    const options = deriveTeamOptions(teams);

    // THEN lower seed wins regardless of name
    expect(options[0].label).toBe('Villanova');
  });
});

describe('deriveUsedTeamIds', () => {
  it('returns empty set for empty slots', () => {
    // GIVEN no slots
    // WHEN deriving used team IDs
    const used = deriveUsedTeamIds([]);

    // THEN the set is empty
    expect(used.size).toBe(0);
  });

  it('returns empty set when all slots are empty', () => {
    // GIVEN only empty slots
    const slots = [createEmptySlot(), createEmptySlot()];

    // WHEN deriving used team IDs
    const used = deriveUsedTeamIds(slots);

    // THEN the set is empty
    expect(used.size).toBe(0);
  });

  it('includes team IDs from filled slots', () => {
    // GIVEN a slot with teamId "t1"
    const slots = [makeSlot({ teamId: 't1', bidAmount: 10 })];

    // WHEN deriving used team IDs
    const used = deriveUsedTeamIds(slots);

    // THEN t1 is in the set
    expect(used.has('t1')).toBe(true);
  });

  it('excludes empty teamId strings', () => {
    // GIVEN a mix of filled and empty slots
    const slots = [
      makeSlot({ teamId: 't1', bidAmount: 10 }),
      createEmptySlot(),
    ];

    // WHEN deriving used team IDs
    const used = deriveUsedTeamIds(slots);

    // THEN the set only contains the filled team ID
    expect(used.size).toBe(1);
  });

  it('includes team IDs even when bidAmount is zero', () => {
    // GIVEN a slot with a teamId but zero bid (team is selected but not yet bid on)
    const slots = [makeSlot({ teamId: 't1', bidAmount: 0 })];

    // WHEN deriving used team IDs
    const used = deriveUsedTeamIds(slots);

    // THEN the team is still considered "used"
    expect(used.has('t1')).toBe(true);
  });

  it('deduplicates team IDs', () => {
    // GIVEN two slots with the same teamId
    const slots = [
      makeSlot({ teamId: 't1', bidAmount: 10 }),
      makeSlot({ teamId: 't1', bidAmount: 20 }),
    ];

    // WHEN deriving used team IDs
    const used = deriveUsedTeamIds(slots);

    // THEN the set has only one entry
    expect(used.size).toBe(1);
  });

  it('collects all distinct team IDs', () => {
    // GIVEN three slots with three different team IDs
    const slots = [
      makeSlot({ teamId: 't1', bidAmount: 10 }),
      makeSlot({ teamId: 't2', bidAmount: 20 }),
      makeSlot({ teamId: 't3', bidAmount: 30 }),
    ];

    // WHEN deriving used team IDs
    const used = deriveUsedTeamIds(slots);

    // THEN all three are in the set
    expect(used.size).toBe(3);
  });
});

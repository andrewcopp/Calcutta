import { describe, it, expect } from 'vitest';
import {
  createEmptyRegion,
  createInitialRegions,
  getRegionList,
  createRegionsFromTeams,
  deriveUsedSchoolIds,
  deriveValidationStats,
  deriveSlotValidation,
  applyUpdateSlot,
  applyAddPlayIn,
  applyRemovePlayIn,
  applySlotBlur,
  collectTeamsForSubmission,
} from './useTeamSetupForm';
import type { RegionState } from './useTeamSetupForm';
import type { Tournament, TournamentTeam } from '../schemas/tournament';
import type { School } from '../schemas/school';

// ---------------------------------------------------------------------------
// The hook orchestrates React state (useState, useMemo, useCallback) around
// pure region-manipulation and validation logic. This file tests the extracted
// pure functions: region creation, team population, duplicate detection,
// play-in validation, slot state transitions, and submission collection.
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// Test helpers -- factory functions for building fixture data
// ---------------------------------------------------------------------------

function makeTournament(overrides: Partial<Tournament> = {}): Tournament {
  return {
    id: 'tourn-1',
    name: 'NCAA 2026',
    rounds: 6,
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
    ...overrides,
  };
}

function makeTeam(overrides: Partial<TournamentTeam> & { id: string; schoolId: string }): TournamentTeam {
  return {
    tournamentId: 'tourn-1',
    seed: 1,
    region: 'East',
    byes: 0,
    wins: 0,
    isEliminated: false,
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
    ...overrides,
  };
}

function makeSchool(overrides: Partial<School> & { id: string }): School {
  return {
    name: 'Unknown',
    ...overrides,
  };
}

/** Build a minimal regions object with a single region containing specific slots. */
function buildRegions(
  regionName: string,
  seedSlots: Record<number, { schoolId: string; searchText: string }[]>,
): Record<string, RegionState> {
  const base = createInitialRegions([regionName]);
  for (const [seed, slots] of Object.entries(seedSlots)) {
    base[regionName][Number(seed)] = slots;
  }
  return base;
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe('createEmptyRegion', () => {
  it('creates 16 seed entries', () => {
    // GIVEN no arguments
    // WHEN creating an empty region
    const region = createEmptyRegion();

    // THEN there are entries for seeds 1 through 16
    expect(Object.keys(region)).toHaveLength(16);
  });

  it('initializes each seed with one empty slot', () => {
    // GIVEN no arguments
    // WHEN creating an empty region
    const region = createEmptyRegion();

    // THEN seed 1 has exactly one slot
    expect(region[1]).toHaveLength(1);
  });

  it('sets empty schoolId on each slot', () => {
    // GIVEN no arguments
    // WHEN creating an empty region
    const region = createEmptyRegion();

    // THEN seed 8's slot has an empty schoolId
    expect(region[8][0].schoolId).toBe('');
  });

  it('sets empty searchText on each slot', () => {
    // GIVEN no arguments
    // WHEN creating an empty region
    const region = createEmptyRegion();

    // THEN seed 16's slot has empty searchText
    expect(region[16][0].searchText).toBe('');
  });

  it('returns a new object each time', () => {
    // GIVEN two calls
    // WHEN creating two empty regions
    const a = createEmptyRegion();
    const b = createEmptyRegion();

    // THEN they are not the same reference
    expect(a).not.toBe(b);
  });

  it('contains keys for exactly seeds 1 through 16', () => {
    // GIVEN no arguments
    // WHEN creating an empty region
    const region = createEmptyRegion();

    // THEN the numeric keys are 1 through 16
    expect(
      Object.keys(region)
        .map(Number)
        .sort((a, b) => a - b),
    ).toEqual([1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16]);
  });
});

describe('createInitialRegions', () => {
  it('creates an entry for each region name', () => {
    // GIVEN two region names
    // WHEN creating initial regions
    const regions = createInitialRegions(['East', 'West']);

    // THEN the result has two region keys
    expect(Object.keys(regions)).toHaveLength(2);
  });

  it('initializes each region with 16 seeds', () => {
    // GIVEN one region name
    // WHEN creating initial regions
    const regions = createInitialRegions(['South']);

    // THEN the region has 16 seed entries
    expect(Object.keys(regions['South'])).toHaveLength(16);
  });

  it('returns empty object for no region names', () => {
    // GIVEN an empty list
    // WHEN creating initial regions
    const regions = createInitialRegions([]);

    // THEN the result is empty
    expect(Object.keys(regions)).toHaveLength(0);
  });

  it('creates separate region state objects for each name', () => {
    // GIVEN two region names
    // WHEN creating initial regions
    const regions = createInitialRegions(['East', 'West']);

    // THEN each region has its own state object (not shared references)
    expect(regions['East']).not.toBe(regions['West']);
  });
});

describe('getRegionList', () => {
  it('returns custom region names from tournament', () => {
    // GIVEN a tournament with custom Final Four region names
    const tournament = makeTournament({
      finalFourTopLeft: 'Albany',
      finalFourBottomLeft: 'Denver',
      finalFourTopRight: 'Memphis',
      finalFourBottomRight: 'Portland',
    });

    // WHEN getting the region list
    const result = getRegionList(tournament);

    // THEN the custom names are returned
    expect(result).toEqual(['Albany', 'Denver', 'Memphis', 'Portland']);
  });

  it('falls back to default names when tournament has no region names', () => {
    // GIVEN a tournament with no Final Four region names set
    const tournament = makeTournament();

    // WHEN getting the region list
    const result = getRegionList(tournament);

    // THEN the defaults are used
    expect(result).toEqual(['East', 'West', 'South', 'Midwest']);
  });

  it('uses default for only the missing region names', () => {
    // GIVEN a tournament with only finalFourTopLeft set
    const tournament = makeTournament({ finalFourTopLeft: 'Custom' });

    // WHEN getting the region list
    const result = getRegionList(tournament);

    // THEN the first region is custom, the rest are defaults
    expect(result[0]).toBe('Custom');
  });

  it('preserves default for bottomRight when only others are set', () => {
    // GIVEN a tournament with three region names set but bottomRight missing
    const tournament = makeTournament({
      finalFourTopLeft: 'A',
      finalFourBottomLeft: 'B',
      finalFourTopRight: 'C',
    });

    // WHEN getting the region list
    const result = getRegionList(tournament);

    // THEN the fourth region falls back to "Midwest"
    expect(result[3]).toBe('Midwest');
  });

  it('always returns exactly four regions', () => {
    // GIVEN any tournament
    const tournament = makeTournament();

    // WHEN getting the region list
    const result = getRegionList(tournament);

    // THEN exactly four regions are returned
    expect(result).toHaveLength(4);
  });
});

describe('createRegionsFromTeams', () => {
  it('places a team in the correct region and seed slot', () => {
    // GIVEN one team in East at seed 1
    const teams = [makeTeam({ id: 't1', schoolId: 's1', region: 'East', seed: 1 })];
    const schools = [makeSchool({ id: 's1', name: 'Duke' })];

    // WHEN creating regions from teams
    const regions = createRegionsFromTeams(['East'], teams, schools);

    // THEN seed 1 in East has the school filled
    expect(regions['East'][1][0].schoolId).toBe('s1');
  });

  it('populates searchText from the school name', () => {
    // GIVEN a team whose school is "Duke"
    const teams = [makeTeam({ id: 't1', schoolId: 's1', region: 'East', seed: 1 })];
    const schools = [makeSchool({ id: 's1', name: 'Duke' })];

    // WHEN creating regions from teams
    const regions = createRegionsFromTeams(['East'], teams, schools);

    // THEN the slot searchText is "Duke"
    expect(regions['East'][1][0].searchText).toBe('Duke');
  });

  it('uses empty searchText when school is not found', () => {
    // GIVEN a team whose schoolId has no matching school
    const teams = [makeTeam({ id: 't1', schoolId: 'missing', region: 'East', seed: 1 })];

    // WHEN creating regions from teams with no schools
    const regions = createRegionsFromTeams(['East'], teams, []);

    // THEN the slot searchText is empty
    expect(regions['East'][1][0].searchText).toBe('');
  });

  it('creates a play-in when two teams share the same seed in a region', () => {
    // GIVEN two teams at seed 11 in East
    const teams = [
      makeTeam({ id: 't1', schoolId: 's1', region: 'East', seed: 11 }),
      makeTeam({ id: 't2', schoolId: 's2', region: 'East', seed: 11 }),
    ];
    const schools = [makeSchool({ id: 's1', name: 'Arizona St' }), makeSchool({ id: 's2', name: 'Texas' })];

    // WHEN creating regions from teams
    const regions = createRegionsFromTeams(['East'], teams, schools);

    // THEN seed 11 in East has two slots (play-in)
    expect(regions['East'][11]).toHaveLength(2);
  });

  it('fills the first slot before adding a play-in slot', () => {
    // GIVEN two teams at the same seed
    const teams = [
      makeTeam({ id: 't1', schoolId: 's1', region: 'East', seed: 11 }),
      makeTeam({ id: 't2', schoolId: 's2', region: 'East', seed: 11 }),
    ];
    const schools = [makeSchool({ id: 's1', name: 'Arizona St' }), makeSchool({ id: 's2', name: 'Texas' })];

    // WHEN creating regions from teams
    const regions = createRegionsFromTeams(['East'], teams, schools);

    // THEN the first slot has the first team
    expect(regions['East'][11][0].schoolId).toBe('s1');
  });

  it('places the second team in the play-in slot', () => {
    // GIVEN two teams at the same seed
    const teams = [
      makeTeam({ id: 't1', schoolId: 's1', region: 'East', seed: 11 }),
      makeTeam({ id: 't2', schoolId: 's2', region: 'East', seed: 11 }),
    ];
    const schools = [makeSchool({ id: 's1', name: 'Arizona St' }), makeSchool({ id: 's2', name: 'Texas' })];

    // WHEN creating regions from teams
    const regions = createRegionsFromTeams(['East'], teams, schools);

    // THEN the second slot has the second team
    expect(regions['East'][11][1].schoolId).toBe('s2');
  });

  it('skips teams whose region is not in the region list', () => {
    // GIVEN a team in "South" but only "East" is in the region list
    const teams = [makeTeam({ id: 't1', schoolId: 's1', region: 'South', seed: 1 })];
    const schools = [makeSchool({ id: 's1', name: 'Duke' })];

    // WHEN creating regions from teams with only "East"
    const regions = createRegionsFromTeams(['East'], teams, schools);

    // THEN seed 1 in East remains empty
    expect(regions['East'][1][0].schoolId).toBe('');
  });

  it('leaves unfilled seeds empty', () => {
    // GIVEN a team only at seed 1
    const teams = [makeTeam({ id: 't1', schoolId: 's1', region: 'East', seed: 1 })];
    const schools = [makeSchool({ id: 's1', name: 'Duke' })];

    // WHEN creating regions from teams
    const regions = createRegionsFromTeams(['East'], teams, schools);

    // THEN seed 16 remains empty
    expect(regions['East'][16][0].schoolId).toBe('');
  });

  it('populates teams across multiple regions', () => {
    // GIVEN teams in both East and West
    const teams = [
      makeTeam({ id: 't1', schoolId: 's1', region: 'East', seed: 1 }),
      makeTeam({ id: 't2', schoolId: 's2', region: 'West', seed: 3 }),
    ];
    const schools = [makeSchool({ id: 's1', name: 'Duke' }), makeSchool({ id: 's2', name: 'Kansas' })];

    // WHEN creating regions from teams
    const regions = createRegionsFromTeams(['East', 'West'], teams, schools);

    // THEN the West team is placed at the correct seed
    expect(regions['West'][3][0].schoolId).toBe('s2');
  });
});

describe('deriveUsedSchoolIds', () => {
  it('returns empty set for empty regions', () => {
    // GIVEN regions with no filled slots
    const regions = createInitialRegions(['East']);

    // WHEN deriving used school IDs
    const used = deriveUsedSchoolIds(['East'], regions);

    // THEN the set is empty
    expect(used.size).toBe(0);
  });

  it('includes a filled school ID', () => {
    // GIVEN one filled slot
    const regions = buildRegions('East', {
      1: [{ schoolId: 's1', searchText: 'Duke' }],
    });

    // WHEN deriving used school IDs
    const used = deriveUsedSchoolIds(['East'], regions);

    // THEN s1 is in the set
    expect(used.has('s1')).toBe(true);
  });

  it('excludes empty schoolId strings', () => {
    // GIVEN a mix of filled and empty slots
    const regions = buildRegions('East', {
      1: [{ schoolId: 's1', searchText: 'Duke' }],
      2: [{ schoolId: '', searchText: '' }],
    });

    // WHEN deriving used school IDs
    const used = deriveUsedSchoolIds(['East'], regions);

    // THEN only one ID is in the set
    expect(used.size).toBe(1);
  });

  it('collects school IDs across multiple regions', () => {
    // GIVEN schools in two different regions
    const regionsEast = buildRegions('East', {
      1: [{ schoolId: 's1', searchText: 'Duke' }],
    });
    const regionsWest = buildRegions('West', {
      1: [{ schoolId: 's2', searchText: 'UNC' }],
    });
    const combined = { ...regionsEast, ...regionsWest };

    // WHEN deriving used school IDs across both regions
    const used = deriveUsedSchoolIds(['East', 'West'], combined);

    // THEN both school IDs are collected
    expect(used.size).toBe(2);
  });

  it('deduplicates the same school ID across regions', () => {
    // GIVEN the same schoolId appearing in two regions (duplicate scenario)
    const regionsEast = buildRegions('East', {
      1: [{ schoolId: 's1', searchText: 'Duke' }],
    });
    const regionsWest = buildRegions('West', {
      1: [{ schoolId: 's1', searchText: 'Duke' }],
    });
    const combined = { ...regionsEast, ...regionsWest };

    // WHEN deriving used school IDs
    const used = deriveUsedSchoolIds(['East', 'West'], combined);

    // THEN the set contains only one entry (deduped)
    expect(used.size).toBe(1);
  });

  it('skips regions not in the region list', () => {
    // GIVEN a region "South" with data but only "East" in the region list
    const regions = {
      ...buildRegions('East', {}),
      ...buildRegions('South', { 1: [{ schoolId: 's1', searchText: 'Duke' }] }),
    };

    // WHEN deriving used school IDs with only "East" in the list
    const used = deriveUsedSchoolIds(['East'], regions);

    // THEN s1 from South is not included
    expect(used.size).toBe(0);
  });
});

describe('deriveValidationStats', () => {
  it('returns zero total for empty regions', () => {
    // GIVEN empty regions
    const regions = createInitialRegions(['East']);

    // WHEN deriving validation stats
    const stats = deriveValidationStats(['East'], regions, []);

    // THEN total is 0
    expect(stats.total).toBe(0);
  });

  it('counts filled slots in total', () => {
    // GIVEN two filled slots in one region
    const regions = buildRegions('East', {
      1: [{ schoolId: 's1', searchText: 'Duke' }],
      2: [{ schoolId: 's2', searchText: 'UNC' }],
    });

    // WHEN deriving validation stats
    const stats = deriveValidationStats(['East'], regions, []);

    // THEN total is 2
    expect(stats.total).toBe(2);
  });

  it('counts play-ins when a seed has two filled slots', () => {
    // GIVEN seed 11 with two filled slots (play-in)
    const regions = buildRegions('East', {
      11: [
        { schoolId: 's1', searchText: 'Arizona St' },
        { schoolId: 's2', searchText: 'Texas' },
      ],
    });

    // WHEN deriving validation stats
    const stats = deriveValidationStats(['East'], regions, []);

    // THEN playIns is 1
    expect(stats.playIns).toBe(1);
  });

  it('does not count a play-in when only one slot is filled', () => {
    // GIVEN seed 11 with one filled and one empty slot
    const regions = buildRegions('East', {
      11: [
        { schoolId: 's1', searchText: 'Arizona St' },
        { schoolId: '', searchText: '' },
      ],
    });

    // WHEN deriving validation stats
    const stats = deriveValidationStats(['East'], regions, []);

    // THEN playIns is 0 (both must be filled to count)
    expect(stats.playIns).toBe(0);
  });

  it('tracks per-region team counts', () => {
    // GIVEN one filled slot in East and two in West
    const regions = {
      ...buildRegions('East', {
        1: [{ schoolId: 's1', searchText: 'Duke' }],
      }),
      ...buildRegions('West', {
        1: [{ schoolId: 's2', searchText: 'UNC' }],
        2: [{ schoolId: 's3', searchText: 'Kansas' }],
      }),
    };

    // WHEN deriving validation stats
    const stats = deriveValidationStats(['East', 'West'], regions, []);

    // THEN perRegion reflects the counts
    expect(stats.perRegion['West']).toBe(2);
  });

  it('detects duplicate school across regions', () => {
    // GIVEN the same schoolId in East and West
    const schools = [makeSchool({ id: 's1', name: 'Duke' })];
    const regions = {
      ...buildRegions('East', { 1: [{ schoolId: 's1', searchText: 'Duke' }] }),
      ...buildRegions('West', { 1: [{ schoolId: 's1', searchText: 'Duke' }] }),
    };

    // WHEN deriving validation stats
    const stats = deriveValidationStats(['East', 'West'], regions, schools);

    // THEN duplicates contains "Duke"
    expect(stats.duplicates).toContain('Duke');
  });

  it('returns empty duplicates when all schools are unique', () => {
    // GIVEN unique schools across regions
    const schools = [makeSchool({ id: 's1', name: 'Duke' }), makeSchool({ id: 's2', name: 'UNC' })];
    const regions = {
      ...buildRegions('East', { 1: [{ schoolId: 's1', searchText: 'Duke' }] }),
      ...buildRegions('West', { 1: [{ schoolId: 's2', searchText: 'UNC' }] }),
    };

    // WHEN deriving validation stats
    const stats = deriveValidationStats(['East', 'West'], regions, schools);

    // THEN duplicates is empty
    expect(stats.duplicates).toHaveLength(0);
  });

  it('uses schoolId as fallback when school name is not found', () => {
    // GIVEN a duplicate schoolId with no matching school in the schools list
    const regions = {
      ...buildRegions('East', { 1: [{ schoolId: 'orphan-id', searchText: '' }] }),
      ...buildRegions('West', { 1: [{ schoolId: 'orphan-id', searchText: '' }] }),
    };

    // WHEN deriving validation stats with empty schools
    const stats = deriveValidationStats(['East', 'West'], regions, []);

    // THEN the schoolId itself is used in duplicates
    expect(stats.duplicates).toContain('orphan-id');
  });

  it('does not count empty slots toward total', () => {
    // GIVEN a region with one filled and one empty slot
    const regions = buildRegions('East', {
      1: [{ schoolId: 's1', searchText: 'Duke' }],
      2: [{ schoolId: '', searchText: '' }],
    });

    // WHEN deriving validation stats
    const stats = deriveValidationStats(['East'], regions, []);

    // THEN total counts only the filled slot
    expect(stats.total).toBe(1);
  });

  it('counts play-in teams in the total', () => {
    // GIVEN a play-in seed with two filled slots
    const regions = buildRegions('East', {
      11: [
        { schoolId: 's1', searchText: 'Arizona St' },
        { schoolId: 's2', searchText: 'Texas' },
      ],
    });

    // WHEN deriving validation stats
    const stats = deriveValidationStats(['East'], regions, []);

    // THEN total includes both play-in teams
    expect(stats.total).toBe(2);
  });

  it('detects duplicate within the same region', () => {
    // GIVEN the same schoolId at two different seeds in the same region
    const schools = [makeSchool({ id: 's1', name: 'Duke' })];
    const regions = buildRegions('East', {
      1: [{ schoolId: 's1', searchText: 'Duke' }],
      2: [{ schoolId: 's1', searchText: 'Duke' }],
    });

    // WHEN deriving validation stats
    const stats = deriveValidationStats(['East'], regions, schools);

    // THEN "Duke" is detected as a duplicate
    expect(stats.duplicates).toContain('Duke');
  });
});

describe('deriveSlotValidation', () => {
  it('marks a slot with schoolId as valid', () => {
    // GIVEN a filled slot at seed 1
    const regions = buildRegions('East', {
      1: [{ schoolId: 's1', searchText: 'Duke' }],
    });

    // WHEN deriving slot validation with no flashing
    const validation = deriveSlotValidation(['East'], regions, {});

    // THEN seed 1 slot 0 is "valid"
    expect(validation['East']['1-0']).toBe('valid');
  });

  it('marks an empty non-flashing slot as none', () => {
    // GIVEN an empty slot with no flash
    const regions = buildRegions('East', {
      1: [{ schoolId: '', searchText: '' }],
    });

    // WHEN deriving slot validation with no flashing
    const validation = deriveSlotValidation(['East'], regions, {});

    // THEN seed 1 slot 0 is "none"
    expect(validation['East']['1-0']).toBe('none');
  });

  it('marks an empty flashing slot as error', () => {
    // GIVEN an empty slot with its flash key set to true
    const regions = buildRegions('East', {
      1: [{ schoolId: '', searchText: '' }],
    });
    const flashingSlots = { 'East-1-0': true };

    // WHEN deriving slot validation
    const validation = deriveSlotValidation(['East'], regions, flashingSlots);

    // THEN seed 1 slot 0 is "error"
    expect(validation['East']['1-0']).toBe('error');
  });

  it('prioritizes valid over flash for a filled slot', () => {
    // GIVEN a filled slot that also has a flash key set to true
    const regions = buildRegions('East', {
      1: [{ schoolId: 's1', searchText: 'Duke' }],
    });
    const flashingSlots = { 'East-1-0': true };

    // WHEN deriving slot validation
    const validation = deriveSlotValidation(['East'], regions, flashingSlots);

    // THEN schoolId takes priority -- the slot is "valid", not "error"
    expect(validation['East']['1-0']).toBe('valid');
  });

  it('handles play-in slots with correct key format', () => {
    // GIVEN two slots at seed 11 (play-in)
    const regions = buildRegions('East', {
      11: [
        { schoolId: 's1', searchText: 'Arizona St' },
        { schoolId: '', searchText: '' },
      ],
    });

    // WHEN deriving slot validation
    const validation = deriveSlotValidation(['East'], regions, {});

    // THEN the second slot key is "11-1"
    expect(validation['East']['11-1']).toBe('none');
  });

  it('creates entries for all 16 seeds', () => {
    // GIVEN a default empty region
    const regions = createInitialRegions(['East']);

    // WHEN deriving slot validation
    const validation = deriveSlotValidation(['East'], regions, {});

    // THEN all 16 seeds are present (each with slot 0)
    expect(Object.keys(validation['East'])).toHaveLength(16);
  });

  it('treats a flash key set to false as none (not error)', () => {
    // GIVEN an empty slot with its flash key explicitly set to false
    const regions = buildRegions('East', {
      1: [{ schoolId: '', searchText: '' }],
    });
    const flashingSlots = { 'East-1-0': false };

    // WHEN deriving slot validation
    const validation = deriveSlotValidation(['East'], regions, flashingSlots);

    // THEN the slot is "none" (falsy flash value does not trigger error)
    expect(validation['East']['1-0']).toBe('none');
  });

  it('skips regions not in the region list', () => {
    // GIVEN regions with "East" and "South" but only "East" in the list
    const regions = {
      ...buildRegions('East', { 1: [{ schoolId: 's1', searchText: 'Duke' }] }),
      ...buildRegions('South', { 1: [{ schoolId: 's2', searchText: 'UNC' }] }),
    };

    // WHEN deriving slot validation with only "East"
    const validation = deriveSlotValidation(['East'], regions, {});

    // THEN there is no entry for "South"
    expect(validation['South']).toBeUndefined();
  });
});

describe('applyUpdateSlot', () => {
  it('updates schoolId for the specified slot', () => {
    // GIVEN a region with an empty slot at seed 1
    const prev = createInitialRegions(['East']);

    // WHEN updating seed 1 slot 0 with a schoolId
    const result = applyUpdateSlot(prev, 'East', 1, 0, { schoolId: 's1' });

    // THEN the slot has the new schoolId
    expect(result['East'][1][0].schoolId).toBe('s1');
  });

  it('updates searchText for the specified slot', () => {
    // GIVEN a region with an empty slot at seed 1
    const prev = createInitialRegions(['East']);

    // WHEN updating seed 1 slot 0 with searchText
    const result = applyUpdateSlot(prev, 'East', 1, 0, { searchText: 'Duk' });

    // THEN the slot has the new searchText
    expect(result['East'][1][0].searchText).toBe('Duk');
  });

  it('preserves existing fields when partially updating', () => {
    // GIVEN a slot with both schoolId and searchText filled
    const prev = buildRegions('East', {
      1: [{ schoolId: 's1', searchText: 'Duke' }],
    });

    // WHEN updating only the searchText
    const result = applyUpdateSlot(prev, 'East', 1, 0, { searchText: 'Duke University' });

    // THEN schoolId is preserved
    expect(result['East'][1][0].schoolId).toBe('s1');
  });

  it('does not modify other seeds', () => {
    // GIVEN a region with filled slot at seed 2
    const prev = buildRegions('East', {
      2: [{ schoolId: 's2', searchText: 'UNC' }],
    });

    // WHEN updating seed 1
    const result = applyUpdateSlot(prev, 'East', 1, 0, { schoolId: 's1' });

    // THEN seed 2 is unchanged
    expect(result['East'][2][0].schoolId).toBe('s2');
  });

  it('returns previous state for non-existent region', () => {
    // GIVEN a region set with only "East"
    const prev = createInitialRegions(['East']);

    // WHEN updating a non-existent region
    const result = applyUpdateSlot(prev, 'South', 1, 0, { schoolId: 's1' });

    // THEN the original state is returned
    expect(result).toBe(prev);
  });

  it('returns a new top-level object (immutability)', () => {
    // GIVEN a region with an empty slot
    const prev = createInitialRegions(['East']);

    // WHEN updating a slot
    const result = applyUpdateSlot(prev, 'East', 1, 0, { schoolId: 's1' });

    // THEN the result is a different reference from prev
    expect(result).not.toBe(prev);
  });

  it('does not mutate the original region state', () => {
    // GIVEN a region with an empty slot
    const prev = createInitialRegions(['East']);
    const originalSchoolId = prev['East'][1][0].schoolId;

    // WHEN updating a slot
    applyUpdateSlot(prev, 'East', 1, 0, { schoolId: 's1' });

    // THEN the original state is not mutated
    expect(prev['East'][1][0].schoolId).toBe(originalSchoolId);
  });
});

describe('applyAddPlayIn', () => {
  it('adds a second slot to a seed with one slot', () => {
    // GIVEN seed 11 with one slot
    const prev = createInitialRegions(['East']);

    // WHEN adding a play-in at seed 11
    const result = applyAddPlayIn(prev, 'East', 11);

    // THEN seed 11 now has two slots
    expect(result['East'][11]).toHaveLength(2);
  });

  it('does not add a third slot when two already exist', () => {
    // GIVEN seed 11 already has two slots
    const prev = buildRegions('East', {
      11: [
        { schoolId: 's1', searchText: 'Arizona St' },
        { schoolId: 's2', searchText: 'Texas' },
      ],
    });

    // WHEN attempting to add another play-in
    const result = applyAddPlayIn(prev, 'East', 11);

    // THEN it still has exactly two slots
    expect(result['East'][11]).toHaveLength(2);
  });

  it('initializes the new play-in slot as empty', () => {
    // GIVEN seed 11 with one filled slot
    const prev = buildRegions('East', {
      11: [{ schoolId: 's1', searchText: 'Arizona St' }],
    });

    // WHEN adding a play-in
    const result = applyAddPlayIn(prev, 'East', 11);

    // THEN the new slot has empty schoolId
    expect(result['East'][11][1].schoolId).toBe('');
  });

  it('returns previous state for non-existent region', () => {
    // GIVEN regions with only "East"
    const prev = createInitialRegions(['East']);

    // WHEN adding a play-in to a non-existent region
    const result = applyAddPlayIn(prev, 'South', 11);

    // THEN the original state is returned
    expect(result).toBe(prev);
  });

  it('does not modify other seeds in the region', () => {
    // GIVEN a region with filled slot at seed 1
    const prev = buildRegions('East', {
      1: [{ schoolId: 's1', searchText: 'Duke' }],
    });

    // WHEN adding a play-in at seed 11
    const result = applyAddPlayIn(prev, 'East', 11);

    // THEN seed 1 is unchanged
    expect(result['East'][1]).toHaveLength(1);
  });

  it('returns a new top-level object (immutability)', () => {
    // GIVEN seed 11 with one slot
    const prev = createInitialRegions(['East']);

    // WHEN adding a play-in
    const result = applyAddPlayIn(prev, 'East', 11);

    // THEN the result is a different reference from prev
    expect(result).not.toBe(prev);
  });

  it('does not mutate the original slots array', () => {
    // GIVEN seed 11 with one slot
    const prev = createInitialRegions(['East']);
    const originalLength = prev['East'][11].length;

    // WHEN adding a play-in
    applyAddPlayIn(prev, 'East', 11);

    // THEN the original slots array is not mutated
    expect(prev['East'][11]).toHaveLength(originalLength);
  });
});

describe('applyRemovePlayIn', () => {
  it('removes the first slot and keeps the second when removing index 0', () => {
    // GIVEN seed 11 with two filled slots
    const prev = buildRegions('East', {
      11: [
        { schoolId: 's1', searchText: 'Arizona St' },
        { schoolId: 's2', searchText: 'Texas' },
      ],
    });

    // WHEN removing slotIndex 0
    const result = applyRemovePlayIn(prev, 'East', 11, 0);

    // THEN the remaining slot has the second team
    expect(result['East'][11][0].schoolId).toBe('s2');
  });

  it('removes the second slot and keeps the first when removing index 1', () => {
    // GIVEN seed 11 with two filled slots
    const prev = buildRegions('East', {
      11: [
        { schoolId: 's1', searchText: 'Arizona St' },
        { schoolId: 's2', searchText: 'Texas' },
      ],
    });

    // WHEN removing slotIndex 1
    const result = applyRemovePlayIn(prev, 'East', 11, 1);

    // THEN the remaining slot has the first team
    expect(result['East'][11][0].schoolId).toBe('s1');
  });

  it('reduces to exactly one slot after removal', () => {
    // GIVEN seed 11 with two slots
    const prev = buildRegions('East', {
      11: [
        { schoolId: 's1', searchText: 'Arizona St' },
        { schoolId: 's2', searchText: 'Texas' },
      ],
    });

    // WHEN removing one play-in slot
    const result = applyRemovePlayIn(prev, 'East', 11, 1);

    // THEN seed 11 has exactly one slot
    expect(result['East'][11]).toHaveLength(1);
  });

  it('does not remove when only one slot exists', () => {
    // GIVEN seed 11 with one slot
    const prev = buildRegions('East', {
      11: [{ schoolId: 's1', searchText: 'Arizona St' }],
    });

    // WHEN attempting to remove a play-in
    const result = applyRemovePlayIn(prev, 'East', 11, 0);

    // THEN the slot count remains 1
    expect(result['East'][11]).toHaveLength(1);
  });

  it('returns previous state for non-existent region', () => {
    // GIVEN regions with only "East"
    const prev = createInitialRegions(['East']);

    // WHEN removing a play-in from a non-existent region
    const result = applyRemovePlayIn(prev, 'South', 11, 0);

    // THEN the original state is returned
    expect(result).toBe(prev);
  });

  it('returns a new top-level object (immutability)', () => {
    // GIVEN seed 11 with two slots
    const prev = buildRegions('East', {
      11: [
        { schoolId: 's1', searchText: 'Arizona St' },
        { schoolId: 's2', searchText: 'Texas' },
      ],
    });

    // WHEN removing one play-in slot
    const result = applyRemovePlayIn(prev, 'East', 11, 0);

    // THEN the result is a different reference from prev
    expect(result).not.toBe(prev);
  });

  it('does not mutate the original slots array', () => {
    // GIVEN seed 11 with two slots
    const prev = buildRegions('East', {
      11: [
        { schoolId: 's1', searchText: 'Arizona St' },
        { schoolId: 's2', searchText: 'Texas' },
      ],
    });

    // WHEN removing one play-in slot
    applyRemovePlayIn(prev, 'East', 11, 0);

    // THEN the original slots array still has two entries
    expect(prev['East'][11]).toHaveLength(2);
  });
});

describe('applySlotBlur', () => {
  it('clears searchText when slot has orphaned search text and no schoolId', () => {
    // GIVEN a slot with searchText but no schoolId (user typed but did not select)
    const prev = buildRegions('East', {
      1: [{ schoolId: '', searchText: 'Duk' }],
    });

    // WHEN blurring that slot
    const result = applySlotBlur(prev, 'East', 1, 0);

    // THEN searchText is cleared
    expect(result.regions['East'][1][0].searchText).toBe('');
  });

  it('signals that a flash should occur for orphaned text', () => {
    // GIVEN a slot with orphaned search text
    const prev = buildRegions('East', {
      1: [{ schoolId: '', searchText: 'Duk' }],
    });

    // WHEN blurring that slot
    const result = applySlotBlur(prev, 'East', 1, 0);

    // THEN shouldFlash is true
    expect(result.shouldFlash).toBe(true);
  });

  it('returns the correct flash key for the slot', () => {
    // GIVEN a slot with orphaned search text at seed 3, slot 0
    const prev = buildRegions('East', {
      3: [{ schoolId: '', searchText: 'Kan' }],
    });

    // WHEN blurring that slot
    const result = applySlotBlur(prev, 'East', 3, 0);

    // THEN flashKey matches the expected format
    expect(result.flashKey).toBe('East-3-0');
  });

  it('does not modify state when slot has a valid schoolId', () => {
    // GIVEN a slot with both schoolId and searchText (valid selection)
    const prev = buildRegions('East', {
      1: [{ schoolId: 's1', searchText: 'Duke' }],
    });

    // WHEN blurring that slot
    const result = applySlotBlur(prev, 'East', 1, 0);

    // THEN regions are unchanged (same reference)
    expect(result.regions).toBe(prev);
  });

  it('does not flash when slot has a valid schoolId', () => {
    // GIVEN a valid slot
    const prev = buildRegions('East', {
      1: [{ schoolId: 's1', searchText: 'Duke' }],
    });

    // WHEN blurring that slot
    const result = applySlotBlur(prev, 'East', 1, 0);

    // THEN shouldFlash is false
    expect(result.shouldFlash).toBe(false);
  });

  it('does not modify state when slot has empty searchText', () => {
    // GIVEN a completely empty slot
    const prev = buildRegions('East', {
      1: [{ schoolId: '', searchText: '' }],
    });

    // WHEN blurring that slot
    const result = applySlotBlur(prev, 'East', 1, 0);

    // THEN regions are unchanged (same reference)
    expect(result.regions).toBe(prev);
  });

  it('does not flash for completely empty slot', () => {
    // GIVEN a completely empty slot
    const prev = buildRegions('East', {
      1: [{ schoolId: '', searchText: '' }],
    });

    // WHEN blurring that slot
    const result = applySlotBlur(prev, 'East', 1, 0);

    // THEN shouldFlash is false
    expect(result.shouldFlash).toBe(false);
  });

  it('returns previous state for non-existent region', () => {
    // GIVEN regions with only "East"
    const prev = createInitialRegions(['East']);

    // WHEN blurring a slot in a non-existent region
    const result = applySlotBlur(prev, 'South', 1, 0);

    // THEN regions are unchanged
    expect(result.regions).toBe(prev);
  });

  it('returns a new regions object when clearing orphaned text (immutability)', () => {
    // GIVEN a slot with orphaned search text
    const prev = buildRegions('East', {
      1: [{ schoolId: '', searchText: 'Duk' }],
    });

    // WHEN blurring that slot
    const result = applySlotBlur(prev, 'East', 1, 0);

    // THEN the returned regions are a different reference from prev
    expect(result.regions).not.toBe(prev);
  });

  it('does not mutate the original slot when clearing orphaned text', () => {
    // GIVEN a slot with orphaned search text
    const prev = buildRegions('East', {
      1: [{ schoolId: '', searchText: 'Duk' }],
    });

    // WHEN blurring that slot
    applySlotBlur(prev, 'East', 1, 0);

    // THEN the original slot still has the orphaned text
    expect(prev['East'][1][0].searchText).toBe('Duk');
  });
});

describe('collectTeamsForSubmission', () => {
  it('returns empty array for empty regions', () => {
    // GIVEN regions with no filled slots
    const regions = createInitialRegions(['East']);

    // WHEN collecting teams for submission
    const teams = collectTeamsForSubmission(['East'], regions);

    // THEN the result is empty
    expect(teams).toHaveLength(0);
  });

  it('collects filled slots as team entries', () => {
    // GIVEN one filled slot
    const regions = buildRegions('East', {
      1: [{ schoolId: 's1', searchText: 'Duke' }],
    });

    // WHEN collecting teams for submission
    const teams = collectTeamsForSubmission(['East'], regions);

    // THEN one team is collected
    expect(teams).toHaveLength(1);
  });

  it('includes schoolId in the collected team', () => {
    // GIVEN a filled slot with schoolId "s1"
    const regions = buildRegions('East', {
      1: [{ schoolId: 's1', searchText: 'Duke' }],
    });

    // WHEN collecting teams for submission
    const teams = collectTeamsForSubmission(['East'], regions);

    // THEN the team has the correct schoolId
    expect(teams[0].schoolId).toBe('s1');
  });

  it('includes seed in the collected team', () => {
    // GIVEN a filled slot at seed 5
    const regions = buildRegions('East', {
      5: [{ schoolId: 's1', searchText: 'Duke' }],
    });

    // WHEN collecting teams for submission
    const teams = collectTeamsForSubmission(['East'], regions);

    // THEN the team has seed 5
    expect(teams[0].seed).toBe(5);
  });

  it('includes region in the collected team', () => {
    // GIVEN a filled slot in "West"
    const regions = buildRegions('West', {
      1: [{ schoolId: 's1', searchText: 'Duke' }],
    });

    // WHEN collecting teams for submission
    const teams = collectTeamsForSubmission(['West'], regions);

    // THEN the team has region "West"
    expect(teams[0].region).toBe('West');
  });

  it('skips slots with empty schoolId', () => {
    // GIVEN one filled and one empty slot
    const regions = buildRegions('East', {
      1: [{ schoolId: 's1', searchText: 'Duke' }],
      2: [{ schoolId: '', searchText: '' }],
    });

    // WHEN collecting teams for submission
    const teams = collectTeamsForSubmission(['East'], regions);

    // THEN only the filled slot is collected
    expect(teams).toHaveLength(1);
  });

  it('collects both play-in teams from a seed with two filled slots', () => {
    // GIVEN a play-in seed with two filled slots
    const regions = buildRegions('East', {
      11: [
        { schoolId: 's1', searchText: 'Arizona St' },
        { schoolId: 's2', searchText: 'Texas' },
      ],
    });

    // WHEN collecting teams for submission
    const teams = collectTeamsForSubmission(['East'], regions);

    // THEN both play-in teams are collected
    expect(teams).toHaveLength(2);
  });

  it('assigns the same seed to both play-in teams', () => {
    // GIVEN a play-in at seed 11
    const regions = buildRegions('East', {
      11: [
        { schoolId: 's1', searchText: 'Arizona St' },
        { schoolId: 's2', searchText: 'Texas' },
      ],
    });

    // WHEN collecting teams for submission
    const teams = collectTeamsForSubmission(['East'], regions);

    // THEN both have seed 11
    expect(teams.every((t) => t.seed === 11)).toBe(true);
  });

  it('collects teams across multiple regions', () => {
    // GIVEN filled slots across East and West
    const regions = {
      ...buildRegions('East', { 1: [{ schoolId: 's1', searchText: 'Duke' }] }),
      ...buildRegions('West', { 1: [{ schoolId: 's2', searchText: 'UNC' }] }),
    };

    // WHEN collecting teams for submission
    const teams = collectTeamsForSubmission(['East', 'West'], regions);

    // THEN two teams are collected
    expect(teams).toHaveLength(2);
  });

  it('skips regions not in the region list', () => {
    // GIVEN a filled slot in "South" but only "East" in the region list
    const regions = {
      ...buildRegions('East', {}),
      ...buildRegions('South', { 1: [{ schoolId: 's1', searchText: 'Duke' }] }),
    };

    // WHEN collecting teams with only "East" in the list
    const teams = collectTeamsForSubmission(['East'], regions);

    // THEN no teams are collected from South
    expect(teams).toHaveLength(0);
  });

  it('collects teams in seed order within a region', () => {
    // GIVEN filled slots at seed 16 and seed 1 (reverse order in the data)
    const regions = buildRegions('East', {
      16: [{ schoolId: 's16', searchText: 'Cinderella' }],
      1: [{ schoolId: 's1', searchText: 'Duke' }],
    });

    // WHEN collecting teams for submission
    const teams = collectTeamsForSubmission(['East'], regions);

    // THEN teams are ordered by seed ascending (1 before 16)
    expect(teams[0].seed).toBe(1);
  });

  it('collects teams in region-list order', () => {
    // GIVEN filled slots in West and East, with region list ordering East first
    const regions = {
      ...buildRegions('East', { 1: [{ schoolId: 's1', searchText: 'Duke' }] }),
      ...buildRegions('West', { 1: [{ schoolId: 's2', searchText: 'Kansas' }] }),
    };

    // WHEN collecting with East listed before West
    const teams = collectTeamsForSubmission(['East', 'West'], regions);

    // THEN the East team comes first
    expect(teams[0].region).toBe('East');
  });
});

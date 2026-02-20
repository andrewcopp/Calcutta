import { describe, it, expect } from 'vitest';
import { createTeamSortComparator, TeamSortRow } from './teamSort';

function makeRow(overrides: Partial<TeamSortRow> & { id: string }): TeamSortRow & { id: string } {
  return {
    seed: 1,
    region: 'East',
    teamName: 'Alpha',
    ...overrides,
  };
}

describe('createTeamSortComparator', () => {
  describe('seed sort', () => {
    it('sorts lower seeds before higher seeds', () => {
      // GIVEN two rows with different seeds
      const rows = [makeRow({ id: 'a', seed: 4 }), makeRow({ id: 'b', seed: 1 })];

      // WHEN sorting by seed
      rows.sort(createTeamSortComparator('seed'));

      // THEN lower seed comes first
      expect(rows[0].id).toBe('b');
    });

    it('breaks seed ties by region alphabetically', () => {
      // GIVEN two rows with the same seed but different regions
      const rows = [
        makeRow({ id: 'a', seed: 1, region: 'West' }),
        makeRow({ id: 'b', seed: 1, region: 'East' }),
      ];

      // WHEN sorting by seed
      rows.sort(createTeamSortComparator('seed'));

      // THEN East comes before West
      expect(rows[0].id).toBe('b');
    });

    it('breaks seed and region ties by team name alphabetically', () => {
      // GIVEN two rows with same seed and region but different names
      const rows = [
        makeRow({ id: 'a', seed: 1, region: 'East', teamName: 'Zephyr' }),
        makeRow({ id: 'b', seed: 1, region: 'East', teamName: 'Alpha' }),
      ];

      // WHEN sorting by seed
      rows.sort(createTeamSortComparator('seed'));

      // THEN Alpha comes before Zephyr
      expect(rows[0].id).toBe('b');
    });

    it('sorts undefined seeds last', () => {
      // GIVEN a row with undefined seed and a row with seed 16
      const rows = [
        makeRow({ id: 'a', seed: undefined }),
        makeRow({ id: 'b', seed: 16 }),
      ];

      // WHEN sorting by seed
      rows.sort(createTeamSortComparator('seed'));

      // THEN defined seed comes first
      expect(rows[0].id).toBe('b');
    });
  });

  describe('region sort', () => {
    it('sorts regions alphabetically', () => {
      // GIVEN two rows with different regions
      const rows = [
        makeRow({ id: 'a', region: 'West' }),
        makeRow({ id: 'b', region: 'East' }),
      ];

      // WHEN sorting by region
      rows.sort(createTeamSortComparator('region'));

      // THEN East comes before West
      expect(rows[0].id).toBe('b');
    });

    it('breaks region ties by seed', () => {
      // GIVEN two rows with the same region but different seeds
      const rows = [
        makeRow({ id: 'a', region: 'East', seed: 8 }),
        makeRow({ id: 'b', region: 'East', seed: 2 }),
      ];

      // WHEN sorting by region
      rows.sort(createTeamSortComparator('region'));

      // THEN lower seed comes first
      expect(rows[0].id).toBe('b');
    });

    it('breaks region and seed ties by team name', () => {
      // GIVEN two rows with same region and seed but different names
      const rows = [
        makeRow({ id: 'a', region: 'East', seed: 1, teamName: 'Zephyr' }),
        makeRow({ id: 'b', region: 'East', seed: 1, teamName: 'Alpha' }),
      ];

      // WHEN sorting by region
      rows.sort(createTeamSortComparator('region'));

      // THEN Alpha comes before Zephyr
      expect(rows[0].id).toBe('b');
    });
  });

  describe('team sort', () => {
    it('sorts team names alphabetically', () => {
      // GIVEN two rows with different names
      const rows = [
        makeRow({ id: 'a', teamName: 'Zephyr' }),
        makeRow({ id: 'b', teamName: 'Alpha' }),
      ];

      // WHEN sorting by team
      rows.sort(createTeamSortComparator('team'));

      // THEN Alpha comes before Zephyr
      expect(rows[0].id).toBe('b');
    });

    it('breaks team name ties by seed', () => {
      // GIVEN two rows with the same name but different seeds
      const rows = [
        makeRow({ id: 'a', teamName: 'Alpha', seed: 8 }),
        makeRow({ id: 'b', teamName: 'Alpha', seed: 2 }),
      ];

      // WHEN sorting by team
      rows.sort(createTeamSortComparator('team'));

      // THEN lower seed comes first
      expect(rows[0].id).toBe('b');
    });
  });

  describe('edge cases', () => {
    it('handles empty region strings', () => {
      // GIVEN rows with empty region strings
      const rows = [
        makeRow({ id: 'a', region: '' }),
        makeRow({ id: 'b', region: '' }),
      ];

      // WHEN sorting by region
      rows.sort(createTeamSortComparator('region'));

      // THEN does not throw and falls through to seed tiebreaker
      expect(rows).toHaveLength(2);
    });

    it('handles empty team name strings', () => {
      // GIVEN rows with empty team names
      const rows = [
        makeRow({ id: 'a', teamName: '', seed: 5 }),
        makeRow({ id: 'b', teamName: '', seed: 2 }),
      ];

      // WHEN sorting by team
      rows.sort(createTeamSortComparator('team'));

      // THEN falls through to seed tiebreaker
      expect(rows[0].id).toBe('b');
    });

    it('produces stable order for identical rows', () => {
      // GIVEN two identical rows
      const rows = [
        makeRow({ id: 'a', seed: 1, region: 'East', teamName: 'Alpha' }),
        makeRow({ id: 'b', seed: 1, region: 'East', teamName: 'Alpha' }),
      ];

      // WHEN sorting by seed
      rows.sort(createTeamSortComparator('seed'));

      // THEN comparator returns 0 (stable)
      const cmp = createTeamSortComparator<TeamSortRow & { id: string }>('seed');
      expect(cmp(rows[0], rows[1])).toBe(0);
    });
  });
});

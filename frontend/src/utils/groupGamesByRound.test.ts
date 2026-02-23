import { describe, it, expect } from 'vitest';
import { groupGamesByRound } from './groupGamesByRound';
import type { BracketGame, BracketRound } from '../schemas/bracket';
import { ROUND_ORDER } from '../schemas/bracket';

// ---------------------------------------------------------------------------
// Helper to create a minimal BracketGame for testing purposes.
// Only the fields relevant to grouping and sorting are required.
// ---------------------------------------------------------------------------
function makeGame(overrides: Partial<BracketGame> & { round: string; sortOrder: number }): BracketGame {
  return {
    gameId: `game-${overrides.round}-${overrides.sortOrder}`,
    round: overrides.round,
    region: overrides.region ?? 'East',
    sortOrder: overrides.sortOrder,
    canSelect: false,
    ...overrides,
  };
}

describe('groupGamesByRound', () => {
  it('groups games correctly by round', () => {
    // GIVEN games across multiple rounds
    const games: BracketGame[] = [
      makeGame({ round: 'round_of_64', sortOrder: 1 }),
      makeGame({ round: 'round_of_64', sortOrder: 2 }),
      makeGame({ round: 'round_of_32', sortOrder: 1 }),
      makeGame({ round: 'sweet_16', sortOrder: 1 }),
    ];

    // WHEN grouping by round
    const result = groupGamesByRound(games);

    // THEN each round contains only its games
    expect(result['round_of_64']).toHaveLength(2);
    expect(result['round_of_32']).toHaveLength(1);
    expect(result['sweet_16']).toHaveLength(1);
  });

  it('returns empty arrays for all rounds when given empty input', () => {
    // GIVEN no games
    const games: BracketGame[] = [];

    // WHEN grouping by round
    const result = groupGamesByRound(games);

    // THEN every round key exists with an empty array
    for (const round of ROUND_ORDER) {
      expect(result[round]).toEqual([]);
    }
  });

  it('includes all round keys even when no games match', () => {
    // GIVEN games only in round_of_64
    const games: BracketGame[] = [makeGame({ round: 'round_of_64', sortOrder: 1 })];

    // WHEN grouping by round
    const result = groupGamesByRound(games);

    // THEN all ROUND_ORDER keys are present
    for (const round of ROUND_ORDER) {
      expect(result).toHaveProperty(round);
    }
  });

  it('handles a single game', () => {
    // GIVEN a single championship game
    const games: BracketGame[] = [makeGame({ round: 'championship', sortOrder: 1 })];

    // WHEN grouping by round
    const result = groupGamesByRound(games);

    // THEN only the championship round has the game
    expect(result['championship']).toHaveLength(1);
    expect(result['championship'][0].round).toBe('championship');
  });

  it('sorts games within each round by sortOrder ascending', () => {
    // GIVEN games in round_of_64 added in reverse sortOrder
    const games: BracketGame[] = [
      makeGame({ gameId: 'g3', round: 'round_of_64', sortOrder: 3 }),
      makeGame({ gameId: 'g1', round: 'round_of_64', sortOrder: 1 }),
      makeGame({ gameId: 'g2', round: 'round_of_64', sortOrder: 2 }),
    ];

    // WHEN grouping by round
    const result = groupGamesByRound(games);

    // THEN games in round_of_64 are sorted by sortOrder
    const sortOrders = result['round_of_64'].map((g) => g.sortOrder);
    expect(sortOrders).toEqual([1, 2, 3]);
  });

  it('preserves game data through grouping', () => {
    // GIVEN a game with specific field values
    const game = makeGame({
      gameId: 'specific-game',
      round: 'elite_8',
      region: 'West',
      sortOrder: 5,
      canSelect: true,
    });

    // WHEN grouping by round
    const result = groupGamesByRound([game]);

    // THEN the grouped game retains all its original fields
    expect(result['elite_8'][0].gameId).toBe('specific-game');
    expect(result['elite_8'][0].region).toBe('West');
    expect(result['elite_8'][0].canSelect).toBe(true);
  });

  it('handles games in every round simultaneously', () => {
    // GIVEN one game in each round
    const games: BracketGame[] = ROUND_ORDER.map((round, index) => makeGame({ round, sortOrder: index + 1 }));

    // WHEN grouping by round
    const result = groupGamesByRound(games);

    // THEN each round has exactly one game
    for (const round of ROUND_ORDER) {
      expect(result[round]).toHaveLength(1);
    }
  });

  it('does not include games with unrecognized round values in any group', () => {
    // GIVEN a game with a round not in ROUND_ORDER alongside a valid game
    const games: BracketGame[] = [
      makeGame({ round: 'round_of_64', sortOrder: 1 }),
      makeGame({ round: 'unknown_round' as BracketRound, sortOrder: 1 }),
    ];

    // WHEN grouping by round
    const result = groupGamesByRound(games);

    // THEN the total count across all known rounds is 1
    const totalGames = ROUND_ORDER.reduce((sum, round) => sum + result[round].length, 0);
    expect(totalGames).toBe(1);
  });
});

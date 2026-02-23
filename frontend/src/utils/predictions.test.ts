import { describe, it, expect } from 'vitest';
import { TeamPrediction } from '../schemas/tournament';
import { getConditionalProbability, formatPercent } from './predictions';

function makeTeam(overrides: Partial<TeamPrediction> = {}): TeamPrediction {
  return {
    teamId: 'team-1',
    schoolName: 'Test U',
    seed: 1,
    region: 'East',
    wins: 0,
    byes: 1,
    isEliminated: false,
    pRound1: 1.0,
    pRound2: 0.95,
    pRound3: 0.8,
    pRound4: 0.6,
    pRound5: 0.4,
    pRound6: 0.2,
    pRound7: 0.1,
    expectedPoints: 10,
    ...overrides,
  };
}

describe('getConditionalProbability', () => {
  describe('pre-tournament (throughRound=0)', () => {
    it('returns raw probability for non-bye team', () => {
      // GIVEN a team with no byes
      const team = makeTeam({ byes: 0 });

      // WHEN getting probability for round 2
      const result = getConditionalProbability(team, 2, 0);

      // THEN returns raw pRound2
      expect(result.value).toBe(0.95);
    });

    it('returns raw probability with empty style', () => {
      // GIVEN a team with no byes
      const team = makeTeam({ byes: 0 });

      // WHEN getting probability for round 3
      const result = getConditionalProbability(team, 3, 0);

      // THEN style is empty
      expect(result.style).toBe('');
    });

    it('returns 100% for bye team in round covered by byes', () => {
      // GIVEN a team with 1 bye
      const team = makeTeam({ byes: 1 });

      // WHEN getting probability for round 1 (First Four)
      const result = getConditionalProbability(team, 1, 0);

      // THEN returns 1
      expect(result.value).toBe(1);
    });

    it('returns green style for bye team in round covered by byes', () => {
      // GIVEN a team with 1 bye
      const team = makeTeam({ byes: 1 });

      // WHEN getting probability for round 1
      const result = getConditionalProbability(team, 1, 0);

      // THEN returns green style
      expect(result.style).toBe('text-green-600 font-semibold');
    });

    it('returns raw probability for bye team in round beyond byes', () => {
      // GIVEN a team with 1 bye
      const team = makeTeam({ byes: 1 });

      // WHEN getting probability for round 2 (beyond bye)
      const result = getConditionalProbability(team, 2, 0);

      // THEN returns raw pRound2
      expect(result.value).toBe(0.95);
    });
  });

  describe('resolved rounds (round <= throughRound)', () => {
    it('returns 1 for team that advanced through round', () => {
      // GIVEN a team with 2 wins + 1 bye (progress=3), throughRound=2
      const team = makeTeam({ wins: 2, byes: 1 });

      // WHEN getting probability for round 2
      const result = getConditionalProbability(team, 2, 2);

      // THEN returns 1
      expect(result.value).toBe(1);
    });

    it('returns green style for team that advanced', () => {
      // GIVEN a team with 2 wins + 1 bye (progress=3), throughRound=2
      const team = makeTeam({ wins: 2, byes: 1 });

      // WHEN getting probability for round 2
      const result = getConditionalProbability(team, 2, 2);

      // THEN returns green style
      expect(result.style).toBe('text-green-600 font-semibold');
    });

    it('returns 0 for team eliminated before round', () => {
      // GIVEN a team with 1 win + 1 bye (progress=2), throughRound=3
      const team = makeTeam({ wins: 1, byes: 1 });

      // WHEN getting probability for round 3
      const result = getConditionalProbability(team, 3, 3);

      // THEN returns 0
      expect(result.value).toBe(0);
    });

    it('returns muted style for eliminated team', () => {
      // GIVEN a team with 1 win + 1 bye (progress=2), throughRound=3
      const team = makeTeam({ wins: 1, byes: 1 });

      // WHEN getting probability for round 3
      const result = getConditionalProbability(team, 3, 3);

      // THEN returns muted style
      expect(result.style).toBe('text-muted-foreground');
    });
  });

  describe('future rounds, team eliminated before checkpoint', () => {
    it('returns 0 for team that did not survive to checkpoint', () => {
      // GIVEN a team with progress=1, throughRound=2
      const team = makeTeam({ wins: 0, byes: 1 });

      // WHEN getting probability for round 3 (future)
      const result = getConditionalProbability(team, 3, 2);

      // THEN returns 0
      expect(result.value).toBe(0);
    });

    it('returns muted style for eliminated team in future round', () => {
      // GIVEN a team with progress=1, throughRound=2
      const team = makeTeam({ wins: 0, byes: 1 });

      // WHEN getting probability for round 3
      const result = getConditionalProbability(team, 3, 2);

      // THEN returns muted style
      expect(result.style).toBe('text-muted-foreground');
    });
  });

  describe('future rounds, team alive at checkpoint', () => {
    it('returns conditional probability for alive team', () => {
      // GIVEN a team with progress=2, throughRound=2, pRound3=0.8, pRound2=0.95
      const team = makeTeam({ wins: 1, byes: 1 });

      // WHEN getting probability for round 3
      const result = getConditionalProbability(team, 3, 2);

      // THEN returns pRound3 / pRound2
      expect(result.value).toBeCloseTo(0.8 / 0.95);
    });

    it('returns empty style for alive team in future round', () => {
      // GIVEN a team alive at checkpoint
      const team = makeTeam({ wins: 1, byes: 1 });

      // WHEN getting probability for round 3
      const result = getConditionalProbability(team, 3, 2);

      // THEN style is empty
      expect(result.style).toBe('');
    });

    it('clamps conditional probability to 1', () => {
      // GIVEN a team where pRound3 > pRound2 due to floating point
      const team = makeTeam({ wins: 1, byes: 1, pRound2: 0.5, pRound3: 0.6 });

      // WHEN getting probability for round 3
      const result = getConditionalProbability(team, 3, 2);

      // THEN clamps to 1
      expect(result.value).toBe(1);
    });

    it('returns 0 when checkpoint probability is 0', () => {
      // GIVEN a team where pRound2 = 0
      const team = makeTeam({ wins: 1, byes: 1, pRound2: 0, pRound3: 0 });

      // WHEN getting probability for round 3
      const result = getConditionalProbability(team, 3, 2);

      // THEN returns 0 (avoids division by zero)
      expect(result.value).toBe(0);
    });
  });

  describe('future rounds, team alive (normalized with roundSums)', () => {
    it('normalizes championship probability using roundSums', () => {
      // GIVEN a team alive at throughRound=5 with pRound7=0.1, and roundSums showing total=0.5 at round 7
      const team = makeTeam({ wins: 4, byes: 1, pRound7: 0.1 });
      const roundSums = { 1: 64, 2: 32, 3: 16, 4: 8, 5: 4, 6: 2, 7: 0.5 };

      // WHEN getting probability for round 7 (championship)
      const result = getConditionalProbability(team, 7, 5, roundSums);

      // THEN normalizes: (0.1 / 0.5) * 1 = 0.2
      expect(result.value).toBeCloseTo(0.2);
    });

    it('clamps to 1 when normalization would exceed 1', () => {
      // GIVEN a team whose normalized probability would exceed 1
      const team = makeTeam({ wins: 4, byes: 1, pRound7: 0.8 });
      const roundSums = { 1: 64, 2: 32, 3: 16, 4: 8, 5: 4, 6: 2, 7: 0.5 };

      // WHEN getting probability for round 7
      const result = getConditionalProbability(team, 7, 5, roundSums);

      // THEN clamps to 1
      expect(result.value).toBe(1);
    });

    it('returns 0 when roundSums is 0 for the round', () => {
      // GIVEN roundSums[7] = 0
      const team = makeTeam({ wins: 4, byes: 1, pRound7: 0.1 });
      const roundSums = { 1: 64, 2: 32, 3: 16, 4: 8, 5: 4, 6: 2, 7: 0 };

      // WHEN getting probability for round 7
      const result = getConditionalProbability(team, 7, 5, roundSums);

      // THEN returns 0
      expect(result.value).toBe(0);
    });

    it('does not affect pre-tournament mode', () => {
      // GIVEN pre-tournament (throughRound=0) with roundSums provided
      const team = makeTeam({ byes: 0 });
      const roundSums = { 1: 64, 2: 32, 3: 16, 4: 8, 5: 4, 6: 2, 7: 1 };

      // WHEN getting probability for round 2
      const result = getConditionalProbability(team, 2, 0, roundSums);

      // THEN returns raw pRound2 (roundSums ignored)
      expect(result.value).toBe(0.95);
    });

    it('does not affect resolved rounds', () => {
      // GIVEN a team that advanced through round 2 with roundSums
      const team = makeTeam({ wins: 2, byes: 1 });
      const roundSums = { 1: 64, 2: 32, 3: 16, 4: 8, 5: 4, 6: 2, 7: 1 };

      // WHEN getting probability for round 2
      const result = getConditionalProbability(team, 2, 2, roundSums);

      // THEN returns 1 (resolved, not affected by roundSums)
      expect(result.value).toBe(1);
    });

    it('does not affect eliminated teams', () => {
      // GIVEN a team eliminated before checkpoint with roundSums
      const team = makeTeam({ wins: 0, byes: 1 });
      const roundSums = { 1: 64, 2: 32, 3: 16, 4: 8, 5: 4, 6: 2, 7: 1 };

      // WHEN getting probability for round 3
      const result = getConditionalProbability(team, 3, 2, roundSums);

      // THEN returns 0 (eliminated, not affected by roundSums)
      expect(result.value).toBe(0);
    });

    it('falls back to per-team conditional when roundSums not provided', () => {
      // GIVEN a team alive at throughRound=2, no roundSums
      const team = makeTeam({ wins: 1, byes: 1 });

      // WHEN getting probability for round 3 without roundSums
      const result = getConditionalProbability(team, 3, 2);

      // THEN falls back to pRound3 / pRound2
      expect(result.value).toBeCloseTo(0.8 / 0.95);
    });
  });
});

describe('formatPercent', () => {
  it('returns 100% for values at or above 0.9995', () => {
    expect(formatPercent(1.0)).toBe('100%');
  });

  it('returns 100% for value just at threshold', () => {
    expect(formatPercent(0.9995)).toBe('100%');
  });

  it('returns em dash for values below 0.0005', () => {
    expect(formatPercent(0)).toBe('—');
  });

  it('returns em dash for very small positive value', () => {
    expect(formatPercent(0.0004)).toBe('—');
  });

  it('formats mid-range value with one decimal', () => {
    expect(formatPercent(0.456)).toBe('45.6%');
  });

  it('formats value near 50%', () => {
    expect(formatPercent(0.5)).toBe('50.0%');
  });
});

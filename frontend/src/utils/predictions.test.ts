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
    it('returns raw pRound for alive team', () => {
      // GIVEN a team with progress=2, throughRound=2, pRound3=0.8
      const team = makeTeam({ wins: 1, byes: 1, pRound3: 0.8 });

      // WHEN getting probability for round 3
      const result = getConditionalProbability(team, 3, 2);

      // THEN returns raw pRound3 (already conditional from backend)
      expect(result.value).toBe(0.8);
    });

    it('returns empty style for alive team in future round', () => {
      // GIVEN a team alive at checkpoint
      const team = makeTeam({ wins: 1, byes: 1 });

      // WHEN getting probability for round 3
      const result = getConditionalProbability(team, 3, 2);

      // THEN style is empty
      expect(result.style).toBe('');
    });

    it('returns raw pRound7 for championship probability', () => {
      // GIVEN a team alive at throughRound=5 with pRound7=0.25
      const team = makeTeam({ wins: 4, byes: 1, pRound7: 0.25 });

      // WHEN getting probability for round 7
      const result = getConditionalProbability(team, 7, 5);

      // THEN returns raw pRound7
      expect(result.value).toBe(0.25);
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

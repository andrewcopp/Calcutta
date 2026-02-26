import { describe, it, expect } from 'vitest';
import {
  computeBudgetRemaining,
  computeTeamCount,
  computeValidationErrors,
  type InvestmentValidationConfig,
  type InvestmentTeamLookup,
} from './investmentValidation';

// ---------------------------------------------------------------------------
// Tests for the extracted investment-validation pure functions used by useInvesting.
// ---------------------------------------------------------------------------

describe('computeBudgetRemaining', () => {
  it('returns full budget when no investments are placed', () => {
    // GIVEN no investments and a budget of 100
    // WHEN computing budget remaining
    const remaining = computeBudgetRemaining({}, 100);

    // THEN the full budget is remaining
    expect(remaining).toBe(100);
  });

  it('subtracts investment amounts from budget', () => {
    // GIVEN investments totaling 35 and a budget of 100
    const investments = { t1: 20, t2: 15 };

    // WHEN computing budget remaining
    const remaining = computeBudgetRemaining(investments, 100);

    // THEN remaining is 65
    expect(remaining).toBe(65);
  });

  it('returns negative when over budget', () => {
    // GIVEN investments totaling 120 and a budget of 100
    const investments = { t1: 60, t2: 60 };

    // WHEN computing budget remaining
    const remaining = computeBudgetRemaining(investments, 100);

    // THEN remaining is -20
    expect(remaining).toBe(-20);
  });

  it('returns zero when investments exactly equal budget', () => {
    // GIVEN investments totaling exactly 100 and a budget of 100
    const investments = { t1: 50, t2: 50 };

    // WHEN computing budget remaining
    const remaining = computeBudgetRemaining(investments, 100);

    // THEN remaining is 0
    expect(remaining).toBe(0);
  });
});

describe('computeTeamCount', () => {
  it('returns zero when no investments exist', () => {
    // GIVEN no investments
    // WHEN computing team count
    const count = computeTeamCount({});

    // THEN count is 0
    expect(count).toBe(0);
  });

  it('counts distinct team IDs with investments', () => {
    // GIVEN investments on 3 teams
    const investments = { t1: 10, t2: 20, t3: 5 };

    // WHEN computing team count
    const count = computeTeamCount(investments);

    // THEN count is 3
    expect(count).toBe(3);
  });
});

describe('computeValidationErrors', () => {
  const defaultConfig: InvestmentValidationConfig = {
    minTeams: 3,
    maxTeams: 10,
    maxInvestmentCredits: 50,
    budget: 100,
  };

  it('reports too few teams when under minimum', () => {
    // GIVEN only 1 team selected but minimum is 3
    const investments = { t1: 10 };

    // WHEN computing validation errors
    const errors = computeValidationErrors(investments, defaultConfig, []);

    // THEN an error about minimum teams is present
    expect(errors).toContain('Select at least 3 teams');
  });

  it('reports too many teams when over maximum', () => {
    // GIVEN 11 teams selected but maximum is 10
    const investments: Record<string, number> = {};
    for (let i = 1; i <= 11; i++) {
      investments[`t${i}`] = 5;
    }

    // WHEN computing validation errors
    const errors = computeValidationErrors(investments, defaultConfig, []);

    // THEN an error about maximum teams is present
    expect(errors).toContain('Select at most 10 teams');
  });

  it('reports over-budget when total investments exceed budget', () => {
    // GIVEN investments totaling 110 with a budget of 100
    const investments = { t1: 40, t2: 40, t3: 30 };

    // WHEN computing validation errors
    const errors = computeValidationErrors(investments, defaultConfig, []);

    // THEN an error about being over budget is present
    expect(errors).toContain('Over budget by 10.00 credits');
  });

  it('reports investment below minimum when an investment is less than 1', () => {
    // GIVEN an investment of 0 (below MIN_INVESTMENT of 1)
    const investments = { t1: 0, t2: 10, t3: 10 };

    // WHEN computing validation errors
    const errors = computeValidationErrors(investments, defaultConfig, []);

    // THEN an error about minimum investment is present
    expect(errors).toContain('All investments must be at least 1 credit');
  });

  it('reports investment exceeds max with team name when available', () => {
    // GIVEN an investment of 60 exceeding maxInvestmentCredits of 50, with a known team name
    const investments = { t1: 60, t2: 10, t3: 10 };
    const teams: InvestmentTeamLookup[] = [{ id: 't1', schoolName: 'Duke' }];

    // WHEN computing validation errors
    const errors = computeValidationErrors(investments, defaultConfig, teams);

    // THEN the error includes the team name
    expect(errors).toContain('Investment in Duke exceeds max 50 credits');
  });

  it('falls back to "team" when school name is not found for over-max investment', () => {
    // GIVEN an investment exceeding max for an unknown team
    const investments = { t1: 60, t2: 10, t3: 10 };

    // WHEN computing validation errors with no team lookup
    const errors = computeValidationErrors(investments, defaultConfig, []);

    // THEN the error falls back to "team"
    expect(errors).toContain('Investment in team exceeds max 50 credits');
  });

  it('returns no errors when investments are valid', () => {
    // GIVEN valid investments: 3 teams, within budget, within investment limits
    const investments = { t1: 30, t2: 30, t3: 30 };

    // WHEN computing validation errors
    const errors = computeValidationErrors(investments, defaultConfig, []);

    // THEN there are no errors
    expect(errors).toEqual([]);
  });

  it('returns no errors at exact budget boundary', () => {
    // GIVEN investments that exactly equal the budget
    const investments = { t1: 34, t2: 33, t3: 33 };

    // WHEN computing validation errors
    const errors = computeValidationErrors(investments, defaultConfig, []);

    // THEN there are no errors
    expect(errors).toEqual([]);
  });

  it('reports multiple errors simultaneously', () => {
    // GIVEN a single team with an investment of 0 (too few teams + investment below minimum)
    const investments = { t1: 0 };

    // WHEN computing validation errors
    const errors = computeValidationErrors(investments, defaultConfig, []);

    // THEN both the min-teams and min-investment errors are present
    expect(errors).toContain('Select at least 3 teams');
    expect(errors).toContain('All investments must be at least 1 credit');
  });

  it('does not report over-budget at exact zero remaining', () => {
    // GIVEN investments that sum to exactly the budget with exactly minTeams
    const investments = { t1: 34, t2: 33, t3: 33 };

    // WHEN computing validation errors
    const errors = computeValidationErrors(investments, defaultConfig, []);

    // THEN no over-budget error is present
    expect(errors.some((e) => e.includes('Over budget'))).toBe(false);
  });

  it('accepts investments at exactly the max investment limit', () => {
    // GIVEN an investment at exactly maxInvestmentCredits of 50
    const investments = { t1: 50, t2: 25, t3: 25 };

    // WHEN computing validation errors
    const errors = computeValidationErrors(investments, defaultConfig, []);

    // THEN no max-investment error is present
    expect(errors.some((e) => e.includes('exceeds max'))).toBe(false);
  });

  it('accepts exactly the minimum number of teams', () => {
    // GIVEN exactly 3 teams (the minimum)
    const investments = { t1: 10, t2: 10, t3: 10 };

    // WHEN computing validation errors
    const errors = computeValidationErrors(investments, defaultConfig, []);

    // THEN no min-teams error is present
    expect(errors.some((e) => e.includes('at least'))).toBe(false);
  });

  it('accepts exactly the maximum number of teams', () => {
    // GIVEN exactly 10 teams (the maximum)
    const investments: Record<string, number> = {};
    for (let i = 1; i <= 10; i++) {
      investments[`t${i}`] = 10;
    }

    // WHEN computing validation errors
    const errors = computeValidationErrors(investments, defaultConfig, []);

    // THEN no max-teams error is present
    expect(errors.some((e) => e.includes('at most'))).toBe(false);
  });
});

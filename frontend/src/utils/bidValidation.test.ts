import { describe, it, expect } from 'vitest';
import {
  computeBudgetRemaining,
  computeTeamCount,
  computeValidationErrors,
  type BidValidationConfig,
  type BidTeamLookup,
} from './bidValidation';

// ---------------------------------------------------------------------------
// Tests for the extracted bid-validation pure functions used by useBidding.
// ---------------------------------------------------------------------------

describe('computeBudgetRemaining', () => {
  it('returns full budget when no bids are placed', () => {
    // GIVEN no bids and a budget of 100
    // WHEN computing budget remaining
    const remaining = computeBudgetRemaining({}, 100);

    // THEN the full budget is remaining
    expect(remaining).toBe(100);
  });

  it('subtracts bid amounts from budget', () => {
    // GIVEN bids totaling 35 and a budget of 100
    const bids = { t1: 20, t2: 15 };

    // WHEN computing budget remaining
    const remaining = computeBudgetRemaining(bids, 100);

    // THEN remaining is 65
    expect(remaining).toBe(65);
  });

  it('returns negative when over budget', () => {
    // GIVEN bids totaling 120 and a budget of 100
    const bids = { t1: 60, t2: 60 };

    // WHEN computing budget remaining
    const remaining = computeBudgetRemaining(bids, 100);

    // THEN remaining is -20
    expect(remaining).toBe(-20);
  });

  it('returns zero when bids exactly equal budget', () => {
    // GIVEN bids totaling exactly 100 and a budget of 100
    const bids = { t1: 50, t2: 50 };

    // WHEN computing budget remaining
    const remaining = computeBudgetRemaining(bids, 100);

    // THEN remaining is 0
    expect(remaining).toBe(0);
  });
});

describe('computeTeamCount', () => {
  it('returns zero when no bids exist', () => {
    // GIVEN no bids
    // WHEN computing team count
    const count = computeTeamCount({});

    // THEN count is 0
    expect(count).toBe(0);
  });

  it('counts distinct team IDs with bids', () => {
    // GIVEN bids on 3 teams
    const bids = { t1: 10, t2: 20, t3: 5 };

    // WHEN computing team count
    const count = computeTeamCount(bids);

    // THEN count is 3
    expect(count).toBe(3);
  });
});

describe('computeValidationErrors', () => {
  const defaultConfig: BidValidationConfig = {
    minTeams: 3,
    maxTeams: 10,
    maxBidPoints: 50,
    budget: 100,
  };

  it('reports too few teams when under minimum', () => {
    // GIVEN only 1 team selected but minimum is 3
    const bids = { t1: 10 };

    // WHEN computing validation errors
    const errors = computeValidationErrors(bids, defaultConfig, []);

    // THEN an error about minimum teams is present
    expect(errors).toContain('Select at least 3 teams');
  });

  it('reports too many teams when over maximum', () => {
    // GIVEN 11 teams selected but maximum is 10
    const bids: Record<string, number> = {};
    for (let i = 1; i <= 11; i++) {
      bids[`t${i}`] = 5;
    }

    // WHEN computing validation errors
    const errors = computeValidationErrors(bids, defaultConfig, []);

    // THEN an error about maximum teams is present
    expect(errors).toContain('Select at most 10 teams');
  });

  it('reports over-budget when total bids exceed budget', () => {
    // GIVEN bids totaling 110 with a budget of 100
    const bids = { t1: 40, t2: 40, t3: 30 };

    // WHEN computing validation errors
    const errors = computeValidationErrors(bids, defaultConfig, []);

    // THEN an error about being over budget is present
    expect(errors).toContain('Over budget by 10.00 credits');
  });

  it('reports bid below minimum when a bid is less than 1', () => {
    // GIVEN a bid of 0 (below MIN_BID of 1)
    const bids = { t1: 0, t2: 10, t3: 10 };

    // WHEN computing validation errors
    const errors = computeValidationErrors(bids, defaultConfig, []);

    // THEN an error about minimum bid is present
    expect(errors).toContain('All bids must be at least 1 credit');
  });

  it('reports bid exceeds max with team name when available', () => {
    // GIVEN a bid of 60 exceeding maxBidPoints of 50, with a known team name
    const bids = { t1: 60, t2: 10, t3: 10 };
    const teams: BidTeamLookup[] = [{ id: 't1', schoolName: 'Duke' }];

    // WHEN computing validation errors
    const errors = computeValidationErrors(bids, defaultConfig, teams);

    // THEN the error includes the team name
    expect(errors).toContain('Bid on Duke exceeds max 50 credits');
  });

  it('falls back to "team" when school name is not found for over-max bid', () => {
    // GIVEN a bid exceeding max for an unknown team
    const bids = { t1: 60, t2: 10, t3: 10 };

    // WHEN computing validation errors with no team lookup
    const errors = computeValidationErrors(bids, defaultConfig, []);

    // THEN the error falls back to "team"
    expect(errors).toContain('Bid on team exceeds max 50 credits');
  });

  it('returns no errors when bids are valid', () => {
    // GIVEN valid bids: 3 teams, within budget, within bid limits
    const bids = { t1: 30, t2: 30, t3: 30 };

    // WHEN computing validation errors
    const errors = computeValidationErrors(bids, defaultConfig, []);

    // THEN there are no errors
    expect(errors).toEqual([]);
  });

  it('returns no errors at exact budget boundary', () => {
    // GIVEN bids that exactly equal the budget
    const bids = { t1: 34, t2: 33, t3: 33 };

    // WHEN computing validation errors
    const errors = computeValidationErrors(bids, defaultConfig, []);

    // THEN there are no errors
    expect(errors).toEqual([]);
  });

  it('reports multiple errors simultaneously', () => {
    // GIVEN a single team with a bid of 0 (too few teams + bid below minimum)
    const bids = { t1: 0 };

    // WHEN computing validation errors
    const errors = computeValidationErrors(bids, defaultConfig, []);

    // THEN both the min-teams and min-bid errors are present
    expect(errors).toContain('Select at least 3 teams');
    expect(errors).toContain('All bids must be at least 1 credit');
  });

  it('does not report over-budget at exact zero remaining', () => {
    // GIVEN bids that sum to exactly the budget with exactly minTeams
    const bids = { t1: 34, t2: 33, t3: 33 };

    // WHEN computing validation errors
    const errors = computeValidationErrors(bids, defaultConfig, []);

    // THEN no over-budget error is present
    expect(errors.some((e) => e.includes('Over budget'))).toBe(false);
  });

  it('accepts bids at exactly the max bid limit', () => {
    // GIVEN a bid at exactly maxBidPoints of 50
    const bids = { t1: 50, t2: 25, t3: 25 };

    // WHEN computing validation errors
    const errors = computeValidationErrors(bids, defaultConfig, []);

    // THEN no max-bid error is present
    expect(errors.some((e) => e.includes('exceeds max'))).toBe(false);
  });

  it('accepts exactly the minimum number of teams', () => {
    // GIVEN exactly 3 teams (the minimum)
    const bids = { t1: 10, t2: 10, t3: 10 };

    // WHEN computing validation errors
    const errors = computeValidationErrors(bids, defaultConfig, []);

    // THEN no min-teams error is present
    expect(errors.some((e) => e.includes('at least'))).toBe(false);
  });

  it('accepts exactly the maximum number of teams', () => {
    // GIVEN exactly 10 teams (the maximum)
    const bids: Record<string, number> = {};
    for (let i = 1; i <= 10; i++) {
      bids[`t${i}`] = 10;
    }

    // WHEN computing validation errors
    const errors = computeValidationErrors(bids, defaultConfig, []);

    // THEN no max-teams error is present
    expect(errors.some((e) => e.includes('at most'))).toBe(false);
  });
});

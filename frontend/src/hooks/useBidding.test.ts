import { describe, it, expect } from 'vitest';
import { getSeedVariant } from './useBidding';

// ---------------------------------------------------------------------------
// The useBidding hook is deeply entangled with React Router (useParams,
// useNavigate), React Query (useQuery, useMutation), and React state hooks.
// Testing the hook as a whole would require a full jsdom environment plus
// mocked providers for Router, QueryClient, and service modules.
//
// Instead, we test the exported pure functions and constants that contain
// meaningful logic:
//   - getSeedVariant: maps a seed number to a UI variant string
//
// We also test the validation and budget computation logic by extracting it
// into pure functions that mirror the hook's useMemo bodies.
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// Extracted pure validation logic (mirrors the useMemo in the hook)
// ---------------------------------------------------------------------------

const MIN_BID = 1;

interface ValidationConfig {
  minTeams: number;
  maxTeams: number;
  maxBid: number;
  budget: number;
}

interface TeamLookup {
  id: string;
  schoolName: string;
}

function computeBudgetRemaining(bidsByTeamId: Record<string, number>, budget: number): number {
  const spent = Object.values(bidsByTeamId).reduce((sum, bid) => sum + bid, 0);
  return budget - spent;
}

function computeTeamCount(bidsByTeamId: Record<string, number>): number {
  return Object.keys(bidsByTeamId).length;
}

function computeValidationErrors(
  bidsByTeamId: Record<string, number>,
  config: ValidationConfig,
  teams: TeamLookup[],
): string[] {
  const teamCount = computeTeamCount(bidsByTeamId);
  const budgetRemaining = computeBudgetRemaining(bidsByTeamId, config.budget);
  const errors: string[] = [];

  if (teamCount < config.minTeams) {
    errors.push(`Select at least ${config.minTeams} teams`);
  }

  if (teamCount > config.maxTeams) {
    errors.push(`Select at most ${config.maxTeams} teams`);
  }

  if (budgetRemaining < 0) {
    errors.push(`Over budget by ${Math.abs(budgetRemaining).toFixed(2)} pts`);
  }

  Object.entries(bidsByTeamId).forEach(([teamId, bid]) => {
    if (bid < MIN_BID) {
      errors.push(`All bids must be at least ${MIN_BID} pts`);
    }
    if (bid > config.maxBid) {
      const team = teams.find((t) => t.id === teamId);
      errors.push(`Bid on ${team?.schoolName || 'team'} exceeds max ${config.maxBid} pts`);
    }
  });

  return errors;
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe('getSeedVariant', () => {
  it('returns "secondary" for uniform styling', () => {
    // GIVEN any seed
    // WHEN getting the variant
    // THEN it returns "secondary" (uniform styling for all seeds)
    expect(getSeedVariant()).toBe('secondary');
  });
});

describe('budget computation', () => {
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
});

describe('team count', () => {
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

describe('validation errors', () => {
  const defaultConfig: ValidationConfig = {
    minTeams: 3,
    maxTeams: 10,
    maxBid: 50,
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
    expect(errors).toContain('Over budget by 10.00 pts');
  });

  it('reports bid below minimum when a bid is less than 1', () => {
    // GIVEN a bid of 0 (below MIN_BID of 1)
    const bids = { t1: 0, t2: 10, t3: 10 };

    // WHEN computing validation errors
    const errors = computeValidationErrors(bids, defaultConfig, []);

    // THEN an error about minimum bid is present
    expect(errors).toContain('All bids must be at least 1 pts');
  });

  it('reports bid exceeds max with team name when available', () => {
    // GIVEN a bid of 60 exceeding maxBid of 50, with a known team name
    const bids = { t1: 60, t2: 10, t3: 10 };
    const teams: TeamLookup[] = [{ id: 't1', schoolName: 'Duke' }];

    // WHEN computing validation errors
    const errors = computeValidationErrors(bids, defaultConfig, teams);

    // THEN the error includes the team name
    expect(errors).toContain('Bid on Duke exceeds max 50 pts');
  });

  it('falls back to "team" when school name is not found for over-max bid', () => {
    // GIVEN a bid exceeding max for an unknown team
    const bids = { t1: 60, t2: 10, t3: 10 };

    // WHEN computing validation errors with no team lookup
    const errors = computeValidationErrors(bids, defaultConfig, []);

    // THEN the error falls back to "team"
    expect(errors).toContain('Bid on team exceeds max 50 pts');
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
});

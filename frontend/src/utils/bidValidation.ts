const MIN_BID = 1;

export interface BidValidationConfig {
  minTeams: number;
  maxTeams: number;
  maxBidPoints: number;
  budget: number;
}

export interface BidTeamLookup {
  id: string;
  schoolName: string;
}

export function computeBudgetRemaining(bidsByTeamId: Record<string, number>, budget: number): number {
  const spent = Object.values(bidsByTeamId).reduce((sum, bid) => sum + bid, 0);
  return budget - spent;
}

export function computeTeamCount(bidsByTeamId: Record<string, number>): number {
  return Object.keys(bidsByTeamId).length;
}

export function computeValidationErrors(
  bidsByTeamId: Record<string, number>,
  config: BidValidationConfig,
  teams: BidTeamLookup[],
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
    errors.push(`Over budget by ${Math.abs(budgetRemaining).toFixed(2)} credits`);
  }

  Object.entries(bidsByTeamId).forEach(([teamId, bid]) => {
    if (bid < MIN_BID) {
      errors.push(`All bids must be at least ${MIN_BID} credit${MIN_BID === 1 ? '' : 's'}`);
    }
    if (bid > config.maxBidPoints) {
      const team = teams.find((t) => t.id === teamId);
      errors.push(`Bid on ${team?.schoolName || 'team'} exceeds max ${config.maxBidPoints} credits`);
    }
  });

  return errors;
}

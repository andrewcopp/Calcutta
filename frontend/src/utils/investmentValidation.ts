const MIN_INVESTMENT = 1;

export interface InvestmentValidationConfig {
  minTeams: number;
  maxTeams: number;
  maxInvestmentCredits: number;
  budget: number;
}

export interface InvestmentTeamLookup {
  id: string;
  schoolName: string;
}

export function computeBudgetRemaining(investmentsByTeamId: Record<string, number>, budget: number): number {
  const spent = Object.values(investmentsByTeamId).reduce((sum, amount) => sum + amount, 0);
  return budget - spent;
}

export function computeTeamCount(investmentsByTeamId: Record<string, number>): number {
  return Object.keys(investmentsByTeamId).length;
}

export function computeValidationErrors(
  investmentsByTeamId: Record<string, number>,
  config: InvestmentValidationConfig,
  teams: InvestmentTeamLookup[],
): string[] {
  const teamCount = computeTeamCount(investmentsByTeamId);
  const budgetRemaining = computeBudgetRemaining(investmentsByTeamId, config.budget);
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

  Object.entries(investmentsByTeamId).forEach(([teamId, amount]) => {
    if (amount < MIN_INVESTMENT) {
      errors.push(`All investments must be at least ${MIN_INVESTMENT} credit${MIN_INVESTMENT === 1 ? '' : 's'}`);
    }
    if (amount > config.maxInvestmentCredits) {
      const team = teams.find((t) => t.id === teamId);
      errors.push(`Investment in ${team?.schoolName || 'team'} exceeds max ${config.maxInvestmentCredits} credits`);
    }
  });

  return errors;
}

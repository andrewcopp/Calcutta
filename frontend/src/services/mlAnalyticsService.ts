import { apiClient } from '../api/apiClient';

export type OptimizationRun = {
  run_id: string;
  name: string;
  calcutta_id: string | null;
  strategy: string;
  n_sims: number;
  seed: number;
  budget_points: number;
  created_at: string;
};

export type OptimizationRunsResponse = {
  year: number;
  runs: OptimizationRun[];
};

export type EntryRanking = {
  rank: number;
  entry_key: string;
  is_our_strategy: boolean;
  n_teams: number;
  total_bid_points: number;
  mean_normalized_payout: number;
  percentile_rank: number;
  p_top1: number;
  p_in_money: number;
};

export type EntryRankingsResponse = {
  run_id: string;
  total_entries: number;
  limit: number;
  offset: number;
  entries: EntryRanking[];
};

export type EntryPortfolioTeam = {
  team_id: string;
  school_name: string;
  seed: number;
  region: string;
  bid_points: number;
};

export type EntryPortfolioResponse = {
  entry_key: string;
  teams: EntryPortfolioTeam[];
  total_bid: number;
  n_teams: number;
};

export type OurEntryDetailsResponse = {
  run: {
    run_id: string;
    name: string;
    calcutta_id: string | null;
    strategy: string;
    n_sims: number;
    seed: number;
    budget_points: number;
    created_at: string;
  };
  portfolio: Array<{
    team_id: string;
    school_name: string;
    seed: number;
    region: string;
    bid_points: number;
    expected_roi: number;
  }>;
  summary: {
    mean_normalized_payout: number;
    p_top1: number;
    p_in_money: number;
    percentile_rank: number;
  };
};

export type StrategyGenerationRun = {
  id: string;
  run_key: string;
  tournament_simulation_batch_id: string | null;
  calcutta_id: string;
  purpose: string;
  returns_model_key: string;
  investment_model_key: string;
  optimizer_key: string;
  params_json: unknown;
  git_sha: string | null;
  created_at: string;
};

export type StrategyGenerationRunsResponse = {
  calcutta_id: string;
  runs: StrategyGenerationRun[];
  count: number;
};

export type TeamPredictedReturns = {
  team_id: string;
  school_name: string;
  seed: number;
  region: string;
  prob_pi: number;
  prob_r64: number;
  prob_r32: number;
  prob_s16: number;
  prob_e8: number;
  prob_ff: number;
  prob_champ: number;
  expected_value: number;
};

export type CalcuttaPredictedReturnsResponse = {
  calcutta_id: string;
  strategy_generation_run_id: string | null;
  teams: TeamPredictedReturns[];
  count: number;
};

export type TeamPredictedInvestment = {
  team_id: string;
  school_name: string;
  seed: number;
  region: string;
  rational: number;
  predicted: number;
  delta: number;
};

export type CalcuttaPredictedInvestmentResponse = {
  calcutta_id: string;
  strategy_generation_run_id: string | null;
  teams: TeamPredictedInvestment[];
  count: number;
};

export const mlAnalyticsService = {
  async getOptimizationRuns(year: number): Promise<OptimizationRunsResponse> {
    return apiClient.get<OptimizationRunsResponse>(`/v1/analytics/tournaments/${year}/runs`);
  },

  async getEntryRankings(year: number, runId: string, limit = 100, offset = 0): Promise<EntryRankingsResponse> {
    const params = new URLSearchParams();
    params.set('limit', String(limit));
    params.set('offset', String(offset));
    return apiClient.get<EntryRankingsResponse>(
      `/v1/analytics/tournaments/${year}/runs/${encodeURIComponent(runId)}/rankings?${params.toString()}`
    );
  },

  async getEntryPortfolio(year: number, runId: string, entryKey: string): Promise<EntryPortfolioResponse> {
    return apiClient.get<EntryPortfolioResponse>(
      `/v1/analytics/tournaments/${year}/runs/${encodeURIComponent(runId)}/entries/${encodeURIComponent(entryKey)}/portfolio`
    );
  },

  async getOurEntryDetails(year: number, runId: string): Promise<OurEntryDetailsResponse> {
    return apiClient.get<OurEntryDetailsResponse>(
      `/v1/analytics/tournaments/${year}/runs/${encodeURIComponent(runId)}/our-entry`
    );
  },

  async listStrategyGenerationRuns(calcuttaId: string): Promise<StrategyGenerationRunsResponse> {
    return apiClient.get<StrategyGenerationRunsResponse>(`/analytics/calcuttas/${encodeURIComponent(calcuttaId)}/strategy-generation-runs`);
  },

  async getCalcuttaPredictedReturns(params: {
    calcuttaId: string;
    strategyGenerationRunId?: string;
  }): Promise<CalcuttaPredictedReturnsResponse> {
    const query = new URLSearchParams();
    if (params.strategyGenerationRunId) {
      query.set('strategy_generation_run_id', params.strategyGenerationRunId);
    }
    const suffix = query.toString() ? `?${query.toString()}` : '';
    return apiClient.get<CalcuttaPredictedReturnsResponse>(
      `/analytics/calcuttas/${encodeURIComponent(params.calcuttaId)}/predicted-returns${suffix}`
    );
  },

  async getCalcuttaPredictedInvestment(params: {
    calcuttaId: string;
    strategyGenerationRunId?: string;
  }): Promise<CalcuttaPredictedInvestmentResponse> {
    const query = new URLSearchParams();
    if (params.strategyGenerationRunId) {
      query.set('strategy_generation_run_id', params.strategyGenerationRunId);
    }
    const suffix = query.toString() ? `?${query.toString()}` : '';
    return apiClient.get<CalcuttaPredictedInvestmentResponse>(
      `/analytics/calcuttas/${encodeURIComponent(params.calcuttaId)}/predicted-investment${suffix}`
    );
  },
};

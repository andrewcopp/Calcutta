import { apiClient } from '../api/apiClient';

export type OptimizationRun = {
  run_id: string;
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
  bid_amount_points: number;
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
    recommended_bid_points: number;
    expected_roi: number;
  }>;
  summary: {
    mean_normalized_payout: number;
    p_top1: number;
    p_in_money: number;
    percentile_rank: number;
  };
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
};

// Shared utility types
export type SortDir = 'asc' | 'desc';

// Types for lab.investment_models
export type InvestmentModel = {
  id: string;
  name: string;
  kind: string;
  params_json: Record<string, unknown>;
  notes?: string | null;
  created_at: string;
  updated_at: string;
  n_entries: number;
  n_evaluations: number;
};

// Types for lab.entries - market predictions (what model thinks market will bid)
export type EnrichedPrediction = {
  team_id: string;
  school_name: string;
  seed: number;
  region: string;
  predicted_market_share: number;
  predicted_bid_points: number;
  expected_points: number;
  expected_roi: number;
  naive_points: number;
  edge_percent: number;
};

// Types for lab.entries - optimized bids (our response to predictions)
export type EnrichedBid = {
  team_id: string;
  school_name: string;
  seed: number;
  region: string;
  bid_points: number;
  naive_points: number;
  edge_percent: number;
  expected_roi?: number | null;
};

export type EntryDetail = {
  id: string;
  investment_model_id: string;
  calcutta_id: string;
  game_outcome_kind: string;
  game_outcome_params_json: Record<string, unknown>;
  optimizer_kind: string;
  optimizer_params_json: Record<string, unknown>;
  starting_state_key: string;
  has_predictions: boolean;
  predictions?: EnrichedPrediction[];
  bids: EnrichedBid[];
  created_at: string;
  updated_at: string;
  model_name: string;
  model_kind: string;
  calcutta_name: string;
  n_evaluations: number;
};

// Types for lab.evaluations
export type EvaluationDetail = {
  id: string;
  entry_id: string;
  n_sims: number;
  seed: number;
  mean_normalized_payout?: number | null;
  median_normalized_payout?: number | null;
  p_top1?: number | null;
  p_in_money?: number | null;
  our_rank?: number | null;
  simulated_calcutta_id?: string | null;
  created_at: string;
  updated_at: string;
  model_name: string;
  model_kind: string;
  calcutta_id: string;
  calcutta_name: string;
  starting_state_key: string;
};

export type ListEvaluationsResponse = {
  items: EvaluationDetail[];
};

// Types for evaluation entry results
export type EvaluationEntryResult = {
  id: string;
  entry_name: string;
  mean_normalized_payout?: number | null;
  p_top1?: number | null;
  p_in_money?: number | null;
  rank: number;
};

// Types for evaluation entry bid
export type EvaluationEntryBid = {
  team_id: string;
  school_name: string;
  seed: number;
  region: string;
  bid_points: number;
  ownership: number;
};

// Types for evaluation entry profile
export type EvaluationEntryProfile = {
  entry_name: string;
  mean_normalized_payout?: number | null;
  p_top1?: number | null;
  p_in_money?: number | null;
  rank: number;
  total_bid_points: number;
  bids: EvaluationEntryBid[];
};

// Types for leaderboard
export type LeaderboardEntry = {
  investment_model_id: string;
  model_name: string;
  model_kind: string;
  n_entries: number;
  n_entries_with_predictions: number;
  n_evaluations: number;
  n_calcuttas_with_entries: number;
  n_calcuttas_with_evaluations: number;
  avg_mean_payout?: number | null;
  avg_median_payout?: number | null;
  avg_p_top1?: number | null;
  avg_p_in_money?: number | null;
  first_eval_at?: string | null;
  last_eval_at?: string | null;
};

export type LeaderboardResponse = {
  items: LeaderboardEntry[];
};

// Types for pipeline
export type StartPipelineRequest = {
  calcutta_ids?: string[];
  budget_points?: number;
  optimizer_kind?: string;
  n_sims?: number;
  seed?: number;
  excluded_entry_name?: string;
  force_rerun?: boolean;
};

export type StartPipelineResponse = {
  pipeline_run_id: string;
  n_calcuttas: number;
  status: string;
};

export type CalcuttaPipelineStatus = {
  calcutta_id: string;
  calcutta_name: string;
  calcutta_year: number;
  stage: string;
  status: string;
  progress: number;
  progress_message?: string | null;
  has_predictions: boolean;
  has_entry: boolean;
  has_evaluation: boolean;
  entry_id?: string | null;
  evaluation_id?: string | null;
  mean_payout?: number | null;
  our_rank?: number | null;
  error_message?: string | null;
};

export type ModelPipelineProgress = {
  model_id: string;
  model_name: string;
  active_pipeline_run_id?: string | null;
  total_calcuttas: number;
  predictions_count: number;
  entries_count: number;
  evaluations_count: number;
  avg_mean_payout?: number | null;
  calcuttas: CalcuttaPipelineStatus[];
};

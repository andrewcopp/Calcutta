import { apiClient } from '../api/apiClient';

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

export type ListModelsResponse = {
  items: InvestmentModel[];
};

// Types for lab.entries (enriched with team data and naive allocation)
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
  bids: EnrichedBid[];
  created_at: string;
  updated_at: string;
  model_name: string;
  model_kind: string;
  calcutta_name: string;
  n_evaluations: number;
};

export type ListEntriesResponse = {
  items: EntryDetail[];
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

// Types for leaderboard
export type LeaderboardEntry = {
  investment_model_id: string;
  model_name: string;
  model_kind: string;
  n_entries: number;
  n_evaluations: number;
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

// Types for generate entries
export type GenerateEntriesRequest = {
  years?: number[];
  budget_points?: number;
  excluded_entry?: string;
};

export type GenerateEntriesResponse = {
  entries_created: number;
  errors?: string[];
};

// Service
export const labService = {
  async listModels(params?: {
    kind?: string;
    limit?: number;
    offset?: number;
  }): Promise<ListModelsResponse> {
    const q = new URLSearchParams();
    if (params?.kind) q.set('kind', params.kind);
    if (params?.limit != null) q.set('limit', String(params.limit));
    if (params?.offset != null) q.set('offset', String(params.offset));
    const suffix = q.toString() ? `?${q.toString()}` : '';
    return apiClient.get<ListModelsResponse>(`/lab/models${suffix}`);
  },

  async getModel(id: string): Promise<InvestmentModel> {
    return apiClient.get<InvestmentModel>(`/lab/models/${encodeURIComponent(id)}`);
  },

  async getLeaderboard(): Promise<LeaderboardResponse> {
    return apiClient.get<LeaderboardResponse>('/lab/models/leaderboard');
  },

  async listEntries(params?: {
    investment_model_id?: string;
    calcutta_id?: string;
    starting_state_key?: string;
    limit?: number;
    offset?: number;
  }): Promise<ListEntriesResponse> {
    const q = new URLSearchParams();
    if (params?.investment_model_id) q.set('investment_model_id', params.investment_model_id);
    if (params?.calcutta_id) q.set('calcutta_id', params.calcutta_id);
    if (params?.starting_state_key) q.set('starting_state_key', params.starting_state_key);
    if (params?.limit != null) q.set('limit', String(params.limit));
    if (params?.offset != null) q.set('offset', String(params.offset));
    const suffix = q.toString() ? `?${q.toString()}` : '';
    return apiClient.get<ListEntriesResponse>(`/lab/entries${suffix}`);
  },

  async getEntry(id: string): Promise<EntryDetail> {
    return apiClient.get<EntryDetail>(`/lab/entries/${encodeURIComponent(id)}`);
  },

  async getEntryByModelAndCalcutta(
    modelName: string,
    calcuttaId: string,
    startingStateKey?: string
  ): Promise<EntryDetail> {
    const q = new URLSearchParams();
    if (startingStateKey) q.set('starting_state_key', startingStateKey);
    const suffix = q.toString() ? `?${q.toString()}` : '';
    return apiClient.get<EntryDetail>(
      `/lab/models/${encodeURIComponent(modelName)}/calcutta/${encodeURIComponent(calcuttaId)}/entry${suffix}`
    );
  },

  async listEvaluations(params?: {
    entry_id?: string;
    investment_model_id?: string;
    calcutta_id?: string;
    limit?: number;
    offset?: number;
  }): Promise<ListEvaluationsResponse> {
    const q = new URLSearchParams();
    if (params?.entry_id) q.set('entry_id', params.entry_id);
    if (params?.investment_model_id) q.set('investment_model_id', params.investment_model_id);
    if (params?.calcutta_id) q.set('calcutta_id', params.calcutta_id);
    if (params?.limit != null) q.set('limit', String(params.limit));
    if (params?.offset != null) q.set('offset', String(params.offset));
    const suffix = q.toString() ? `?${q.toString()}` : '';
    return apiClient.get<ListEvaluationsResponse>(`/lab/evaluations${suffix}`);
  },

  async getEvaluation(id: string): Promise<EvaluationDetail> {
    return apiClient.get<EvaluationDetail>(`/lab/evaluations/${encodeURIComponent(id)}`);
  },

  async generateEntries(
    modelId: string,
    request?: GenerateEntriesRequest
  ): Promise<GenerateEntriesResponse> {
    return apiClient.post<GenerateEntriesResponse>(
      `/lab/models/${encodeURIComponent(modelId)}/generate-entries`,
      request ?? {}
    );
  },
};

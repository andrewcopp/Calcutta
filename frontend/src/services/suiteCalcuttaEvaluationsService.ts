import { apiClient } from '../api/apiClient';

export type SuiteCalcuttaEvaluation = {
  id: string;
  suite_id: string;
  suite_name: string;
  optimizer_key: string;
  n_sims: number;
  seed: number;
  calcutta_id: string;
  game_outcome_run_id?: string | null;
  market_share_run_id?: string | null;
  strategy_generation_run_id?: string | null;
  calcutta_evaluation_run_id?: string | null;
  starting_state_key: string;
  excluded_entry_name?: string | null;
  status: string;
  claimed_at?: string | null;
  claimed_by?: string | null;
  error_message?: string | null;
  created_at: string;
  updated_at: string;
};

export type ListSuiteCalcuttaEvaluationsResponse = {
  items: SuiteCalcuttaEvaluation[];
};

export const suiteCalcuttaEvaluationsService = {
  async list(params?: {
    suiteId?: string;
    calcuttaId?: string;
    limit?: number;
    offset?: number;
  }): Promise<ListSuiteCalcuttaEvaluationsResponse> {
    const q = new URLSearchParams();
    if (params?.suiteId) q.set('suite_id', params.suiteId);
    if (params?.calcuttaId) q.set('calcutta_id', params.calcuttaId);
    if (params?.limit != null) q.set('limit', String(params.limit));
    if (params?.offset != null) q.set('offset', String(params.offset));

    const suffix = q.toString() ? `?${q.toString()}` : '';
    return apiClient.get<ListSuiteCalcuttaEvaluationsResponse>(`/suite-calcutta-evaluations${suffix}`);
  },

  async get(id: string): Promise<SuiteCalcuttaEvaluation> {
    return apiClient.get<SuiteCalcuttaEvaluation>(`/suite-calcutta-evaluations/${encodeURIComponent(id)}`);
  },
};

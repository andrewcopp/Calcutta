import { apiClient } from '../api/apiClient';

export type SuiteListItem = {
  id: string;
  name: string;
  description?: string | null;
  game_outcomes_algorithm_id: string;
  market_share_algorithm_id: string;
  optimizer_key: string;
  n_sims: number;
  seed: number;
  latest_execution_id?: string | null;
  latest_execution_name?: string | null;
  latest_execution_status?: string | null;
  latest_execution_created_at?: string | null;
  latest_execution_updated_at?: string | null;
  created_at: string;
  updated_at: string;
};

export type ListSuitesResponse = {
  items: SuiteListItem[];
};

export const suitesService = {
  async list(params?: { limit?: number; offset?: number }): Promise<ListSuitesResponse> {
    const q = new URLSearchParams();
    if (params?.limit != null) q.set('limit', String(params.limit));
    if (params?.offset != null) q.set('offset', String(params.offset));

    const suffix = q.toString() ? `?${q.toString()}` : '';
    return apiClient.get<ListSuitesResponse>(`/suites${suffix}`);
  },

  async get(id: string): Promise<SuiteListItem> {
    return apiClient.get<SuiteListItem>(`/suites/${encodeURIComponent(id)}`);
  },
};

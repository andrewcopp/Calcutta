import { apiClient } from '../api/apiClient';

export type CohortListItem = {
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

export type ListCohortsResponse = {
  items: CohortListItem[];
};

export const cohortsService = {
  async list(params?: { limit?: number; offset?: number }): Promise<ListCohortsResponse> {
    const q = new URLSearchParams();
    if (params?.limit != null) q.set('limit', String(params.limit));
    if (params?.offset != null) q.set('offset', String(params.offset));

    const suffix = q.toString() ? `?${q.toString()}` : '';
    return apiClient.get<ListCohortsResponse>(`/cohorts${suffix}`);
  },

  async get(id: string): Promise<CohortListItem> {
    return apiClient.get<CohortListItem>(`/cohorts/${encodeURIComponent(id)}`);
  },
};

import { apiClient } from '../api/apiClient';

export type SuiteExecutionListItem = {
  id: string;
  suite_id: string;
  suite_name: string;
  name?: string | null;
  optimizer_key?: string | null;
  n_sims?: number | null;
  seed?: number | null;
  starting_state_key: string;
  excluded_entry_name?: string | null;
  status: string;
  error_message?: string | null;
  created_at: string;
  updated_at: string;
};

export type ListSuiteExecutionsResponse = {
  items: SuiteExecutionListItem[];
};

export const suiteExecutionsService = {
  async list(params?: { suiteId?: string; limit?: number; offset?: number }): Promise<ListSuiteExecutionsResponse> {
    const q = new URLSearchParams();
    if (params?.suiteId) q.set('cohort_id', params.suiteId);
    if (params?.limit != null) q.set('limit', String(params.limit));
    if (params?.offset != null) q.set('offset', String(params.offset));

    const suffix = q.toString() ? `?${q.toString()}` : '';
    return apiClient.get<ListSuiteExecutionsResponse>(`/simulation-run-batches${suffix}`);
  },

  async get(id: string): Promise<SuiteExecutionListItem> {
    return apiClient.get<SuiteExecutionListItem>(`/simulation-run-batches/${encodeURIComponent(id)}`);
  },
};

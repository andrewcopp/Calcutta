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

export type SimulationRunBatchListItem = SuiteExecutionListItem;

export type ListSimulationRunBatchesResponse = ListSuiteExecutionsResponse;

export const simulationRunBatchesService = {
  async list(params: { cohortId: string; limit?: number; offset?: number }): Promise<ListSimulationRunBatchesResponse> {
    const q = new URLSearchParams();
    if (params?.limit != null) q.set('limit', String(params.limit));
    if (params?.offset != null) q.set('offset', String(params.offset));

    const suffix = q.toString() ? `?${q.toString()}` : '';
    return apiClient.get<ListSimulationRunBatchesResponse>(
      `/cohorts/${encodeURIComponent(params.cohortId)}/simulation-batches${suffix}`
    );
  },

  async get(params: { cohortId: string; id: string }): Promise<SimulationRunBatchListItem> {
    return apiClient.get<SimulationRunBatchListItem>(
      `/cohorts/${encodeURIComponent(params.cohortId)}/simulation-batches/${encodeURIComponent(params.id)}`
    );
  },
};

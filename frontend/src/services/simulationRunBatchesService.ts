import { apiClient } from '../api/apiClient';

export type SimulationRunBatch = {
  id: string;
  cohort_id: string;
  cohort_name: string;
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

export type ListSimulationRunBatchesResponse = {
  items: SimulationRunBatch[];
};

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

  async get(params: { cohortId: string; id: string }): Promise<SimulationRunBatch> {
    return apiClient.get<SimulationRunBatch>(
      `/cohorts/${encodeURIComponent(params.cohortId)}/simulation-batches/${encodeURIComponent(params.id)}`
    );
  },
};

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
  starting_state_key?: string;
  excluded_entry_name?: string | null;
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

export type CreateCohortRequest = {
  name: string;
  description?: string | null;
	gameOutcomesAlgorithmId?: string;
	marketShareAlgorithmId?: string;
	optimizerKey?: string;
	nSims?: number;
	seed?: number;
	startingStateKey?: string;
	excludedEntryName?: string;
};

export type PatchCohortRequest = {
	optimizerKey?: string | null;
	nSims?: number | null;
	seed?: number | null;
	startingStateKey?: string | null;
	excludedEntryName?: string | null;
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

  async create(req: CreateCohortRequest): Promise<CohortListItem> {
    return apiClient.post<CohortListItem>('/cohorts', req);
  },

	async patch(id: string, req: PatchCohortRequest): Promise<CohortListItem> {
		return apiClient.patch<CohortListItem>(`/cohorts/${encodeURIComponent(id)}`, req);
	},
};

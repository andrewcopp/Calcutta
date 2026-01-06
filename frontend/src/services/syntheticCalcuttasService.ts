import { apiClient } from '../api/apiClient';

export type SyntheticCalcuttaListItem = {
  id: string;
  suite_id: string;
  calcutta_id: string;
  calcutta_snapshot_id?: string | null;
  focus_strategy_generation_run_id?: string | null;
  focus_entry_name?: string | null;
  starting_state_key?: string | null;
  excluded_entry_name?: string | null;
  created_at: string;
  updated_at: string;
};

export type ListSyntheticCalcuttasResponse = {
  items: SyntheticCalcuttaListItem[];
};

export type CreateSyntheticCalcuttaRequest = {
  cohortId: string;
  calcuttaId: string;
  calcuttaSnapshotId?: string;
  focusStrategyGenerationRunId?: string;
  focusEntryName?: string;
  startingStateKey?: string;
  excludedEntryName?: string;
};

export type CreateSyntheticCalcuttaResponse = {
  id: string;
};

export const syntheticCalcuttasService = {
  async list(params?: { cohortId?: string; calcuttaId?: string; limit?: number; offset?: number }): Promise<ListSyntheticCalcuttasResponse> {
    const q = new URLSearchParams();
    if (params?.cohortId) q.set('cohort_id', params.cohortId);
    if (params?.calcuttaId) q.set('calcutta_id', params.calcuttaId);
    if (params?.limit != null) q.set('limit', String(params.limit));
    if (params?.offset != null) q.set('offset', String(params.offset));

    const suffix = q.toString() ? `?${q.toString()}` : '';
    return apiClient.get<ListSyntheticCalcuttasResponse>(`/synthetic-calcuttas${suffix}`);
  },

  async get(id: string): Promise<SyntheticCalcuttaListItem> {
    return apiClient.get<SyntheticCalcuttaListItem>(`/synthetic-calcuttas/${encodeURIComponent(id)}`);
  },

  async create(req: CreateSyntheticCalcuttaRequest): Promise<CreateSyntheticCalcuttaResponse> {
    return apiClient.post<CreateSyntheticCalcuttaResponse>('/synthetic-calcuttas', req);
  },
};

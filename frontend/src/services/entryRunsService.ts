import { apiClient } from '../api/apiClient';

export type EntryRunListItem = {
  id: string;
  run_key?: string | null;
  name?: string | null;
  calcutta_id?: string | null;
  simulated_tournament_id?: string | null;
  purpose: string;
  returns_model_key: string;
  investment_model_key: string;
  optimizer_key: string;
  created_at: string;
};

export type ListEntryRunsResponse = {
  items: EntryRunListItem[];
};

export type EntryRunArtifact = {
  id: string;
  run_id: string;
  run_key?: string | null;
  artifact_kind: string;
  schema_version: string;
  storage_uri?: string | null;
  summary_json: unknown;
  created_at: string;
  updated_at: string;
};

export const entryRunsService = {
  async list(params?: { calcuttaId?: string; limit?: number; offset?: number }): Promise<ListEntryRunsResponse> {
    const q = new URLSearchParams();
    if (params?.calcuttaId) q.set('calcutta_id', params.calcuttaId);
    if (params?.limit != null) q.set('limit', String(params.limit));
    if (params?.offset != null) q.set('offset', String(params.offset));

    const suffix = q.toString() ? `?${q.toString()}` : '';
    return apiClient.get<ListEntryRunsResponse>(`/entry-runs${suffix}`);
  },

  async get(id: string): Promise<EntryRunListItem> {
    return apiClient.get<EntryRunListItem>(`/entry-runs/${encodeURIComponent(id)}`);
  },

  async getArtifact(entryRunId: string, artifactKind: string): Promise<EntryRunArtifact> {
    return apiClient.get<EntryRunArtifact>(
      `/entry-runs/${encodeURIComponent(entryRunId)}/artifacts/${encodeURIComponent(artifactKind)}`
    );
  },
};

import { apiClient } from '../api/apiClient';

export type EntryArtifact = {
  id: string;
  run_id: string;
  run_key?: string | null;
  artifact_kind: string;
  schema_version: string;
  storage_uri?: string | null;
  summary_json: unknown;
  input_market_share_artifact_id?: string | null;
  input_advancement_artifact_id?: string | null;
  created_at: string;
  updated_at: string;
};

export const entryArtifactsService = {
  async get(id: string): Promise<EntryArtifact> {
    return apiClient.get<EntryArtifact>(`/entry-artifacts/${encodeURIComponent(id)}`);
  },
};

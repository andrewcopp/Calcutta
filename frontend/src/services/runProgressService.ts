import { apiClient } from '../api/apiClient';

export type RunProgressEvent = {
  id: string;
  event_kind: string;
  status?: string | null;
  percent?: number | null;
  phase?: string | null;
  message?: string | null;
  source: string;
  payload_json: unknown;
  created_at: string;
};

export type RunProgressResponse = {
  run_kind: string;
  run_id: string;
  run_key?: string | null;
  status: string;
  attempt: number;
  params_json: unknown;
  progress_json: unknown;
  progress_updated_at?: string | null;
  claimed_at?: string | null;
  claimed_by?: string | null;
  started_at?: string | null;
  finished_at?: string | null;
  error_message?: string | null;
  created_at: string;
  updated_at: string;
  events: RunProgressEvent[];
};

export const runProgressService = {
  async get(runKind: string, runId: string): Promise<RunProgressResponse> {
    return apiClient.get<RunProgressResponse>(
      `/runs/${encodeURIComponent(runKind)}/${encodeURIComponent(runId)}/progress`
    );
  },
};

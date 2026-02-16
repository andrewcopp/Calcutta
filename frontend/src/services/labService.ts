import { apiClient } from '../api/apiClient';

import type {
  InvestmentModel,
  ListModelsResponse,
  ListEntriesResponse,
  EntryDetail,
  ListEvaluationsResponse,
  EvaluationDetail,
  EvaluationEntryResult,
  EvaluationEntryProfile,
  GenerateEntriesRequest,
  GenerateEntriesResponse,
  StartPipelineRequest,
  StartPipelineResponse,
  ModelPipelineProgress,
  PipelineProgressResponse,
  LeaderboardResponse,
} from '../types/lab';

// Service
export const labService = {
  async listModels(params?: {
    kind?: string;
    limit?: number;
    offset?: number;
  }): Promise<ListModelsResponse> {
    const q = new URLSearchParams();
    if (params?.kind) q.set('kind', params.kind);
    if (params?.limit != null) q.set('limit', String(params.limit));
    if (params?.offset != null) q.set('offset', String(params.offset));
    const suffix = q.toString() ? `?${q.toString()}` : '';
    return apiClient.get<ListModelsResponse>(`/lab/models${suffix}`);
  },

  async getModel(id: string): Promise<InvestmentModel> {
    return apiClient.get<InvestmentModel>(`/lab/models/${encodeURIComponent(id)}`);
  },

  async getLeaderboard(): Promise<LeaderboardResponse> {
    return apiClient.get<LeaderboardResponse>('/lab/models/leaderboard');
  },

  async listEntries(params?: {
    investment_model_id?: string;
    calcutta_id?: string;
    starting_state_key?: string;
    limit?: number;
    offset?: number;
  }): Promise<ListEntriesResponse> {
    const q = new URLSearchParams();
    if (params?.investment_model_id) q.set('investment_model_id', params.investment_model_id);
    if (params?.calcutta_id) q.set('calcutta_id', params.calcutta_id);
    if (params?.starting_state_key) q.set('starting_state_key', params.starting_state_key);
    if (params?.limit != null) q.set('limit', String(params.limit));
    if (params?.offset != null) q.set('offset', String(params.offset));
    const suffix = q.toString() ? `?${q.toString()}` : '';
    return apiClient.get<ListEntriesResponse>(`/lab/entries${suffix}`);
  },

  async getEntry(id: string): Promise<EntryDetail> {
    return apiClient.get<EntryDetail>(`/lab/entries/${encodeURIComponent(id)}`);
  },

  async getEntryByModelAndCalcutta(
    modelName: string,
    calcuttaId: string,
    startingStateKey?: string
  ): Promise<EntryDetail> {
    const q = new URLSearchParams();
    if (startingStateKey) q.set('starting_state_key', startingStateKey);
    const suffix = q.toString() ? `?${q.toString()}` : '';
    return apiClient.get<EntryDetail>(
      `/lab/models/${encodeURIComponent(modelName)}/calcutta/${encodeURIComponent(calcuttaId)}/entry${suffix}`
    );
  },

  async listEvaluations(params?: {
    entry_id?: string;
    investment_model_id?: string;
    calcutta_id?: string;
    limit?: number;
    offset?: number;
  }): Promise<ListEvaluationsResponse> {
    const q = new URLSearchParams();
    if (params?.entry_id) q.set('entry_id', params.entry_id);
    if (params?.investment_model_id) q.set('investment_model_id', params.investment_model_id);
    if (params?.calcutta_id) q.set('calcutta_id', params.calcutta_id);
    if (params?.limit != null) q.set('limit', String(params.limit));
    if (params?.offset != null) q.set('offset', String(params.offset));
    const suffix = q.toString() ? `?${q.toString()}` : '';
    return apiClient.get<ListEvaluationsResponse>(`/lab/evaluations${suffix}`);
  },

  async getEvaluation(id: string): Promise<EvaluationDetail> {
    return apiClient.get<EvaluationDetail>(`/lab/evaluations/${encodeURIComponent(id)}`);
  },

  async getEvaluationEntryResults(id: string): Promise<EvaluationEntryResult[]> {
    const response = await apiClient.get<{ items: EvaluationEntryResult[] }>(
      `/lab/evaluations/${encodeURIComponent(id)}/entries`
    );
    return response.items;
  },

  async getEvaluationEntryProfile(entryResultId: string): Promise<EvaluationEntryProfile> {
    return apiClient.get<EvaluationEntryProfile>(
      `/lab/entry-results/${encodeURIComponent(entryResultId)}`
    );
  },

  async generateEntries(
    modelId: string,
    request?: GenerateEntriesRequest
  ): Promise<GenerateEntriesResponse> {
    return apiClient.post<GenerateEntriesResponse>(
      `/lab/models/${encodeURIComponent(modelId)}/generate-entries`,
      request ?? {}
    );
  },

  async startPipeline(
    modelId: string,
    request?: StartPipelineRequest
  ): Promise<StartPipelineResponse> {
    return apiClient.post<StartPipelineResponse>(
      `/lab/models/${encodeURIComponent(modelId)}/pipeline/start`,
      request ?? {}
    );
  },

  async getModelPipelineProgress(modelId: string): Promise<ModelPipelineProgress> {
    return apiClient.get<ModelPipelineProgress>(
      `/lab/models/${encodeURIComponent(modelId)}/pipeline/progress`
    );
  },

  async getPipelineRunProgress(pipelineRunId: string): Promise<PipelineProgressResponse> {
    return apiClient.get<PipelineProgressResponse>(
      `/lab/pipeline-runs/${encodeURIComponent(pipelineRunId)}`
    );
  },

  async cancelPipeline(pipelineRunId: string): Promise<void> {
    await apiClient.post<{ status: string }>(
      `/lab/pipeline-runs/${encodeURIComponent(pipelineRunId)}/cancel`,
      {}
    );
  },
};

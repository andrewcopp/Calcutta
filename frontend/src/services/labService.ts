import { apiClient } from '../api/apiClient';

import type {
  InvestmentModel,
  EntryDetail,
  ListEvaluationsResponse,
  EvaluationDetail,
  EvaluationEntryResult,
  EvaluationEntryProfile,
  StartPipelineRequest,
  StartPipelineResponse,
  ModelPipelineProgress,
  LeaderboardResponse,
} from '../types/lab';

// Service
export const labService = {
  async getModel(id: string): Promise<InvestmentModel> {
    return apiClient.get<InvestmentModel>(`/lab/models/${encodeURIComponent(id)}`);
  },

  async getLeaderboard(): Promise<LeaderboardResponse> {
    return apiClient.get<LeaderboardResponse>('/lab/models/leaderboard');
  },

  async getEntryByModelAndCalcutta(
    modelName: string,
    calcuttaId: string,
    startingStateKey?: string,
  ): Promise<EntryDetail> {
    const q = new URLSearchParams();
    if (startingStateKey) q.set('starting_state_key', startingStateKey);
    const suffix = q.toString() ? `?${q.toString()}` : '';
    return apiClient.get<EntryDetail>(
      `/lab/models/${encodeURIComponent(modelName)}/calcutta/${encodeURIComponent(calcuttaId)}/entry${suffix}`,
    );
  },

  async listEvaluations(params?: {
    entryId?: string;
    investmentModelId?: string;
    calcuttaId?: string;
    limit?: number;
    offset?: number;
  }): Promise<ListEvaluationsResponse> {
    const q = new URLSearchParams();
    if (params?.entryId) q.set('entry_id', params.entryId);
    if (params?.investmentModelId) q.set('investment_model_id', params.investmentModelId);
    if (params?.calcuttaId) q.set('calcutta_id', params.calcuttaId);
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
      `/lab/evaluations/${encodeURIComponent(id)}/entries`,
    );
    return response.items;
  },

  async getEvaluationEntryProfile(entryResultId: string): Promise<EvaluationEntryProfile> {
    return apiClient.get<EvaluationEntryProfile>(`/lab/entry-results/${encodeURIComponent(entryResultId)}`);
  },

  async startPipeline(modelId: string, request?: StartPipelineRequest): Promise<StartPipelineResponse> {
    return apiClient.post<StartPipelineResponse>(
      `/lab/models/${encodeURIComponent(modelId)}/pipeline/start`,
      request ?? {},
    );
  },

  async getModelPipelineProgress(modelId: string): Promise<ModelPipelineProgress> {
    return apiClient.get<ModelPipelineProgress>(`/lab/models/${encodeURIComponent(modelId)}/pipeline/progress`);
  },

  async cancelPipeline(pipelineRunId: string): Promise<void> {
    await apiClient.post<{ status: string }>(`/lab/pipeline-runs/${encodeURIComponent(pipelineRunId)}/cancel`, {});
  },
};

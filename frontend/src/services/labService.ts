import { apiClient } from '../api/apiClient';
import type { StartPipelineRequest } from '../schemas/lab';
import {
  InvestmentModelSchema,
  LeaderboardResponseSchema,
  EntryDetailSchema,
  ListEvaluationsResponseSchema,
  EvaluationDetailSchema,
  EvaluationEntryResultsResponseSchema,
  EvaluationEntryProfileSchema,
  EvaluationSummarySchema,
  StartPipelineResponseSchema,
  ModelPipelineProgressSchema,
} from '../schemas/lab';

export const labService = {
  async getModel(id: string) {
    return apiClient.get(`/lab/models/${encodeURIComponent(id)}`, { schema: InvestmentModelSchema });
  },

  async getLeaderboard() {
    return apiClient.get('/lab/models/leaderboard', { schema: LeaderboardResponseSchema });
  },

  async getEntryByModelAndCalcutta(modelName: string, calcuttaId: string, startingStateKey?: string) {
    const q = new URLSearchParams();
    if (startingStateKey) q.set('starting_state_key', startingStateKey);
    const suffix = q.toString() ? `?${q.toString()}` : '';
    return apiClient.get(
      `/lab/models/${encodeURIComponent(modelName)}/calcutta/${encodeURIComponent(calcuttaId)}/entry${suffix}`,
      { schema: EntryDetailSchema },
    );
  },

  async listEvaluations(params?: {
    entryId?: string;
    investmentModelId?: string;
    calcuttaId?: string;
    limit?: number;
    offset?: number;
  }) {
    const q = new URLSearchParams();
    if (params?.entryId) q.set('entry_id', params.entryId);
    if (params?.investmentModelId) q.set('investment_model_id', params.investmentModelId);
    if (params?.calcuttaId) q.set('calcutta_id', params.calcuttaId);
    if (params?.limit != null) q.set('limit', String(params.limit));
    if (params?.offset != null) q.set('offset', String(params.offset));
    const suffix = q.toString() ? `?${q.toString()}` : '';
    return apiClient.get(`/lab/evaluations${suffix}`, { schema: ListEvaluationsResponseSchema });
  },

  async getEvaluation(id: string) {
    return apiClient.get(`/lab/evaluations/${encodeURIComponent(id)}`, { schema: EvaluationDetailSchema });
  },

  async getEvaluationSummary(evaluationId: string) {
    return apiClient.get(`/lab/evaluations/${encodeURIComponent(evaluationId)}/summary`, {
      schema: EvaluationSummarySchema,
    });
  },

  async getEvaluationEntryResults(id: string) {
    const response = await apiClient.get(`/lab/evaluations/${encodeURIComponent(id)}/entries`, {
      schema: EvaluationEntryResultsResponseSchema,
    });
    return response.items;
  },

  async getEvaluationEntryProfile(entryResultId: string) {
    return apiClient.get(`/lab/entry-results/${encodeURIComponent(entryResultId)}`, {
      schema: EvaluationEntryProfileSchema,
    });
  },

  async startPipeline(modelId: string, request?: StartPipelineRequest) {
    return apiClient.post(
      `/lab/models/${encodeURIComponent(modelId)}/pipeline/start`,
      request ?? {},
      { schema: StartPipelineResponseSchema },
    );
  },

  async getModelPipelineProgress(modelId: string) {
    return apiClient.get(`/lab/models/${encodeURIComponent(modelId)}/pipeline/progress`, {
      schema: ModelPipelineProgressSchema,
    });
  },

  async cancelPipeline(pipelineRunId: string): Promise<void> {
    await apiClient.post<{ status: string }>(`/lab/pipeline-runs/${encodeURIComponent(pipelineRunId)}/cancel`, {});
  },
};

import { apiClient } from '../api/apiClient';

export type SuiteCalcuttaEvaluation = {
  id: string;
  suite_execution_id?: string | null;
  suite_id: string;
  suite_name: string;
  optimizer_key: string;
  n_sims: number;
  seed: number;
  our_rank?: number | null;
  our_mean_normalized_payout?: number | null;
  our_median_normalized_payout?: number | null;
  our_p_top1?: number | null;
  our_p_in_money?: number | null;
  total_simulations?: number | null;
  calcutta_id: string;
  game_outcome_run_id?: string | null;
  market_share_run_id?: string | null;
  strategy_generation_run_id?: string | null;
  calcutta_evaluation_run_id?: string | null;
  realized_finish_position?: number | null;
  realized_is_tied?: boolean | null;
  realized_in_the_money?: boolean | null;
  realized_payout_cents?: number | null;
  realized_total_points?: number | null;
  starting_state_key: string;
  excluded_entry_name?: string | null;
  status: string;
  claimed_at?: string | null;
  claimed_by?: string | null;
  error_message?: string | null;
  created_at: string;
  updated_at: string;
};

export type SimulationRun = SuiteCalcuttaEvaluation;

export type ListSuiteCalcuttaEvaluationsResponse = {
  items: SuiteCalcuttaEvaluation[];
};

export type ListSimulationRunsResponse = ListSuiteCalcuttaEvaluationsResponse;

export type SuiteCalcuttaEvaluationPortfolioBid = {
  team_id: string;
  school_name: string;
  seed: number;
  region: string;
  bid_points: number;
  expected_roi: number;
};

export type SuiteCalcuttaEvaluationOurStrategyPerformance = {
  rank: number;
  entry_name: string;
  mean_normalized_payout: number;
  median_normalized_payout: number;
  p_top1: number;
  p_in_money: number;
  total_simulations: number;
};

export type SuiteCalcuttaEvaluationEntryPerformance = {
  rank: number;
  entry_name: string;
  snapshot_entry_id?: string | null;
  mean_normalized_payout: number;
  p_top1: number;
  p_in_money: number;
  finish_position?: number | null;
  is_tied?: boolean | null;
  in_the_money?: boolean | null;
  payout_cents?: number | null;
  total_points?: number | null;
};

export type SuiteCalcuttaEvaluationResult = {
  evaluation: SuiteCalcuttaEvaluation;
  portfolio: SuiteCalcuttaEvaluationPortfolioBid[];
  our_strategy?: SuiteCalcuttaEvaluationOurStrategyPerformance | null;
  entries: SuiteCalcuttaEvaluationEntryPerformance[];
};

export type SimulationRunResult = SuiteCalcuttaEvaluationResult;

export type SuiteCalcuttaEvaluationSnapshotEntryTeam = {
  team_id: string;
  school_name: string;
  seed: number;
  region: string;
  bid_points: number;
};

export type SuiteCalcuttaEvaluationSnapshotEntryResponse = {
  snapshot_entry_id: string;
  display_name: string;
  is_synthetic: boolean;
  teams: SuiteCalcuttaEvaluationSnapshotEntryTeam[];
};

export type CreateSuiteCalcuttaEvaluationRequest = {
  calcuttaId: string;
  suiteId?: string;
  suiteName?: string;
  optimizerKey?: string;
  gameOutcomeRunId: string;
  marketShareRunId: string;
  nSims: number;
  seed: number;
  startingStateKey?: string;
  excludedEntryName?: string;
};

export type CreateSuiteCalcuttaEvaluationResponse = {
  id: string;
  status: string;
};

export type CreateSimulationRunResponse = CreateSuiteCalcuttaEvaluationResponse;

export const simulationRunsService = {
  async list(params: {
    cohortId: string;
    simulationBatchId?: string;
    calcuttaId?: string;
    limit?: number;
    offset?: number;
  }): Promise<ListSimulationRunsResponse> {
    const q = new URLSearchParams();
    if (params?.simulationBatchId) q.set('simulation_batch_id', params.simulationBatchId);
    if (params?.calcuttaId) q.set('calcutta_id', params.calcuttaId);
    if (params?.limit != null) q.set('limit', String(params.limit));
    if (params?.offset != null) q.set('offset', String(params.offset));

    const suffix = q.toString() ? `?${q.toString()}` : '';
    return apiClient.get<ListSimulationRunsResponse>(
      `/cohorts/${encodeURIComponent(params.cohortId)}/simulations${suffix}`
    );
  },

  async get(params: { cohortId: string; id: string }): Promise<SimulationRun> {
    return apiClient.get<SimulationRun>(
      `/cohorts/${encodeURIComponent(params.cohortId)}/simulations/${encodeURIComponent(params.id)}`
    );
  },

  async getResult(params: { cohortId: string; id: string }): Promise<SimulationRunResult> {
    return apiClient.get<SimulationRunResult>(
      `/cohorts/${encodeURIComponent(params.cohortId)}/simulations/${encodeURIComponent(params.id)}/result`
    );
  },

  async getSnapshotEntry(params: {
    cohortId: string;
    id: string;
    snapshotEntryId: string;
  }): Promise<SuiteCalcuttaEvaluationSnapshotEntryResponse> {
    return apiClient.get<SuiteCalcuttaEvaluationSnapshotEntryResponse>(
      `/cohorts/${encodeURIComponent(params.cohortId)}/simulations/${encodeURIComponent(params.id)}/entries/${encodeURIComponent(
        params.snapshotEntryId
      )}`
    );
  },

  async create(params: { cohortId: string; req: CreateSuiteCalcuttaEvaluationRequest }): Promise<CreateSimulationRunResponse> {
    return apiClient.post<CreateSimulationRunResponse>(
      `/cohorts/${encodeURIComponent(params.cohortId)}/simulations`,
      params.req
    );
  },
};

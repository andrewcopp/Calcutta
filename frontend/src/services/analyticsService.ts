import { API_URL, apiClient } from '../api/apiClient';
import { AnalyticsResponse, SeedInvestmentDistributionResponse } from '../types/analytics';

export const analyticsService = {
  async getAnalytics(): Promise<AnalyticsResponse> {
    return apiClient.get<AnalyticsResponse>('/analytics');
  },

  async listAlgorithms<T>(kind?: string): Promise<T> {
    const query = new URLSearchParams();
    if (kind) {
      query.set('kind', kind);
    }
    const suffix = query.toString() ? `?${query.toString()}` : '';
    return apiClient.get<T>(`/analytics/algorithms${suffix}`);
  },

  async getGameOutcomesAlgorithmCoverage<T>(): Promise<T> {
    return apiClient.get<T>('/analytics/coverage/game-outcomes');
  },

  async getMarketShareAlgorithmCoverage<T>(): Promise<T> {
    return apiClient.get<T>('/analytics/coverage/market-share');
  },

  async getGameOutcomesAlgorithmCoverageDetail<T>(algorithmId: string): Promise<T> {
    return apiClient.get<T>(`/analytics/algorithms/${encodeURIComponent(algorithmId)}/coverage/game-outcomes`);
  },

  async getMarketShareAlgorithmCoverageDetail<T>(algorithmId: string): Promise<T> {
    return apiClient.get<T>(`/analytics/algorithms/${encodeURIComponent(algorithmId)}/coverage/market-share`);
  },

  async listGameOutcomeRunsForTournament<T>(tournamentId: string): Promise<T> {
    return apiClient.get<T>(`/analytics/tournaments/${encodeURIComponent(tournamentId)}/game-outcome-runs`);
  },

  async listMarketShareRunsForCalcutta<T>(calcuttaId: string): Promise<T> {
    return apiClient.get<T>(`/analytics/calcuttas/${encodeURIComponent(calcuttaId)}/market-share-runs`);
  },

  async getLatestPredictionRunsForCalcutta<T>(calcuttaId: string): Promise<T> {
    return apiClient.get<T>(`/analytics/calcuttas/${encodeURIComponent(calcuttaId)}/latest-prediction-runs`);
  },

  async getTournamentPredictedAdvancement<T>(params: { tournamentId: string; gameOutcomeRunId?: string }): Promise<T> {
    const query = new URLSearchParams();
    if (params.gameOutcomeRunId) {
      query.set('game_outcome_run_id', params.gameOutcomeRunId);
    }
    const suffix = query.toString() ? `?${query.toString()}` : '';
    return apiClient.get<T>(`/analytics/tournaments/${encodeURIComponent(params.tournamentId)}/predicted-advancement${suffix}`);
  },

  async getCalcuttaPredictedMarketShare<T>(params: {
    calcuttaId: string;
    marketShareRunId?: string;
    gameOutcomeRunId?: string;
  }): Promise<T> {
    const query = new URLSearchParams();
    if (params.marketShareRunId) {
      query.set('market_share_run_id', params.marketShareRunId);
    }
    if (params.gameOutcomeRunId) {
      query.set('game_outcome_run_id', params.gameOutcomeRunId);
    }
    const suffix = query.toString() ? `?${query.toString()}` : '';
    return apiClient.get<T>(`/analytics/calcuttas/${encodeURIComponent(params.calcuttaId)}/predicted-market-share${suffix}`);
  },

  async getSeedInvestmentDistribution(): Promise<SeedInvestmentDistributionResponse> {
    return apiClient.get<SeedInvestmentDistributionResponse>('/analytics/seed-investment-distribution');
  },

  async getTournamentSimulationStats<T>(tournamentId: string): Promise<T> {
    return apiClient.get<T>(`/analytics/tournaments/${tournamentId}/simulations`);
  },

  async getCalcuttaPredictedReturns<T>(calcuttaId: string): Promise<T> {
    return apiClient.get<T>(`/analytics/calcuttas/${calcuttaId}/predicted-returns`);
  },

  async getCalcuttaPredictedInvestment<T>(calcuttaId: string): Promise<T> {
    return apiClient.get<T>(`/analytics/calcuttas/${calcuttaId}/predicted-investment`);
  },

  async getCalcuttaSimulatedEntry<T>(calcuttaId: string): Promise<T> {
    return apiClient.get<T>(`/analytics/calcuttas/${calcuttaId}/simulated-entry`);
  },

  async getCalcuttaSimulatedCalcuttas<T>(calcuttaId: string): Promise<T> {
    return apiClient.get<T>(`/analytics/calcuttas/${encodeURIComponent(calcuttaId)}/simulated-calcuttas`);
  },

  async listLabEntriesCoverage<T>(): Promise<T> {
    return apiClient.get<T>('/lab/entries');
  },

  async getLabEntriesSuiteDetail<T>(suiteId: string): Promise<T> {
    return apiClient.get<T>(`/lab/entries/cohorts/${encodeURIComponent(suiteId)}`);
  },

  async createLabSuiteSandboxExecution<T>(suiteId: string): Promise<T> {
    return apiClient.post<T>(`/lab/entries/cohorts/${encodeURIComponent(suiteId)}/sandbox-executions`, {});
  },

  async getLabEntryReport<T>(scenarioId: string): Promise<T> {
    return apiClient.get<T>(`/lab/entries/scenarios/${encodeURIComponent(scenarioId)}`);
  },

  async exportAnalyticsSnapshot(tournamentId: string, calcuttaId: string): Promise<{ blob: Blob; filename: string }> {
    const url = new URL(`${API_URL}/api/admin/analytics/export`);
    url.searchParams.set('tournamentId', tournamentId);
    url.searchParams.set('calcuttaId', calcuttaId);

    const res = await apiClient.fetch(url.toString(), { credentials: 'include' });
    if (!res.ok) {
      const txt = await res.text().catch(() => '');
      throw new Error(txt || `Export failed (${res.status})`);
    }

    const blob = await res.blob();
    const cd = res.headers.get('content-disposition') || '';
    const match = /filename="([^"]+)"/i.exec(cd);
    const filename = match?.[1] || 'analytics-snapshot.zip';

    return { blob, filename };
  },
};

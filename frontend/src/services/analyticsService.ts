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

	async bulkCreateGameOutcomeRunsForAlgorithm<T>(algorithmId: string): Promise<T> {
		return apiClient.post<T>(`/analytics/algorithms/${encodeURIComponent(algorithmId)}/game-outcome-runs/bulk`, {});
	},

	async bulkCreateMarketShareRunsForAlgorithm<T>(algorithmId: string, params: Record<string, unknown>): Promise<T> {
		return apiClient.post<T>(`/analytics/algorithms/${encodeURIComponent(algorithmId)}/market-share-runs/bulk`, params || {});
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

  async getCalcuttaPredictedReturns<T>(params: { calcuttaId: string; entryRunId?: string; gameOutcomeRunId: string }): Promise<T> {
    const query = new URLSearchParams();
    if (params.entryRunId) {
      query.set('entry_run_id', params.entryRunId);
    }
    query.set('game_outcome_run_id', params.gameOutcomeRunId);
    const suffix = query.toString() ? `?${query.toString()}` : '';
    return apiClient.get<T>(`/analytics/calcuttas/${encodeURIComponent(params.calcuttaId)}/predicted-returns${suffix}`);
  },

  async getCalcuttaPredictedInvestment<T>(params: {
    calcuttaId: string;
    entryRunId?: string;
    marketShareRunId: string;
    gameOutcomeRunId: string;
  }): Promise<T> {
    const query = new URLSearchParams();
    if (params.entryRunId) {
      query.set('entry_run_id', params.entryRunId);
    }
    query.set('market_share_run_id', params.marketShareRunId);
    query.set('game_outcome_run_id', params.gameOutcomeRunId);
    const suffix = query.toString() ? `?${query.toString()}` : '';
    return apiClient.get<T>(`/analytics/calcuttas/${encodeURIComponent(params.calcuttaId)}/predicted-investment${suffix}`);
  },

  async getCalcuttaSimulatedEntry<T>(params: { calcuttaId: string; entryRunId: string }): Promise<T> {
    const query = new URLSearchParams();
    query.set('entry_run_id', params.entryRunId);
    const suffix = query.toString() ? `?${query.toString()}` : '';
    return apiClient.get<T>(`/analytics/calcuttas/${encodeURIComponent(params.calcuttaId)}/simulated-entry${suffix}`);
  },

  async getCalcuttaSimulatedCalcuttas<T>(calcuttaId: string): Promise<T> {
    return apiClient.get<T>(`/analytics/calcuttas/${encodeURIComponent(calcuttaId)}/simulated-calcuttas`);
  },

	async listLabCandidates<T>(params?: {
		calcuttaId?: string;
		tournamentId?: string;
		strategyGenerationRunId?: string;
		marketShareArtifactId?: string;
		advancementRunId?: string;
		optimizerKey?: string;
		startingStateKey?: string;
		excludedEntryName?: string;
		sourceKind?: string;
		limit?: number;
		offset?: number;
	}): Promise<T> {
		const q = new URLSearchParams();
		if (params?.calcuttaId) q.set('calcutta_id', params.calcuttaId);
		if (params?.tournamentId) q.set('tournament_id', params.tournamentId);
		if (params?.strategyGenerationRunId) q.set('strategy_generation_run_id', params.strategyGenerationRunId);
		if (params?.marketShareArtifactId) q.set('market_share_artifact_id', params.marketShareArtifactId);
		if (params?.advancementRunId) q.set('advancement_run_id', params.advancementRunId);
		if (params?.optimizerKey) q.set('optimizer_key', params.optimizerKey);
		if (params?.startingStateKey) q.set('starting_state_key', params.startingStateKey);
		if (params?.excludedEntryName) q.set('excluded_entry_name', params.excludedEntryName);
		if (params?.sourceKind) q.set('source_kind', params.sourceKind);
		if (params?.limit != null) q.set('limit', String(params.limit));
		if (params?.offset != null) q.set('offset', String(params.offset));
		const suffix = q.toString() ? `?${q.toString()}` : '';
		return apiClient.get<T>(`/lab/candidates${suffix}`);
	},

	async getLabCandidate<T>(candidateId: string): Promise<T> {
		return apiClient.get<T>(`/lab/candidates/${encodeURIComponent(candidateId)}`);
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

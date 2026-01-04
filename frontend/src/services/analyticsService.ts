import { apiClient } from '../api/apiClient';
import { AnalyticsResponse, SeedInvestmentDistributionResponse } from '../types/analytics';

const API_URL = import.meta.env.VITE_API_URL || import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';

export const analyticsService = {
  async getAnalytics(): Promise<AnalyticsResponse> {
    return apiClient.get<AnalyticsResponse>('/analytics');
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
    return apiClient.get<T>(`/analytics/calcuttas/${calcuttaId}/simulated-calcuttas`);
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

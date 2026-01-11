import { apiClient } from '../api/apiClient';

export type SimulatedCalcuttaListItem = {
  id: string;
  name: string;
  description?: string | null;
  tournament_id: string;
  base_calcutta_id?: string | null;
  starting_state_key: string;
  excluded_entry_name?: string | null;
  highlighted_simulated_entry_id?: string | null;
  metadata: unknown;
  created_at: string;
  updated_at: string;
};

export type ListSimulatedCalcuttasResponse = {
  items: SimulatedCalcuttaListItem[];
};

export type CreateSimulatedCalcuttaFromCalcuttaRequest = {
  calcuttaId: string;
  name?: string;
  description?: string;
  startingStateKey?: string;
  excludedEntryName?: string;
  metadata?: Record<string, unknown>;
};

export type CreateSimulatedCalcuttaFromCalcuttaResponse = {
  id: string;
  copiedEntries: number;
};

export type PatchSimulatedCalcuttaRequest = {
  name?: string;
  description?: string | null;
  highlightedSimulatedEntryId?: string | null;
  metadata?: Record<string, unknown>;
};

export type SimulatedCalcuttaPayout = {
  position: number;
  amountCents: number;
};

export type SimulatedCalcuttaScoringRule = {
  winIndex: number;
  pointsAwarded: number;
};

export type GetSimulatedCalcuttaResponse = {
  simulated_calcutta: SimulatedCalcuttaListItem;
  payouts: SimulatedCalcuttaPayout[];
  scoring_rules: SimulatedCalcuttaScoringRule[];
};

export const simulatedCalcuttasService = {
  async list(params?: {
    tournamentId?: string;
    baseCalcuttaId?: string;
    cohortId?: string;
    limit?: number;
    offset?: number;
  }): Promise<ListSimulatedCalcuttasResponse> {
    const q = new URLSearchParams();
    if (params?.tournamentId) q.set('tournament_id', params.tournamentId);
    if (params?.baseCalcuttaId) q.set('base_calcutta_id', params.baseCalcuttaId);
    if (params?.cohortId) q.set('cohort_id', params.cohortId);
    if (params?.limit != null) q.set('limit', String(params.limit));
    if (params?.offset != null) q.set('offset', String(params.offset));

    const suffix = q.toString() ? `?${q.toString()}` : '';
    return apiClient.get<ListSimulatedCalcuttasResponse>(`/simulated-calcuttas${suffix}`);
  },

  async get(id: string): Promise<GetSimulatedCalcuttaResponse> {
    return apiClient.get<GetSimulatedCalcuttaResponse>(`/simulated-calcuttas/${encodeURIComponent(id)}`);
  },

  async createFromCalcutta(
    req: CreateSimulatedCalcuttaFromCalcuttaRequest
  ): Promise<CreateSimulatedCalcuttaFromCalcuttaResponse> {
    return apiClient.post<CreateSimulatedCalcuttaFromCalcuttaResponse>('/simulated-calcuttas/from-calcutta', req);
  },

  async patch(id: string, req: PatchSimulatedCalcuttaRequest): Promise<{ ok: boolean }> {
    return apiClient.patch<{ ok: boolean }>(`/simulated-calcuttas/${encodeURIComponent(id)}`, req);
  },
};

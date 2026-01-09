import { apiClient } from '../api/apiClient';

export type SimulatedEntryTeam = {
  team_id: string;
  bid_points: number;
};

export type SimulatedEntryListItem = {
  id: string;
  simulated_calcutta_id: string;
  display_name: string;
  source_kind: string;
  source_entry_id?: string | null;
  source_candidate_id?: string | null;
  teams: SimulatedEntryTeam[];
  created_at: string;
  updated_at: string;
};

export type ListSimulatedEntriesResponse = {
  items: SimulatedEntryListItem[];
};

export const simulatedEntriesService = {
  async list(simulatedCalcuttaId: string): Promise<ListSimulatedEntriesResponse> {
    return apiClient.get<ListSimulatedEntriesResponse>(
      `/simulated-calcuttas/${encodeURIComponent(simulatedCalcuttaId)}/entries`
    );
  },
};

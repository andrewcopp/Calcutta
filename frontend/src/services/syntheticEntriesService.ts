import { apiClient } from '../api/apiClient';

export type SyntheticEntryTeam = {
  team_id: string;
  bid_points: number;
};

export type SyntheticEntryListItem = {
  id: string;
  entry_id?: string | null;
  display_name: string;
  is_synthetic: boolean;
  teams: SyntheticEntryTeam[];
  created_at: string;
  updated_at: string;
};

export type ListSyntheticEntriesResponse = {
  items: SyntheticEntryListItem[];
};

export type CreateSyntheticEntryRequest = {
  displayName: string;
  teams?: { team_id: string; bid_points: number }[];
};

export type CreateSyntheticEntryResponse = {
  id: string;
};

export type PatchSyntheticEntryRequest = {
  displayName?: string;
  teams?: { team_id: string; bid_points: number }[];
};

export const syntheticEntriesService = {
  async list(syntheticCalcuttaId: string): Promise<ListSyntheticEntriesResponse> {
    return apiClient.get<ListSyntheticEntriesResponse>(
      `/synthetic-calcuttas/${encodeURIComponent(syntheticCalcuttaId)}/synthetic-entries`
    );
  },

  async create(syntheticCalcuttaId: string, req: CreateSyntheticEntryRequest): Promise<CreateSyntheticEntryResponse> {
    return apiClient.post<CreateSyntheticEntryResponse>(
      `/synthetic-calcuttas/${encodeURIComponent(syntheticCalcuttaId)}/synthetic-entries`,
      req
    );
  },

  async patch(syntheticEntryId: string, req: PatchSyntheticEntryRequest): Promise<{ ok: boolean }> {
    return apiClient.patch<{ ok: boolean }>(`/synthetic-entries/${encodeURIComponent(syntheticEntryId)}`, req);
  },

  async delete(syntheticEntryId: string): Promise<{ ok: boolean }> {
    return apiClient.delete<{ ok: boolean }>(`/synthetic-entries/${encodeURIComponent(syntheticEntryId)}`);
  },
};

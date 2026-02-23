import { apiClient } from '../api/apiClient';
import type { TournamentExportResult, CreateAPIKeyRequest } from '../schemas/admin';
import {
  AdminUsersListResponseSchema,
  AdminUserDetailResponseSchema,
  UserResponseSchema,
  AdminInviteUserResponseSchema,
  ListAPIKeysResponseSchema,
  CreateAPIKeyResponseSchema,
  TournamentImportStartResponseSchema,
  TournamentImportStatusResponseSchema,
} from '../schemas/admin';

export const adminService = {
  async listUsers(status?: string) {
    const params = status ? `?status=${encodeURIComponent(status)}` : '';
    return apiClient.get(`/admin/users${params}`, { schema: AdminUsersListResponseSchema });
  },

  async setUserEmail(userId: string, email: string) {
    return apiClient.patch(`/admin/users/${userId}/email`, { email }, { schema: UserResponseSchema });
  },

  async sendInvite(userId: string) {
    return apiClient.post(`/admin/users/${userId}/invite/send`, undefined, { schema: AdminInviteUserResponseSchema });
  },

  async getUser(userId: string) {
    return apiClient.get(`/admin/users/${userId}`, { schema: AdminUserDetailResponseSchema });
  },

  async grantRole(userId: string, roleKey: string, scopeType?: string, scopeId?: string): Promise<void> {
    return apiClient.post<void>(`/admin/users/${userId}/roles`, { roleKey, scopeType, scopeId });
  },

  async revokeRole(userId: string, roleKey: string, scopeType?: string, scopeId?: string): Promise<void> {
    const params = new URLSearchParams();
    if (scopeType) params.set('scopeType', scopeType);
    if (scopeId) params.set('scopeId', scopeId);
    const qs = params.toString() ? `?${params.toString()}` : '';
    return apiClient.delete<void>(`/admin/users/${userId}/roles/${roleKey}${qs}`);
  },

  async listApiKeys() {
    return apiClient.get('/admin/api-keys', { schema: ListAPIKeysResponseSchema });
  },

  async createApiKey(label: string) {
    const body: CreateAPIKeyRequest = {};
    if (label) body.label = label;
    return apiClient.post('/admin/api-keys', body, { schema: CreateAPIKeyResponseSchema });
  },

  async revokeApiKey(id: string): Promise<void> {
    return apiClient.delete<void>(`/admin/api-keys/${id}`);
  },

  // --- Tournament Import/Export ---

  async exportTournamentData(): Promise<TournamentExportResult> {
    const res = await apiClient.fetch('/admin/tournament-imports/export');
    if (!res.ok) {
      const txt = await res.text().catch(() => '');
      throw new Error(txt || `Export failed (${res.status})`);
    }
    const blob = await res.blob();
    const cd = res.headers.get('content-disposition') || '';
    const match = /filename="([^"]+)"/i.exec(cd);
    const filename = match?.[1] || 'tournament-data.zip';
    return { blob, filename };
  },

  async startTournamentImport(file: File) {
    const form = new FormData();
    form.append('file', file);
    return apiClient.post('/admin/tournament-imports/import', form, { schema: TournamentImportStartResponseSchema });
  },

  async getTournamentImportStatus(uploadId: string) {
    return apiClient.get(`/admin/tournament-imports/import/${uploadId}`, {
      schema: TournamentImportStatusResponseSchema,
    });
  },
};

import { apiClient } from '../api/apiClient';
import type {
  AdminUsersListResponse,
  AdminUserDetailResponse,
  UserResponse,
  AdminInviteUserResponse,
  ListAPIKeysResponse,
  CreateAPIKeyRequest,
  CreateAPIKeyResponse,
  BundleExportResult,
  BundleImportStartResponse,
  BundleImportStatusResponse,
} from '../types/admin';

export const adminService = {
  async listUsers(status?: string): Promise<AdminUsersListResponse> {
    const params = status ? `?status=${encodeURIComponent(status)}` : '';
    return apiClient.get<AdminUsersListResponse>(`/admin/users${params}`);
  },

  async setUserEmail(userId: string, email: string): Promise<UserResponse> {
    return apiClient.patch<UserResponse>(`/admin/users/${userId}/email`, { email });
  },

  async sendInvite(userId: string): Promise<AdminInviteUserResponse> {
    return apiClient.post<AdminInviteUserResponse>(`/admin/users/${userId}/invite/send`);
  },

  async getUser(userId: string): Promise<AdminUserDetailResponse> {
    return apiClient.get<AdminUserDetailResponse>(`/admin/users/${userId}`);
  },

  async grantLabel(userId: string, labelKey: string, scopeType?: string, scopeId?: string): Promise<void> {
    return apiClient.post<void>(`/admin/users/${userId}/labels`, { labelKey, scopeType, scopeId });
  },

  async revokeLabel(userId: string, labelKey: string, scopeType?: string, scopeId?: string): Promise<void> {
    const params = new URLSearchParams();
    if (scopeType) params.set('scopeType', scopeType);
    if (scopeId) params.set('scopeId', scopeId);
    const qs = params.toString() ? `?${params.toString()}` : '';
    return apiClient.delete<void>(`/admin/users/${userId}/labels/${labelKey}${qs}`);
  },

  async listApiKeys(): Promise<ListAPIKeysResponse> {
    return apiClient.get<ListAPIKeysResponse>('/admin/api-keys');
  },

  async createApiKey(label: string): Promise<CreateAPIKeyResponse> {
    const body: CreateAPIKeyRequest = {};
    if (label) body.label = label;
    return apiClient.post<CreateAPIKeyResponse>('/admin/api-keys', body);
  },

  async revokeApiKey(id: string): Promise<void> {
    return apiClient.delete<void>(`/admin/api-keys/${id}`);
  },

  // --- Bundle Import/Export ---

  async exportBundle(): Promise<BundleExportResult> {
    const res = await apiClient.fetch('/admin/bundles/export');
    if (!res.ok) {
      const txt = await res.text().catch(() => '');
      throw new Error(txt || `Export failed (${res.status})`);
    }
    const blob = await res.blob();
    const cd = res.headers.get('content-disposition') || '';
    const match = /filename="([^"]+)"/i.exec(cd);
    const filename = match?.[1] || 'bundles.zip';
    return { blob, filename };
  },

  async startBundleImport(file: File): Promise<BundleImportStartResponse> {
    const form = new FormData();
    form.append('file', file);
    return apiClient.post<BundleImportStartResponse>('/admin/bundles/import', form);
  },

  async getBundleImportStatus(uploadId: string): Promise<BundleImportStatusResponse> {
    return apiClient.get<BundleImportStatusResponse>(`/admin/bundles/import/${uploadId}`);
  },
};

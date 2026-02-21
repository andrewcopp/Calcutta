import { apiClient } from '../api/apiClient';
import type {
  AdminUsersListResponse,
  UserResponse,
  AdminInviteUserResponse,
  ListAPIKeysResponse,
  CreateAPIKeyRequest,
  CreateAPIKeyResponse,
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
};

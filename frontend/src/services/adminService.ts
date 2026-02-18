import { apiClient } from '../api/apiClient';

export type AdminUserListItem = {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  created_at: string;
  updated_at: string;
  labels: string[];
  permissions: string[];
};

export type AdminUsersListResponse = {
  items: AdminUserListItem[];
};

export type CreateAPIKeyRequest = {
  label?: string;
};

export type CreateAPIKeyResponse = {
  id: string;
  key: string;
  label?: string;
  created_at: string;
};

export type APIKeyListItem = {
  id: string;
  label?: string;
  created_at: string;
  revoked_at?: string;
  last_used_at?: string;
};

export type ListAPIKeysResponse = {
  items: APIKeyListItem[];
};

export const adminService = {
  async listUsers(): Promise<AdminUsersListResponse> {
    return apiClient.get<AdminUsersListResponse>('/admin/users');
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

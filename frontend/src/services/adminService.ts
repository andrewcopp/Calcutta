import { apiClient } from '../api/apiClient';

export type AdminUserListItem = {
  id: string;
  email: string | null;
  first_name: string;
  last_name: string;
  status: string;
  invited_at: string | null;
  last_invite_sent_at: string | null;
  invite_expires_at: string | null;
  invite_consumed_at: string | null;
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

export type AdminInviteUserResponse = {
  userId: string;
  email: string;
  inviteToken: string;
  inviteExpiresAt: string;
  status: string;
};

export type UserResponse = {
  id: string;
  email: string | null;
  first_name: string;
  last_name: string;
  status: string;
};

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

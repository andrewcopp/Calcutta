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

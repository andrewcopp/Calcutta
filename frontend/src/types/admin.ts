export type AdminUserListItem = {
  id: string;
  email: string | null;
  firstName: string;
  lastName: string;
  status: string;
  invitedAt: string | null;
  lastInviteSentAt: string | null;
  inviteExpiresAt: string | null;
  inviteConsumedAt: string | null;
  createdAt: string;
  updatedAt: string;
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
  createdAt: string;
};

export type APIKeyListItem = {
  id: string;
  label?: string;
  createdAt: string;
  revokedAt?: string;
  lastUsedAt?: string;
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
  firstName: string;
  lastName: string;
  status: string;
};

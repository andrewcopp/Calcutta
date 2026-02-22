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

// --- Bundle Import/Export ---

export type BundleImportStatus = 'pending' | 'running' | 'succeeded' | 'failed';

export type BundleImportReport = {
  startedAt: string;
  finishedAt: string;
  dryRun: boolean;
  schools: number;
  tournaments: number;
  tournamentTeams: number;
  calcuttas: number;
  entries: number;
  bids: number;
  payouts: number;
  rounds: number;
};

export type BundleVerifyReport = {
  ok: boolean;
  mismatchCount: number;
  mismatches?: { where: string; what: string }[];
};

export type BundleImportStartResponse = {
  uploadId: string;
  status: BundleImportStatus;
  filename: string;
  sha256: string;
  sizeBytes: number;
};

export type BundleImportStatusResponse = {
  uploadId: string;
  filename: string;
  sha256: string;
  sizeBytes: number;
  status: BundleImportStatus;
  startedAt?: string;
  finishedAt?: string;
  errorMessage?: string;
  importReport?: BundleImportReport;
  verifyReport?: BundleVerifyReport;
};

export type BundleExportResult = {
  blob: Blob;
  filename: string;
};

export type UserProfileResponse = {
  id: string;
  email: string | null;
  firstName: string;
  lastName: string;
  status: string;
  labels: string[];
  permissions: string[];
  createdAt: string;
  updatedAt: string;
};

export type LabelGrant = {
  key: string;
  scopeType: 'global' | 'calcutta' | 'tournament';
  scopeId?: string;
  scopeName?: string;
};

export type AdminUserDetailResponse = {
  id: string;
  email: string | null;
  firstName: string;
  lastName: string;
  status: string;
  labels: LabelGrant[];
  permissions: string[];
  createdAt: string;
  updatedAt: string;
};

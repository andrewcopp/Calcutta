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
  roles: string[];
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

// --- Tournament Import/Export ---

export type TournamentImportStatus = 'pending' | 'running' | 'succeeded' | 'failed';

export type TournamentImportReport = {
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

export type TournamentVerifyReport = {
  ok: boolean;
  mismatchCount: number;
  mismatches?: { where: string; what: string }[];
};

export type TournamentImportStartResponse = {
  uploadId: string;
  status: TournamentImportStatus;
  filename: string;
  sha256: string;
  sizeBytes: number;
};

export type TournamentImportStatusResponse = {
  uploadId: string;
  filename: string;
  sha256: string;
  sizeBytes: number;
  status: TournamentImportStatus;
  startedAt?: string;
  finishedAt?: string;
  errorMessage?: string;
  importReport?: TournamentImportReport;
  verifyReport?: TournamentVerifyReport;
};

export type TournamentExportResult = {
  blob: Blob;
  filename: string;
};

export type UserProfileResponse = {
  id: string;
  email: string | null;
  firstName: string;
  lastName: string;
  status: string;
  roles: string[];
  permissions: string[];
  createdAt: string;
  updatedAt: string;
};

export type RoleGrant = {
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
  roles: RoleGrant[];
  permissions: string[];
  createdAt: string;
  updatedAt: string;
};

import { z } from 'zod';

export const AdminUserListItemSchema = z.object({
  id: z.string(),
  email: z.string().nullable(),
  firstName: z.string(),
  lastName: z.string(),
  status: z.string(),
  invitedAt: z.string().nullable(),
  lastInviteSentAt: z.string().nullable(),
  inviteExpiresAt: z.string().nullable(),
  inviteConsumedAt: z.string().nullable(),
  createdAt: z.string(),
  updatedAt: z.string(),
  roles: z.array(z.string()),
  permissions: z.array(z.string()),
});

export type AdminUserListItem = z.infer<typeof AdminUserListItemSchema>;

export const AdminUsersListResponseSchema = z.object({
  items: z.array(AdminUserListItemSchema),
});

export type AdminUsersListResponse = z.infer<typeof AdminUsersListResponseSchema>;

export const CreateAPIKeyRequestSchema = z.object({
  label: z.string().optional(),
});

export type CreateAPIKeyRequest = z.infer<typeof CreateAPIKeyRequestSchema>;

export const CreateAPIKeyResponseSchema = z.object({
  id: z.string(),
  key: z.string(),
  label: z.string().optional(),
  createdAt: z.string(),
});

export type CreateAPIKeyResponse = z.infer<typeof CreateAPIKeyResponseSchema>;

export const APIKeyListItemSchema = z.object({
  id: z.string(),
  label: z.string().optional(),
  createdAt: z.string(),
  revokedAt: z.string().optional(),
  lastUsedAt: z.string().optional(),
});

export type APIKeyListItem = z.infer<typeof APIKeyListItemSchema>;

export const ListAPIKeysResponseSchema = z.object({
  items: z.array(APIKeyListItemSchema),
});

export type ListAPIKeysResponse = z.infer<typeof ListAPIKeysResponseSchema>;

export const AdminInviteUserResponseSchema = z.object({
  userId: z.string(),
  email: z.string(),
  inviteToken: z.string(),
  inviteExpiresAt: z.string(),
  status: z.string(),
});

export type AdminInviteUserResponse = z.infer<typeof AdminInviteUserResponseSchema>;

export const UserResponseSchema = z.object({
  id: z.string(),
  email: z.string().nullable(),
  firstName: z.string(),
  lastName: z.string(),
  status: z.string(),
});

export type UserResponse = z.infer<typeof UserResponseSchema>;

export const TournamentImportStatusValueSchema = z.enum(['pending', 'running', 'succeeded', 'failed']);

export type TournamentImportStatus = z.infer<typeof TournamentImportStatusValueSchema>;

export const TournamentImportReportSchema = z.object({
  startedAt: z.string(),
  finishedAt: z.string(),
  dryRun: z.boolean(),
  schools: z.number(),
  tournaments: z.number(),
  tournamentTeams: z.number(),
  calcuttas: z.number(),
  entries: z.number(),
  bids: z.number(),
  payouts: z.number(),
  rounds: z.number(),
});

export type TournamentImportReport = z.infer<typeof TournamentImportReportSchema>;

export const TournamentVerifyReportSchema = z.object({
  ok: z.boolean(),
  mismatchCount: z.number(),
  mismatches: z
    .array(
      z.object({
        where: z.string(),
        what: z.string(),
      }),
    )
    .optional(),
});

export type TournamentVerifyReport = z.infer<typeof TournamentVerifyReportSchema>;

export const TournamentImportStartResponseSchema = z.object({
  uploadId: z.string(),
  status: TournamentImportStatusValueSchema,
  filename: z.string(),
  sha256: z.string(),
  sizeBytes: z.number(),
});

export type TournamentImportStartResponse = z.infer<typeof TournamentImportStartResponseSchema>;

export const TournamentImportStatusResponseSchema = z.object({
  uploadId: z.string(),
  filename: z.string(),
  sha256: z.string(),
  sizeBytes: z.number(),
  status: TournamentImportStatusValueSchema,
  startedAt: z.string().optional(),
  finishedAt: z.string().optional(),
  errorMessage: z.string().optional(),
  importReport: TournamentImportReportSchema.optional(),
  verifyReport: TournamentVerifyReportSchema.optional(),
});

export type TournamentImportStatusResponse = z.infer<typeof TournamentImportStatusResponseSchema>;

// TournamentExportResult is not a server response — it's locally constructed from a blob.
// Keep as plain type (no schema needed).
export type TournamentExportResult = {
  blob: Blob;
  filename: string;
};

// --- User Merge Schemas ---

export const StubUserSchema = z.object({
  id: z.string(),
  firstName: z.string(),
  lastName: z.string(),
  email: z.string().nullable().optional(),
  status: z.string(),
  createdAt: z.string(),
});

export type StubUser = z.infer<typeof StubUserSchema>;

export const StubUsersListResponseSchema = z.object({
  items: z.array(StubUserSchema),
});

export type StubUsersListResponse = z.infer<typeof StubUsersListResponseSchema>;

export const MergeCandidateSchema = z.object({
  id: z.string(),
  firstName: z.string(),
  lastName: z.string(),
  email: z.string().nullable().optional(),
  status: z.string(),
  createdAt: z.string(),
});

export type MergeCandidate = z.infer<typeof MergeCandidateSchema>;

export const MergeCandidatesListResponseSchema = z.object({
  items: z.array(MergeCandidateSchema),
});

export type MergeCandidatesListResponse = z.infer<typeof MergeCandidatesListResponseSchema>;

export const UserMergeResponseSchema = z.object({
  id: z.string(),
  sourceUserId: z.string(),
  targetUserId: z.string(),
  mergedBy: z.string(),
  entriesMoved: z.number(),
  invitationsMoved: z.number(),
  grantsMoved: z.number(),
  createdAt: z.string(),
});

export type UserMergeResponse = z.infer<typeof UserMergeResponseSchema>;

export const MergeHistoryResponseSchema = z.object({
  items: z.array(UserMergeResponseSchema),
});

export type MergeHistoryResponse = z.infer<typeof MergeHistoryResponseSchema>;

export const UserProfileResponseSchema = z.object({
  id: z.string(),
  email: z.string().nullable(),
  firstName: z.string(),
  lastName: z.string(),
  status: z.string(),
  roles: z.array(z.string()),
  permissions: z.array(z.string()),
  createdAt: z.string(),
  updatedAt: z.string(),
});

export type UserProfileResponse = z.infer<typeof UserProfileResponseSchema>;

export const RoleGrantSchema = z.object({
  key: z.string(),
  scopeType: z.enum(['global', 'calcutta', 'tournament']),
  scopeId: z.string().optional(),
  scopeName: z.string().optional(),
});

export type RoleGrant = z.infer<typeof RoleGrantSchema>;

export const AdminUserDetailResponseSchema = z.object({
  id: z.string(),
  email: z.string().nullable(),
  firstName: z.string(),
  lastName: z.string(),
  status: z.string(),
  roles: z.array(RoleGrantSchema),
  permissions: z.array(z.string()),
  createdAt: z.string(),
  updatedAt: z.string(),
});

export type AdminUserDetailResponse = z.infer<typeof AdminUserDetailResponseSchema>;

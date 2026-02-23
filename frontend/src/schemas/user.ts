import { z } from 'zod';

export const UserSchema = z.object({
  id: z.string(),
  email: z.string().optional(),
  firstName: z.string(),
  lastName: z.string(),
  status: z.string(),
  createdAt: z.string(),
  updatedAt: z.string(),
});

export type User = z.infer<typeof UserSchema>;

export const LoginRequestSchema = z.object({
  email: z.string(),
  password: z.string(),
});

export type LoginRequest = z.infer<typeof LoginRequestSchema>;

export const SignupRequestSchema = z.object({
  email: z.string(),
  firstName: z.string(),
  lastName: z.string(),
  password: z.string(),
});

export type SignupRequest = z.infer<typeof SignupRequestSchema>;

export const AuthResponseSchema = z.object({
  user: UserSchema,
  accessToken: z.string(),
});

export type AuthResponse = z.infer<typeof AuthResponseSchema>;

export const InvitePreviewSchema = z.object({
  firstName: z.string(),
  calcuttaName: z.string(),
  commissionerName: z.string(),
  tournamentStartingAt: z.string().optional(),
});

export type InvitePreview = z.infer<typeof InvitePreviewSchema>;

export const PermissionsResponseSchema = z.object({
  permissions: z.array(z.string()),
});

export type PermissionsResponse = z.infer<typeof PermissionsResponseSchema>;

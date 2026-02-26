import type { User, LoginRequest } from '../schemas/user';
import { AuthResponseSchema, InvitePreviewSchema, PermissionsResponseSchema } from '../schemas/user';
import { UserProfileResponseSchema } from '../schemas/admin';
import { apiClient, USER_KEY, PERMISSIONS_KEY } from '../api/apiClient';

export const userService = {
  async login(request: LoginRequest): Promise<User> {
    const res = await apiClient.post('/auth/login', request, {
      credentials: 'include',
      schema: AuthResponseSchema,
    });
    localStorage.setItem(USER_KEY, JSON.stringify(res.user));
    apiClient.setAccessToken(res.accessToken);
    return res.user;
  },

  async previewInvite(token: string) {
    return apiClient.get(`/auth/invite/preview?token=${encodeURIComponent(token)}`, {
      schema: InvitePreviewSchema,
    });
  },

  async acceptInvite(token: string, password: string): Promise<User> {
    const res = await apiClient.post(
      '/auth/invite/accept',
      { token, password },
      { credentials: 'include', schema: AuthResponseSchema },
    );
    localStorage.setItem(USER_KEY, JSON.stringify(res.user));
    apiClient.setAccessToken(res.accessToken);
    return res.user;
  },

  async forgotPassword(email: string): Promise<void> {
    await apiClient.post<void>('/auth/forgot-password', { email });
  },

  async resetPassword(token: string, password: string): Promise<User> {
    const res = await apiClient.post(
      '/auth/reset-password',
      { token, password },
      { credentials: 'include', schema: AuthResponseSchema },
    );
    localStorage.setItem(USER_KEY, JSON.stringify(res.user));
    apiClient.setAccessToken(res.accessToken);
    return res.user;
  },

  logout(): void {
    void apiClient.post<void>('/auth/logout', undefined, { credentials: 'include' }).catch((err) => {
      console.error('Logout failed', err);
    });
    localStorage.removeItem(USER_KEY);
    localStorage.removeItem(PERMISSIONS_KEY);
    apiClient.setAccessToken(null);
  },

  getCurrentUser(): User | null {
    const userStr = localStorage.getItem(USER_KEY);
    if (!userStr) return null;
    try {
      return JSON.parse(userStr) as User;
    } catch (e) {
      console.error('Failed to parse stored user data', e);
      localStorage.removeItem(USER_KEY);
      return null;
    }
  },

  async fetchPermissions(): Promise<string[]> {
    try {
      const res = await apiClient.get('/me/permissions', { schema: PermissionsResponseSchema });
      const permissions = res.permissions ?? [];
      localStorage.setItem(PERMISSIONS_KEY, JSON.stringify(permissions));
      return permissions;
    } catch (e) {
      console.error('Failed to fetch permissions', e);
      return [];
    }
  },

  async fetchProfile() {
    return apiClient.get('/me/profile', { schema: UserProfileResponseSchema });
  },

  getStoredPermissions(): string[] {
    const permStr = localStorage.getItem(PERMISSIONS_KEY);
    if (!permStr) return [];
    try {
      return JSON.parse(permStr) as string[];
    } catch (e) {
      console.error('Failed to parse stored permissions', e);
      localStorage.removeItem(PERMISSIONS_KEY);
      return [];
    }
  },
};

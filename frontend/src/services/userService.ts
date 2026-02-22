import { User, LoginRequest, SignupRequest, AuthResponse, InvitePreview } from '../types/user';
import { apiClient, USER_KEY, PERMISSIONS_KEY } from '../api/apiClient';
import type { UserProfileResponse } from '../types/admin';

interface PermissionsResponse {
  permissions: string[];
}

export const userService = {
  async login(request: LoginRequest): Promise<User> {
    const res = await apiClient.post<AuthResponse>('/auth/login', request, { credentials: 'include' });
    localStorage.setItem(USER_KEY, JSON.stringify(res.user));
    apiClient.setAccessToken(res.accessToken);
    return res.user;
  },

  async signup(request: SignupRequest): Promise<User> {
    const res = await apiClient.post<AuthResponse>('/auth/signup', request, { credentials: 'include' });
    localStorage.setItem(USER_KEY, JSON.stringify(res.user));
    apiClient.setAccessToken(res.accessToken);
    return res.user;
  },

  async previewInvite(token: string): Promise<InvitePreview> {
    return apiClient.get<InvitePreview>(`/auth/invite/preview?token=${encodeURIComponent(token)}`);
  },

  async acceptInvite(token: string, password: string): Promise<User> {
    const res = await apiClient.post<AuthResponse>('/auth/invite/accept', { token, password }, { credentials: 'include' });
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
      const res = await apiClient.get<PermissionsResponse>('/me/permissions');
      const permissions = res.permissions ?? [];
      localStorage.setItem(PERMISSIONS_KEY, JSON.stringify(permissions));
      return permissions;
    } catch (e) {
      console.error('Failed to fetch permissions', e);
      return [];
    }
  },

  async fetchProfile(): Promise<UserProfileResponse> {
    return apiClient.get<UserProfileResponse>('/me/profile');
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
import { User, LoginRequest, SignupRequest, AuthResponse } from '../types/user';
import { apiClient } from '../api/apiClient';

interface PermissionsResponse {
  permissions: string[];
}

export const userService = {
  async login(request: LoginRequest): Promise<User> {
    const res = await apiClient.post<AuthResponse>('/auth/login', request, { credentials: 'include' });
    localStorage.setItem('user', JSON.stringify(res.user));
    apiClient.setAccessToken(res.accessToken);
    return res.user;
  },

  async signup(request: SignupRequest): Promise<User> {
    const res = await apiClient.post<AuthResponse>('/auth/signup', request, { credentials: 'include' });
    localStorage.setItem('user', JSON.stringify(res.user));
    apiClient.setAccessToken(res.accessToken);
    return res.user;
  },

  async acceptInvite(token: string, password: string): Promise<User> {
    const res = await apiClient.post<AuthResponse>('/auth/invite/accept', { token, password }, { credentials: 'include' });
    localStorage.setItem('user', JSON.stringify(res.user));
    apiClient.setAccessToken(res.accessToken);
    return res.user;
  },

  logout(): void {
    void apiClient.post<void>('/auth/logout', undefined, { credentials: 'include' }).catch((err) => {
      console.error('Logout failed', err);
    });
    localStorage.removeItem('user');
    localStorage.removeItem('permissions');
    apiClient.setAccessToken(null);
  },

  getCurrentUser(): User | null {
    const userStr = localStorage.getItem('user');
    if (!userStr) return null;
    try {
      return JSON.parse(userStr) as User;
    } catch (e) {
      console.error('Failed to parse stored user data', e);
      localStorage.removeItem('user');
      return null;
    }
  },

  async fetchPermissions(): Promise<string[]> {
    try {
      const res = await apiClient.get<PermissionsResponse>('/me/permissions');
      const permissions = res.permissions ?? [];
      localStorage.setItem('permissions', JSON.stringify(permissions));
      return permissions;
    } catch (e) {
      console.error('Failed to fetch permissions', e);
      return [];
    }
  },

  getStoredPermissions(): string[] {
    const permStr = localStorage.getItem('permissions');
    if (!permStr) return [];
    try {
      return JSON.parse(permStr) as string[];
    } catch (e) {
      console.error('Failed to parse stored permissions', e);
      localStorage.removeItem('permissions');
      return [];
    }
  },
};
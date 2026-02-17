import { User, LoginRequest, SignupRequest, AuthResponse } from '../types/user';
import { apiClient } from '../api/apiClient';

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

  logout(): void {
    void apiClient.post<void>('/auth/logout', undefined, { credentials: 'include' }).catch(() => undefined);
    localStorage.removeItem('user');
    apiClient.setAccessToken(null);
  },

  getCurrentUser(): User | null {
    const userStr = localStorage.getItem('user');
    if (!userStr) return null;
    try {
      return JSON.parse(userStr) as User;
    } catch {
      localStorage.removeItem('user');
      return null;
    }
  },
};
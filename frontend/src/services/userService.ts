import { User, LoginRequest, SignupRequest } from '../types/user';
import { apiClient } from '../api/apiClient';

export const userService = {
  async login(request: LoginRequest): Promise<User> {
    const user = await apiClient.post<User>('/auth/login', request, { credentials: 'include' });
    localStorage.setItem('user', JSON.stringify(user));
    return user;
  },

  async signup(request: SignupRequest): Promise<User> {
    const user = await apiClient.post<User>('/auth/signup', request, { credentials: 'include' });
    localStorage.setItem('user', JSON.stringify(user));
    return user;
  },

  logout(): void {
    localStorage.removeItem('user');
  },

  getCurrentUser(): User | null {
    const userStr = localStorage.getItem('user');
    if (!userStr) return null;
    return JSON.parse(userStr);
  },
}; 
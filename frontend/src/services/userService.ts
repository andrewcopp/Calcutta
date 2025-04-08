import { User, LoginRequest, SignupRequest } from '../types/user';

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

export const userService = {
  async login(request: LoginRequest): Promise<User> {
    const response = await fetch(`${API_URL}/api/auth/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      throw new Error('Login failed');
    }

    const user = await response.json();
    localStorage.setItem('user', JSON.stringify(user));
    return user;
  },

  async signup(request: SignupRequest): Promise<User> {
    const response = await fetch(`${API_URL}/api/auth/signup`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      throw new Error('Signup failed');
    }

    const user = await response.json();
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
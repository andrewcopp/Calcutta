import { describe, it, expect, beforeEach } from 'vitest';
import { http, HttpResponse } from 'msw';
import { server } from '../test/msw/server';
import { userService } from './userService';
import { USER_KEY, PERMISSIONS_KEY } from '../api/apiClient';

const BASE = 'http://localhost:8080/api';

beforeEach(() => {
  localStorage.clear();
});

describe('userService', () => {
  describe('login', () => {
    it('returns user and stores credentials', async () => {
      const user = await userService.login({ email: 'test@example.com', password: 'pass' });

      expect(user.id).toBe('user-1');
      expect(user.email).toBe('test@example.com');
      expect(localStorage.getItem(USER_KEY)).toBeTruthy();
    });

    it('throws when response missing accessToken', async () => {
      server.use(
        http.post(`${BASE}/auth/login`, () => {
          return HttpResponse.json({ user: { id: 'u' } });
        }),
      );

      await expect(userService.login({ email: 'a', password: 'b' })).rejects.toThrow();
    });
  });

  describe('signup', () => {
    it('returns user and stores credentials', async () => {
      const user = await userService.signup({
        email: 'test@example.com',
        firstName: 'Test',
        lastName: 'User',
        password: 'pass',
      });

      expect(user.id).toBe('user-1');
      expect(localStorage.getItem(USER_KEY)).toBeTruthy();
    });
  });

  describe('previewInvite', () => {
    it('returns invite preview', async () => {
      const preview = await userService.previewInvite('tok-abc');

      expect(preview.calcuttaName).toBe('Test Pool');
      expect(preview.commissionerName).toBe('Commissioner');
    });

    it('throws when response missing required field', async () => {
      server.use(
        http.get(`${BASE}/auth/invite/preview`, () => {
          return HttpResponse.json({ firstName: 'Test' });
        }),
      );

      await expect(userService.previewInvite('tok')).rejects.toThrow();
    });
  });

  describe('acceptInvite', () => {
    it('returns user and stores credentials', async () => {
      const user = await userService.acceptInvite('tok-abc', 'newpass');

      expect(user.id).toBe('user-1');
      expect(localStorage.getItem(USER_KEY)).toBeTruthy();
    });
  });

  describe('logout', () => {
    it('clears stored credentials', () => {
      localStorage.setItem(USER_KEY, '{}');
      localStorage.setItem(PERMISSIONS_KEY, '[]');

      userService.logout();

      expect(localStorage.getItem(USER_KEY)).toBeNull();
      expect(localStorage.getItem(PERMISSIONS_KEY)).toBeNull();
    });
  });

  describe('getCurrentUser', () => {
    it('returns null when no stored user', () => {
      expect(userService.getCurrentUser()).toBeNull();
    });

    it('returns parsed user from localStorage', () => {
      localStorage.setItem(USER_KEY, JSON.stringify({ id: 'u1', firstName: 'A', lastName: 'B', status: 'active' }));

      const user = userService.getCurrentUser();

      expect(user?.id).toBe('u1');
    });

    it('returns null and clears on corrupt data', () => {
      localStorage.setItem(USER_KEY, 'not-json');

      expect(userService.getCurrentUser()).toBeNull();
      expect(localStorage.getItem(USER_KEY)).toBeNull();
    });
  });

  describe('fetchPermissions', () => {
    it('returns permissions and stores them', async () => {
      const perms = await userService.fetchPermissions();

      expect(perms).toEqual(['admin', 'manage_calcuttas']);
      expect(localStorage.getItem(PERMISSIONS_KEY)).toBeTruthy();
    });

    it('returns empty array on failure', async () => {
      server.use(
        http.get(`${BASE}/me/permissions`, () => {
          return new HttpResponse(null, { status: 500 });
        }),
      );

      const perms = await userService.fetchPermissions();

      expect(perms).toEqual([]);
    });
  });

  describe('fetchProfile', () => {
    it('returns user profile', async () => {
      const profile = await userService.fetchProfile();

      expect(profile.id).toBe('user-1');
      expect(profile.roles).toEqual(['admin']);
    });
  });

  describe('getStoredPermissions', () => {
    it('returns empty array when nothing stored', () => {
      expect(userService.getStoredPermissions()).toEqual([]);
    });

    it('returns parsed permissions', () => {
      localStorage.setItem(PERMISSIONS_KEY, JSON.stringify(['admin']));

      expect(userService.getStoredPermissions()).toEqual(['admin']);
    });

    it('returns empty array and clears on corrupt data', () => {
      localStorage.setItem(PERMISSIONS_KEY, 'bad');

      expect(userService.getStoredPermissions()).toEqual([]);
      expect(localStorage.getItem(PERMISSIONS_KEY)).toBeNull();
    });
  });
});

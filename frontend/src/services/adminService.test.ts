import { describe, it, expect } from 'vitest';
import { http, HttpResponse } from 'msw';
import { server } from '../test/msw/server';
import { adminService } from './adminService';

const BASE = 'http://localhost:8080/api';

describe('adminService', () => {
  describe('listUsers', () => {
    it('returns parsed users list', async () => {
      const result = await adminService.listUsers();

      expect(result.items).toHaveLength(1);
      expect(result.items[0].email).toBe('test@example.com');
    });

    it('throws when user missing required field', async () => {
      server.use(
        http.get(`${BASE}/admin/users`, () => {
          return HttpResponse.json({ items: [{ id: 'u1' }] });
        }),
      );

      await expect(adminService.listUsers()).rejects.toThrow();
    });
  });

  describe('setUserEmail', () => {
    it('returns updated user response', async () => {
      const user = await adminService.setUserEmail('user-1', 'new@example.com');

      expect(user.email).toBe('new@example.com');
    });
  });

  describe('sendInvite', () => {
    it('returns invite response', async () => {
      const result = await adminService.sendInvite('user-1');

      expect(result.inviteToken).toBe('tok-invite');
      expect(result.status).toBe('invited');
    });
  });

  describe('getUser', () => {
    it('returns parsed user detail', async () => {
      const user = await adminService.getUser('user-1');

      expect(user.id).toBe('user-1');
      expect(user.roles).toHaveLength(1);
      expect(user.roles[0].key).toBe('admin');
    });

    it('throws when user detail missing roles array', async () => {
      server.use(
        http.get(`${BASE}/admin/users/:id`, () => {
          return HttpResponse.json({
            id: 'user-1',
            email: 'test@example.com',
            firstName: 'T',
            lastName: 'U',
            status: 'active',
            permissions: [],
            createdAt: '2026-01-01T00:00:00Z',
            updatedAt: '2026-01-01T00:00:00Z',
          });
        }),
      );

      await expect(adminService.getUser('user-1')).rejects.toThrow();
    });
  });

  describe('grantRole', () => {
    it('completes without error', async () => {
      await expect(adminService.grantRole('user-1', 'admin')).resolves.toBeUndefined();
    });
  });

  describe('revokeRole', () => {
    it('completes without error', async () => {
      await expect(adminService.revokeRole('user-1', 'admin')).resolves.toBeUndefined();
    });
  });

  describe('listApiKeys', () => {
    it('returns parsed API keys', async () => {
      const result = await adminService.listApiKeys();

      expect(result.items).toHaveLength(1);
      expect(result.items[0].label).toBe('Test Key');
    });
  });

  describe('createApiKey', () => {
    it('returns created API key with secret', async () => {
      const result = await adminService.createApiKey('New Key');

      expect(result.id).toBe('key-2');
      expect(result.key).toBe('secret-key-value');
    });
  });

  describe('revokeApiKey', () => {
    it('completes without error', async () => {
      await expect(adminService.revokeApiKey('key-1')).resolves.toBeUndefined();
    });
  });

  describe('startTournamentImport', () => {
    it('returns import start response', async () => {
      const file = new File(['data'], 'test.zip', { type: 'application/zip' });
      const result = await adminService.startTournamentImport(file);

      expect(result.uploadId).toBe('upload-1');
      expect(result.status).toBe('pending');
    });
  });

  describe('getTournamentImportStatus', () => {
    it('returns import status response', async () => {
      const result = await adminService.getTournamentImportStatus('upload-1');

      expect(result.status).toBe('succeeded');
      expect(result.filename).toBe('data.zip');
    });

    it('throws when status has invalid enum value', async () => {
      server.use(
        http.get(`${BASE}/admin/tournament-imports/import/:id`, () => {
          return HttpResponse.json({
            uploadId: 'upload-1',
            filename: 'data.zip',
            sha256: 'abc',
            sizeBytes: 1024,
            status: 'invalid_status',
          });
        }),
      );

      await expect(adminService.getTournamentImportStatus('upload-1')).rejects.toThrow();
    });
  });
});

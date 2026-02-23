import { http, HttpResponse } from 'msw';

const BASE = 'http://localhost:8080/api';

const validUser = {
  id: 'user-1',
  email: 'test@example.com',
  firstName: 'Test',
  lastName: 'User',
  status: 'active',
  invitedAt: null,
  lastInviteSentAt: null,
  inviteExpiresAt: null,
  inviteConsumedAt: null,
  createdAt: '2026-01-01T00:00:00Z',
  updatedAt: '2026-01-01T00:00:00Z',
  roles: ['admin'],
  permissions: ['admin'],
};

export const adminHandlers = [
  http.get(`${BASE}/admin/users/:id`, () => {
    return HttpResponse.json({
      id: 'user-1',
      email: 'test@example.com',
      firstName: 'Test',
      lastName: 'User',
      status: 'active',
      roles: [{ key: 'admin', scopeType: 'global' }],
      permissions: ['admin'],
      createdAt: '2026-01-01T00:00:00Z',
      updatedAt: '2026-01-01T00:00:00Z',
    });
  }),

  http.get(`${BASE}/admin/users`, () => {
    return HttpResponse.json({ items: [validUser] });
  }),

  http.patch(`${BASE}/admin/users/:id/email`, () => {
    return HttpResponse.json({
      id: 'user-1',
      email: 'new@example.com',
      firstName: 'Test',
      lastName: 'User',
      status: 'active',
    });
  }),

  http.post(`${BASE}/admin/users/:id/invite/send`, () => {
    return HttpResponse.json({
      userId: 'user-1',
      email: 'test@example.com',
      inviteToken: 'tok-invite',
      inviteExpiresAt: '2026-02-01T00:00:00Z',
      status: 'invited',
    });
  }),

  http.post(`${BASE}/admin/users/:id/roles`, () => {
    return new HttpResponse(null, { status: 204 });
  }),

  http.delete(`${BASE}/admin/users/:id/roles/:roleKey`, () => {
    return new HttpResponse(null, { status: 204 });
  }),

  http.get(`${BASE}/admin/api-keys`, () => {
    return HttpResponse.json({
      items: [{ id: 'key-1', label: 'Test Key', createdAt: '2026-01-01T00:00:00Z' }],
    });
  }),

  http.post(`${BASE}/admin/api-keys`, () => {
    return HttpResponse.json({
      id: 'key-2',
      key: 'secret-key-value',
      label: 'New Key',
      createdAt: '2026-01-01T00:00:00Z',
    });
  }),

  http.delete(`${BASE}/admin/api-keys/:id`, () => {
    return new HttpResponse(null, { status: 204 });
  }),

  http.post(`${BASE}/admin/tournament-imports/import`, () => {
    return HttpResponse.json({
      uploadId: 'upload-1',
      status: 'pending',
      filename: 'data.zip',
      sha256: 'abc123',
      sizeBytes: 1024,
    });
  }),

  http.get(`${BASE}/admin/tournament-imports/import/:id`, () => {
    return HttpResponse.json({
      uploadId: 'upload-1',
      filename: 'data.zip',
      sha256: 'abc123',
      sizeBytes: 1024,
      status: 'succeeded',
    });
  }),
];

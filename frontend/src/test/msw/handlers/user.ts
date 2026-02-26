import { http, HttpResponse } from 'msw';

const BASE = 'http://localhost:8080/api/v1';

const validUser = {
  id: 'user-1',
  email: 'test@example.com',
  firstName: 'Test',
  lastName: 'User',
  status: 'active',
  createdAt: '2026-01-01T00:00:00Z',
  updatedAt: '2026-01-01T00:00:00Z',
};

export const userHandlers = [
  http.post(`${BASE}/auth/login`, () => {
    return HttpResponse.json({
      user: validUser,
      accessToken: 'tok-123',
    });
  }),

  http.get(`${BASE}/auth/invite/preview`, () => {
    return HttpResponse.json({
      firstName: 'Test',
      poolName: 'Test Pool',
      commissionerName: 'Commissioner',
    });
  }),

  http.post(`${BASE}/auth/invite/accept`, () => {
    return HttpResponse.json({
      user: validUser,
      accessToken: 'tok-789',
    });
  }),

  http.post(`${BASE}/auth/logout`, () => {
    return new HttpResponse(null, { status: 204 });
  }),

  http.get(`${BASE}/me/permissions`, () => {
    return HttpResponse.json({
      permissions: ['admin', 'manage_pools'],
    });
  }),

  http.get(`${BASE}/me/profile`, () => {
    return HttpResponse.json({
      id: 'user-1',
      email: 'test@example.com',
      firstName: 'Test',
      lastName: 'User',
      status: 'active',
      roles: ['admin'],
      permissions: ['admin'],
      createdAt: '2026-01-01T00:00:00Z',
      updatedAt: '2026-01-01T00:00:00Z',
    });
  }),
];

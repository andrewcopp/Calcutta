import { describe, it, expect, vi, beforeAll } from 'vitest';

// Stub import.meta.env before importing the module
beforeAll(() => {
  vi.stubEnv('VITE_API_URL', 'http://localhost:8080');
});

describe('normalizeBaseUrl', () => {
  it('strips single trailing slash', async () => {
    // GIVEN a URL with trailing slash
    const { normalizeBaseUrl } = await import('./apiClient');

    // WHEN normalizing
    const result = normalizeBaseUrl('http://localhost:8080/');

    // THEN trailing slash is removed
    expect(result).toBe('http://localhost:8080');
  });

  it('strips multiple trailing slashes', async () => {
    // GIVEN a URL with multiple trailing slashes
    const { normalizeBaseUrl } = await import('./apiClient');

    // WHEN normalizing
    const result = normalizeBaseUrl('http://localhost:8080///');

    // THEN all trailing slashes are removed
    expect(result).toBe('http://localhost:8080');
  });

  it('is a no-op without trailing slash', async () => {
    // GIVEN a URL without trailing slash
    const { normalizeBaseUrl } = await import('./apiClient');

    // WHEN normalizing
    const result = normalizeBaseUrl('http://localhost:8080');

    // THEN URL is unchanged
    expect(result).toBe('http://localhost:8080');
  });
});

describe('isAuthUrl', () => {
  it('returns true for auth path', async () => {
    // GIVEN a URL containing /api/auth/
    const { isAuthUrl } = await import('./apiClient');

    // WHEN checking
    const result = isAuthUrl('http://localhost:8080/api/auth/refresh');

    // THEN returns true
    expect(result).toBe(true);
  });

  it('returns false for non-auth path', async () => {
    // GIVEN a URL not containing /api/auth/
    const { isAuthUrl } = await import('./apiClient');

    // WHEN checking
    const result = isAuthUrl('http://localhost:8080/api/calcuttas');

    // THEN returns false
    expect(result).toBe(false);
  });
});

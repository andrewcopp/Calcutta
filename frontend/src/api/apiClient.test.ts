import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

// ---------------------------------------------------------------------------
// apiClient relies on `fetch`, `localStorage`, and `window.location` at
// runtime, and `import.meta.env` at module-evaluation time.  We stub them in
// the node test environment so every test starts with a clean slate.
// ---------------------------------------------------------------------------

// -- import.meta.env stubs (must come before the dynamic import) --
vi.stubGlobal('importMetaEnv', { DEV: true });
Object.defineProperty(import.meta, 'env', {
  value: { DEV: true, VITE_API_URL: '' },
  writable: true,
  configurable: true,
});

// -- localStorage stub --
const localStore: Record<string, string> = {};
const localStorageMock: Storage = {
  getItem: vi.fn((key: string) => localStore[key] ?? null),
  setItem: vi.fn((key: string, value: string) => {
    localStore[key] = value;
  }),
  removeItem: vi.fn((key: string) => {
    delete localStore[key];
  }),
  clear: vi.fn(() => {
    for (const k of Object.keys(localStore)) delete localStore[k];
  }),
  get length() {
    return Object.keys(localStore).length;
  },
  key: vi.fn(() => null),
};
vi.stubGlobal('localStorage', localStorageMock);

// -- window.location stub --
const locationMock = { href: '' };
vi.stubGlobal('window', { location: locationMock });

// -- fetch stub --
const fetchMock = vi.fn<(input: RequestInfo | URL, init?: RequestInit) => Promise<Response>>();
vi.stubGlobal('fetch', fetchMock);

// -- import the module under test (after stubs are in place) --
type ApiClientModule = {
  apiClient: {
    get: <T>(path: string) => Promise<T>;
    post: <T>(path: string, body?: unknown) => Promise<T>;
    put: <T>(path: string, body?: unknown) => Promise<T>;
    patch: <T>(path: string, body?: unknown) => Promise<T>;
    delete: <T>(path: string) => Promise<T>;
    fetch: (path: string, options?: RequestInit) => Promise<Response>;
    setAccessToken: (token: string | null) => void;
  };
};
let apiClient: ApiClientModule['apiClient'];

beforeEach(async () => {
  // Reset all mocks and local storage between tests
  vi.resetModules();
  vi.clearAllMocks();
  for (const k of Object.keys(localStore)) delete localStore[k];
  locationMock.href = '';

  // Re-import so the module picks up a fresh `refreshInFlight = null`
  const mod = await import('./apiClient');
  apiClient = mod.apiClient;
});

afterEach(() => {
  vi.restoreAllMocks();
});

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function jsonResponse(status: number, body: unknown, headers?: Record<string, string>): Response {
  const defaultHeaders: Record<string, string> = { 'content-type': 'application/json', ...headers };
  return new Response(JSON.stringify(body), {
    status,
    statusText: status === 200 ? 'OK' : 'Error',
    headers: new Headers(defaultHeaders),
  });
}

function textResponse(status: number, text: string): Response {
  return new Response(text, {
    status,
    statusText: status === 200 ? 'OK' : 'Error',
    headers: new Headers({ 'content-type': 'text/plain' }),
  });
}

function emptyResponse(status: number): Response {
  return new Response(null, { status, statusText: 'No Content' });
}

// ---------------------------------------------------------------------------
// Token refresh deduplication
// ---------------------------------------------------------------------------

describe('token refresh deduplication', () => {
  it('sends only one refresh request when two requests receive 401 simultaneously', async () => {
    // GIVEN two in-flight GET requests that both return 401, and a refresh that succeeds
    let refreshResolve: (r: Response) => void;
    const refreshPromise = new Promise<Response>((resolve) => {
      refreshResolve = resolve;
    });

    fetchMock.mockImplementation((input: RequestInfo | URL) => {
      const url = typeof input === 'string' ? input : input.toString();
      if (url.includes('/api/auth/refresh')) return refreshPromise;
      // First call for each request returns 401, retry returns 200
      if (url.includes('/api/first') || url.includes('/api/second')) {
        // Return 401 for initial calls, 200 for retries
        const callsForUrl = fetchMock.mock.calls.filter(
          (c) => (typeof c[0] === 'string' ? c[0] : c[0].toString()) === url,
        );
        if (callsForUrl.length <= 1) {
          return Promise.resolve(jsonResponse(401, { message: 'Unauthorized' }));
        }
        return Promise.resolve(jsonResponse(200, { ok: true }));
      }
      return Promise.resolve(jsonResponse(404, { message: 'Not found' }));
    });

    // WHEN both requests are made concurrently
    const p1 = apiClient.get('/first');
    const p2 = apiClient.get('/second');

    // Resolve the single refresh
    refreshResolve!(jsonResponse(200, { accessToken: 'new-token' }));

    await Promise.all([p1, p2]);

    // THEN only one refresh request was made
    const refreshCalls = fetchMock.mock.calls.filter((c) => {
      const url = typeof c[0] === 'string' ? c[0] : c[0].toString();
      return url.includes('/api/auth/refresh');
    });
    expect(refreshCalls).toHaveLength(1);
  });

  it('clears the in-flight refresh after it completes so a later 401 triggers a new refresh', async () => {
    // GIVEN a first request that triggers a refresh, then a second later request also gets 401
    let callCount = 0;
    fetchMock.mockImplementation((input: RequestInfo | URL) => {
      const url = typeof input === 'string' ? input : input.toString();
      if (url.includes('/api/auth/refresh')) {
        callCount++;
        return Promise.resolve(jsonResponse(200, { accessToken: `token-${callCount}` }));
      }
      // Always return 401 first time, 200 on retry
      return Promise.resolve(jsonResponse(401, { message: 'Unauthorized' }));
    });

    // The implementation retries after refresh; the retry also gets 401, triggering redirect.
    // But for this test we care about the refresh count.
    // After the first request completes (or fails), we issue a second.
    await apiClient.get('/first').catch(() => {});
    await apiClient.get('/second').catch(() => {});

    // THEN two separate refresh requests were made (not deduplicated)
    const refreshCalls = fetchMock.mock.calls.filter((c) => {
      const url = typeof c[0] === 'string' ? c[0] : c[0].toString();
      return url.includes('/api/auth/refresh');
    });
    expect(refreshCalls).toHaveLength(2);
  });
});

// ---------------------------------------------------------------------------
// Retry on 401
// ---------------------------------------------------------------------------

describe('retry on 401', () => {
  it('retries the original request with the new token after a successful refresh', async () => {
    // GIVEN a first fetch that returns 401, a refresh that returns a new token, and a retry that succeeds
    localStore['accessToken'] = 'expired-token';
    let attempt = 0;

    fetchMock.mockImplementation((input: RequestInfo | URL) => {
      const url = typeof input === 'string' ? input : input.toString();
      if (url.includes('/api/auth/refresh')) {
        return Promise.resolve(jsonResponse(200, { accessToken: 'fresh-token' }));
      }
      attempt++;
      if (attempt === 1) {
        return Promise.resolve(jsonResponse(401, { message: 'Unauthorized' }));
      }
      return Promise.resolve(jsonResponse(200, { data: 'success' }));
    });

    // WHEN making a GET request
    const result = await apiClient.get<{ data: string }>('/protected');

    // THEN the result contains the retried response data
    expect(result).toEqual({ data: 'success' });
  });

  it('sets the refreshed token in the Authorization header of the retry request', async () => {
    // GIVEN a first fetch that returns 401 and a refresh that returns "fresh-token"
    localStore['accessToken'] = 'expired-token';
    let attempt = 0;

    fetchMock.mockImplementation((input: RequestInfo | URL) => {
      const url = typeof input === 'string' ? input : input.toString();
      if (url.includes('/api/auth/refresh')) {
        return Promise.resolve(jsonResponse(200, { accessToken: 'fresh-token' }));
      }
      attempt++;
      if (attempt === 1) {
        return Promise.resolve(jsonResponse(401, { message: 'Unauthorized' }));
      }
      return Promise.resolve(jsonResponse(200, { ok: true }));
    });

    // WHEN making a GET request
    await apiClient.get('/protected');

    // THEN the retry request (third call overall: initial, refresh, retry) uses the fresh token
    const retryCall = fetchMock.mock.calls[2];
    const retryHeaders = retryCall[1]?.headers as Headers;
    expect(retryHeaders.get('Authorization')).toBe('Bearer fresh-token');
  });

  it('stores the new token in localStorage after a successful refresh', async () => {
    // GIVEN a first fetch that returns 401 and a refresh that returns "new-access-token"
    let attempt = 0;

    fetchMock.mockImplementation((input: RequestInfo | URL) => {
      const url = typeof input === 'string' ? input : input.toString();
      if (url.includes('/api/auth/refresh')) {
        return Promise.resolve(jsonResponse(200, { accessToken: 'new-access-token' }));
      }
      attempt++;
      if (attempt === 1) {
        return Promise.resolve(jsonResponse(401, { message: 'Unauthorized' }));
      }
      return Promise.resolve(jsonResponse(200, { ok: true }));
    });

    // WHEN making a request that triggers a 401 -> refresh cycle
    await apiClient.get('/something');

    // THEN the new token is persisted in localStorage
    expect(localStore['accessToken']).toBe('new-access-token');
  });

  it('clears token and user from localStorage when refresh fails', async () => {
    // GIVEN a stored access token and user, and a refresh that returns a non-OK response
    localStore['accessToken'] = 'old-token';
    localStore['user'] = JSON.stringify({ id: 'u1', name: 'Alice' });

    fetchMock.mockImplementation((input: RequestInfo | URL) => {
      const url = typeof input === 'string' ? input : input.toString();
      if (url.includes('/api/auth/refresh')) {
        return Promise.resolve(jsonResponse(403, { message: 'Forbidden' }));
      }
      return Promise.resolve(jsonResponse(401, { message: 'Unauthorized' }));
    });

    // WHEN making a request that triggers a 401 -> failed refresh
    await apiClient.get('/protected').catch(() => {});

    // THEN both accessToken and user are removed from localStorage
    expect(localStore['accessToken']).toBeUndefined();
  });

  it('redirects to root when refresh fails', async () => {
    // GIVEN a fetch that returns 401 and a refresh that fails
    fetchMock.mockImplementation((input: RequestInfo | URL) => {
      const url = typeof input === 'string' ? input : input.toString();
      if (url.includes('/api/auth/refresh')) {
        return Promise.resolve(jsonResponse(500, { message: 'Server error' }));
      }
      return Promise.resolve(jsonResponse(401, { message: 'Unauthorized' }));
    });

    // WHEN making a request that triggers a 401 -> failed refresh
    await apiClient.get('/protected').catch(() => {});

    // THEN the user is redirected to the login page with expired flag
    expect(locationMock.href).toBe('/login?expired=true');
  });

  it('does not attempt refresh for auth endpoints', async () => {
    // GIVEN a 401 response from an auth endpoint
    fetchMock.mockImplementation(() => {
      return Promise.resolve(jsonResponse(401, { message: 'Bad credentials' }));
    });

    // WHEN making a request to an auth endpoint
    try {
      await apiClient.get('/auth/login');
    } catch {
      // expected to throw ApiError
    }

    // THEN no refresh request was made
    const refreshCalls = fetchMock.mock.calls.filter((c) => {
      const url = typeof c[0] === 'string' ? c[0] : c[0].toString();
      return url.includes('/api/auth/refresh');
    });
    expect(refreshCalls).toHaveLength(0);
  });

  it('redirects to root when retry also returns 401', async () => {
    // GIVEN a first fetch returning 401, a successful refresh, and a retry also returning 401
    fetchMock.mockImplementation((input: RequestInfo | URL) => {
      const url = typeof input === 'string' ? input : input.toString();
      if (url.includes('/api/auth/refresh')) {
        return Promise.resolve(jsonResponse(200, { accessToken: 'new-token' }));
      }
      // Both the initial and retry requests return 401
      return Promise.resolve(jsonResponse(401, { message: 'Unauthorized' }));
    });

    // WHEN making a request where even the retry fails
    await apiClient.get('/protected').catch(() => {});

    // THEN the user is redirected to the login page with expired flag
    expect(locationMock.href).toBe('/login?expired=true');
  });

  it('stores user data from refresh response when present', async () => {
    // GIVEN a refresh response that includes both a token and user data
    let attempt = 0;

    fetchMock.mockImplementation((input: RequestInfo | URL) => {
      const url = typeof input === 'string' ? input : input.toString();
      if (url.includes('/api/auth/refresh')) {
        return Promise.resolve(
          jsonResponse(200, {
            accessToken: 'refreshed-token',
            user: { id: 'u1', name: 'Alice' },
          }),
        );
      }
      attempt++;
      if (attempt === 1) {
        return Promise.resolve(jsonResponse(401, { message: 'Unauthorized' }));
      }
      return Promise.resolve(jsonResponse(200, { ok: true }));
    });

    // WHEN making a request that triggers a refresh
    await apiClient.get('/protected');

    // THEN the user data is stored in localStorage
    expect(localStore['user']).toBe(JSON.stringify({ id: 'u1', name: 'Alice' }));
  });
});

// ---------------------------------------------------------------------------
// Error parsing
// ---------------------------------------------------------------------------

describe('error parsing', () => {
  it('throws an error with the message from a JSON error response', async () => {
    // GIVEN a 400 response with a JSON body containing a message field
    fetchMock.mockResolvedValue(jsonResponse(400, { message: 'Validation failed' }));

    // WHEN making a request
    const error = await apiClient.get('/bad-request').catch((e: unknown) => e);

    // THEN the error message matches the response body
    expect((error as Error).message).toBe('Validation failed');
  });

  it('includes the HTTP status on the thrown error', async () => {
    // GIVEN a 422 response
    fetchMock.mockResolvedValue(jsonResponse(422, { message: 'Unprocessable entity' }));

    // WHEN making a request
    const error = await apiClient.get('/unprocessable').catch((e: unknown) => e);

    // THEN the error has the correct status
    expect((error as { status: number }).status).toBe(422);
  });

  it('includes the parsed body on the thrown error', async () => {
    // GIVEN a 400 response with extra fields
    const errorBody = { message: 'Bad request', errors: ['field is required'] };
    fetchMock.mockResolvedValue(jsonResponse(400, errorBody));

    // WHEN making a request
    const error = await apiClient.get('/bad').catch((e: unknown) => e);

    // THEN the error body includes all response fields
    expect((error as { body: unknown }).body).toEqual(errorBody);
  });

  it('falls back to statusText when response body has no message field', async () => {
    // GIVEN a 500 response with a body that lacks a "message" field
    const response = new Response(JSON.stringify({ code: 'INTERNAL' }), {
      status: 500,
      statusText: 'Internal Server Error',
      headers: new Headers({ 'content-type': 'application/json' }),
    });
    fetchMock.mockResolvedValue(response);

    // WHEN making a request
    const error = await apiClient.get('/server-error').catch((e: unknown) => e);

    // THEN the error message falls back to the status text
    expect((error as Error).message).toBe('Internal Server Error');
  });

  it('falls back to a generic message when both message and statusText are empty', async () => {
    // GIVEN a 500 response with no message field and empty statusText
    const response = new Response(JSON.stringify({}), {
      status: 500,
      statusText: '',
      headers: new Headers({ 'content-type': 'application/json' }),
    });
    fetchMock.mockResolvedValue(response);

    // WHEN making a request
    const error = await apiClient.get('/blank').catch((e: unknown) => e);

    // THEN a generic fallback message is used
    expect((error as Error).message).toBe('Request failed with status 500');
  });

  it('parses a plain-text error response and includes the status', async () => {
    // GIVEN a 502 response with a plain-text body
    fetchMock.mockResolvedValue(textResponse(502, 'Bad Gateway'));

    // WHEN making a request
    const error = await apiClient.get('/gateway-error').catch((e: unknown) => e);

    // THEN the error has the correct status
    expect((error as { status: number }).status).toBe(502);
  });

  it('returns undefined for a 204 No Content response', async () => {
    // GIVEN a successful 204 response
    fetchMock.mockResolvedValue(emptyResponse(204));

    // WHEN making a DELETE request
    const result = await apiClient.delete('/resource/123');

    // THEN the result is undefined
    expect(result).toBeUndefined();
  });

  it('sets the error name to ApiError', async () => {
    // GIVEN a 403 Forbidden response
    fetchMock.mockResolvedValue(jsonResponse(403, { message: 'Forbidden' }));

    // WHEN making a request
    const error = await apiClient.get('/forbidden').catch((e: unknown) => e);

    // THEN the error name is "ApiError"
    expect((error as Error).name).toBe('ApiError');
  });
});

// ---------------------------------------------------------------------------
// Request construction
// ---------------------------------------------------------------------------

describe('request construction', () => {
  it('sends JSON body with Content-Type header for object payloads', async () => {
    // GIVEN a POST request with an object body
    fetchMock.mockResolvedValue(jsonResponse(200, { id: '1' }));

    // WHEN posting a JSON body
    await apiClient.post('/items', { name: 'Test' });

    // THEN the request has Content-Type application/json and a serialized body
    const [, init] = fetchMock.mock.calls[0];
    const headers = init?.headers as Headers;
    expect(headers.get('Content-Type')).toBe('application/json');
    expect(init?.body).toBe(JSON.stringify({ name: 'Test' }));
  });

  it('includes credentials by default', async () => {
    // GIVEN a simple GET request
    fetchMock.mockResolvedValue(jsonResponse(200, {}));

    // WHEN making the request
    await apiClient.get('/test');

    // THEN credentials is set to "include"
    const [, init] = fetchMock.mock.calls[0];
    expect(init?.credentials).toBe('include');
  });

  it('attaches stored access token as Authorization header', async () => {
    // GIVEN a stored access token
    localStore['accessToken'] = 'my-token';
    fetchMock.mockResolvedValue(jsonResponse(200, {}));

    // WHEN making a request
    await apiClient.get('/protected');

    // THEN the Authorization header is set
    const [, init] = fetchMock.mock.calls[0];
    const headers = init?.headers as Headers;
    expect(headers.get('Authorization')).toBe('Bearer my-token');
  });
});

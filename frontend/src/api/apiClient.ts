export class ApiError extends Error {
  status: number;
  body: unknown;

  constructor(message: string, status: number, body: unknown) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.body = body;
  }
}

function normalizeBaseUrl(v: string): string {
  return v.replace(/\/+$/, '');
}

function resolveApiUrl(): string {
  const configured = (import.meta.env.VITE_API_URL || import.meta.env.VITE_API_BASE_URL) as string | undefined;
  if (configured && configured.trim() !== '') return normalizeBaseUrl(configured);
  if (import.meta.env.DEV) return 'http://localhost:8080';
  throw new Error('Missing required frontend env var: VITE_API_URL (or VITE_API_BASE_URL)');
}

export const API_URL = resolveApiUrl();
export const API_BASE_URL = `${API_URL}/api`;

const ACCESS_TOKEN_KEY = 'accessToken';

function getAccessToken(): string | null {
  return localStorage.getItem(ACCESS_TOKEN_KEY);
}

function setAccessToken(token: string | null): void {
  if (token) {
    localStorage.setItem(ACCESS_TOKEN_KEY, token);
    return;
  }
  localStorage.removeItem(ACCESS_TOKEN_KEY);
}

let refreshInFlight: Promise<string | null> | null = null;

async function refreshAccessToken(): Promise<string | null> {
  if (refreshInFlight) return refreshInFlight;

  refreshInFlight = (async () => {
    try {
      const res = await fetch(`${API_BASE_URL}/auth/refresh`, {
        method: 'POST',
        credentials: 'include',
        headers: {
          Accept: 'application/json',
        },
      });

      if (!res.ok) return null;
      const body = (await res.json().catch(() => undefined)) as any;
      const tok = body?.accessToken;
      if (typeof tok !== 'string' || tok.trim() === '') return null;

      setAccessToken(tok);
      if (body?.user) {
        localStorage.setItem('user', JSON.stringify(body.user));
      }

      return tok;
    } catch {
      return null;
    } finally {
      refreshInFlight = null;
    }
  })();

  return refreshInFlight;
}

function isAuthUrl(url: string): boolean {
  return url.includes('/api/auth/');
}

async function fetchWithAuth(url: string, init: RequestInit, allowRefresh: boolean): Promise<Response> {
  const headers = new Headers(init.headers);
  const tok = getAccessToken();
  if (tok) {
    headers.set('Authorization', `Bearer ${tok}`);
  }

  const response = await fetch(url, { ...init, headers });

  if (response.status !== 401 || !allowRefresh || isAuthUrl(url)) {
    return response;
  }

  const refreshed = await refreshAccessToken();
  if (!refreshed) {
    setAccessToken(null);
    localStorage.removeItem('user');
    if (typeof window !== 'undefined') {
      window.location.href = '/';
    }
    return response;
  }

  const retryHeaders = new Headers(init.headers);
  retryHeaders.set('Authorization', `Bearer ${refreshed}`);

  const retried = await fetch(url, { ...init, headers: retryHeaders });
  if (retried.status === 401) {
    setAccessToken(null);
    localStorage.removeItem('user');
    if (typeof window !== 'undefined') {
      window.location.href = '/';
    }
  }

  return retried;
}

type RequestOptions = Omit<RequestInit, 'body'> & { body?: unknown };

async function request<T>(path: string, options?: RequestOptions): Promise<T> {
  const url = path.startsWith('http') ? path : `${API_BASE_URL}${path.startsWith('/') ? '' : '/'}${path}`;

  const { body: rawBody, ...requestInit } = options ?? {};

  const headers = new Headers(requestInit.headers);
  headers.set('Accept', 'application/json');

  const init: RequestInit = {
    ...requestInit,
    headers,
    credentials: requestInit.credentials ?? 'include',
  };

  if (rawBody !== undefined && rawBody !== null) {
    if (rawBody instanceof FormData) {
      init.body = rawBody;
    } else if (typeof rawBody === 'string') {
      init.body = rawBody;
    } else {
      headers.set('Content-Type', headers.get('Content-Type') ?? 'application/json');
      init.body = JSON.stringify(rawBody);
    }
  }

  const response = await fetchWithAuth(url, init, true);

  if (response.status === 204) {
    return undefined as T;
  }

  const contentType = response.headers.get('content-type') || '';
  const isJson = contentType.includes('application/json');

  const parseBody = async () => {
    if (isJson) return response.json().catch(() => undefined);
    return response.text().catch(() => undefined);
  };

  const body = await parseBody();

  if (!response.ok) {
    const message =
      (body && typeof body === 'object' && 'message' in (body as Record<string, unknown>)
        ? String((body as Record<string, unknown>).message)
        : response.statusText) || `Request failed with status ${response.status}`;

    throw new ApiError(message, response.status, body);
  }

  return body as T;
}

export const apiClient = {
  request,
  fetch: (path: string, options?: RequestInit) => {
    const url = path.startsWith('http') ? path : `${API_BASE_URL}${path.startsWith('/') ? '' : '/'}${path}`;

    const headers = new Headers(options?.headers);
    if (!headers.has('Accept')) {
      headers.set('Accept', 'application/json');
    }

    const init: RequestInit = {
      ...options,
      headers,
      credentials: options?.credentials ?? 'include',
    };

    return fetchWithAuth(url, init, true);
  },
  getAccessToken,
  setAccessToken,
  get: <T>(path: string, options?: Omit<RequestOptions, 'method'>) => request<T>(path, { ...options, method: 'GET' }),
  post: <T>(path: string, body?: unknown, options?: Omit<RequestOptions, 'method' | 'body'>) =>
    request<T>(path, { ...options, method: 'POST', body }),
  put: <T>(path: string, body?: unknown, options?: Omit<RequestOptions, 'method' | 'body'>) =>
    request<T>(path, { ...options, method: 'PUT', body }),
  patch: <T>(path: string, body?: unknown, options?: Omit<RequestOptions, 'method' | 'body'>) =>
    request<T>(path, { ...options, method: 'PATCH', body }),
  delete: <T>(path: string, options?: Omit<RequestOptions, 'method'>) => request<T>(path, { ...options, method: 'DELETE' }),
};

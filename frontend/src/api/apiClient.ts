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

const API_URL = import.meta.env.VITE_API_URL || import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';
const API_BASE_URL = `${API_URL}/api`;

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

  const response = await fetch(url, init);

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
  get: <T>(path: string, options?: Omit<RequestOptions, 'method'>) => request<T>(path, { ...options, method: 'GET' }),
  post: <T>(path: string, body?: unknown, options?: Omit<RequestOptions, 'method' | 'body'>) =>
    request<T>(path, { ...options, method: 'POST', body }),
  put: <T>(path: string, body?: unknown, options?: Omit<RequestOptions, 'method' | 'body'>) =>
    request<T>(path, { ...options, method: 'PUT', body }),
  patch: <T>(path: string, body?: unknown, options?: Omit<RequestOptions, 'method' | 'body'>) =>
    request<T>(path, { ...options, method: 'PATCH', body }),
  delete: <T>(path: string, options?: Omit<RequestOptions, 'method'>) => request<T>(path, { ...options, method: 'DELETE' }),
};

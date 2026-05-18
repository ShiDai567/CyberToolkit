export interface AuthUser {
  id: string;
  email: string;
  displayName: string;
  role: string;
  createdAt?: string;
}

export interface AuthSession {
  accessToken: string;
  refreshToken: string;
  user: AuthUser;
}

const STORAGE_KEY = 'cybertoolkit.auth';

export function loadStoredSession(): AuthSession | null {
  if (typeof window === 'undefined') {
    return null;
  }

  const raw = window.localStorage.getItem(STORAGE_KEY);
  if (!raw) {
    return null;
  }

  try {
    return JSON.parse(raw) as AuthSession;
  } catch {
    window.localStorage.removeItem(STORAGE_KEY);
    return null;
  }
}

export function storeSession(session: AuthSession): void {
  if (typeof window === 'undefined') {
    return;
  }

  window.localStorage.setItem(STORAGE_KEY, JSON.stringify(session));
}

export function clearSession(): void {
  if (typeof window === 'undefined') {
    return;
  }

  window.localStorage.removeItem(STORAGE_KEY);
}

export function getClientAPIBaseURL(): string {
  const url = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080';
  return url.trim().replace(/\/+$/, '');
}

interface ApiError {
  error?: {
    code?: string;
    message?: string;
  };
}

async function postAuthJSON<T>(path: string, body: unknown, token?: string): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  };
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const response = await fetch(`${getClientAPIBaseURL()}${path}`, {
    method: 'POST',
    headers,
    body: JSON.stringify(body),
  });

  const payload = (await response.json()) as T | ApiError;

  if (!response.ok) {
    const errorPayload = payload as ApiError;
    const message = errorPayload.error?.message || `Request failed: ${response.status}`;
    throw new Error(message);
  }

  return payload as T;
}

interface AuthResponse {
  data: AuthSession;
}

export async function login(email: string, password: string): Promise<AuthSession> {
  const payload = await postAuthJSON<AuthResponse>('/api/v1/auth/login', { email, password });
  return payload.data;
}

export async function register(email: string, password: string, displayName: string): Promise<AuthSession> {
  const payload = await postAuthJSON<AuthResponse>('/api/v1/auth/register', { email, password, displayName });
  return payload.data;
}

export async function logout(token: string): Promise<void> {
  await postAuthJSON<{ data: { loggedOut: boolean } }>('/api/v1/auth/logout', {}, token);
}

export async function me(token: string): Promise<AuthUser> {
  const response = await fetch(`${getClientAPIBaseURL()}/api/v1/auth/me`, {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  const payload = (await response.json()) as { data: AuthUser } | ApiError;
  if (!response.ok) {
    const errorPayload = payload as ApiError;
    throw new Error(errorPayload.error?.message || `Request failed: ${response.status}`);
  }

  return (payload as { data: AuthUser }).data;
}

export async function refreshToken(refreshTokenValue: string): Promise<AuthSession> {
  const payload = await postAuthJSON<AuthResponse>('/api/v1/auth/refresh', { refreshToken: refreshTokenValue });
  return payload.data;
}

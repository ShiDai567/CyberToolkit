export interface AuthUser {
  id: string;
  username: string;
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

async function requestAuthJSON<T>(
  method: string,
  path: string,
  body?: unknown,
  token?: string,
): Promise<T> {
  const headers: Record<string, string> = {};
  if (body !== undefined) {
    headers['Content-Type'] = 'application/json';
  }
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const response = await fetch(`${getClientAPIBaseURL()}${path}`, {
    method,
    headers,
    body: body === undefined ? undefined : JSON.stringify(body),
  });

  const text = await response.text();
  let payload: T | ApiError;
  try {
    payload = JSON.parse(text) as T | ApiError;
  } catch {
    throw new Error(`Request failed: ${response.status}`);
  }

  if (!response.ok) {
    const errorPayload = payload as ApiError;
    const message = errorPayload.error?.message || `Request failed: ${response.status}`;
    throw new Error(message);
  }

  return payload as T;
}

async function postAuthJSON<T>(path: string, body: unknown, token?: string): Promise<T> {
  return requestAuthJSON<T>('POST', path, body, token);
}

interface AuthResponse {
  data: AuthSession;
}

export async function login(account: string, password: string): Promise<AuthSession> {
  const payload = await postAuthJSON<AuthResponse>('/api/v1/auth/login', { account, password });
  return payload.data;
}

export async function register(username: string, email: string, password: string, confirmPassword: string, displayName: string): Promise<AuthSession> {
  const payload = await postAuthJSON<AuthResponse>('/api/v1/auth/register', { username, email, password, confirmPassword, displayName });
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

  const text = await response.text();
  let payload: { data: AuthUser } | ApiError;
  try {
    payload = JSON.parse(text) as { data: AuthUser } | ApiError;
  } catch {
    throw new Error(`Request failed: ${response.status}`);
  }

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

export async function updateProfile(token: string, displayName: string): Promise<AuthUser> {
  const payload = await requestAuthJSON<{ data: AuthUser }>(
    'PATCH',
    '/api/v1/auth/me',
    { displayName },
    token,
  );
  return payload.data;
}

interface ChangePasswordResult {
  updated: boolean;
  revokedSessions: number;
}

export async function changePassword(
  token: string,
  currentPassword: string,
  newPassword: string,
): Promise<ChangePasswordResult> {
  const payload = await postAuthJSON<{ data: ChangePasswordResult }>(
    '/api/v1/auth/password',
    { currentPassword, newPassword },
    token,
  );
  return payload.data;
}

interface RevokeSessionsResult {
  revokedSessions: number;
  keepCurrent: boolean;
}

export interface UserSession {
  accessToken: string;
  tokenPreview: string;
  isCurrent: boolean;
  ipAddress: string;
  userAgent: string;
  createdAt: string;
  lastActiveAt: string;
  expiresAt: string;
}

export async function listUserSessions(token: string): Promise<UserSession[]> {
  const response = await fetch(`${getClientAPIBaseURL()}/api/v1/auth/sessions`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  const text = await response.text();
  const payload = JSON.parse(text) as { data: UserSession[] } | ApiError;
  if (!response.ok) {
    throw new Error((payload as ApiError).error?.message || `Request failed: ${response.status}`);
  }
  return (payload as { data: UserSession[] }).data ?? [];
}

export async function revokeUserSession(token: string, accessToken: string): Promise<void> {
  const response = await fetch(`${getClientAPIBaseURL()}/api/v1/auth/sessions`, {
    method: 'DELETE',
    headers: {
      Authorization: `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ accessToken }),
  });
  if (!response.ok) {
    const text = await response.text();
    const payload = JSON.parse(text) as ApiError;
    throw new Error(payload.error?.message || `Request failed: ${response.status}`);
  }
}

export async function revokeOtherSessions(
  token: string,
  keepCurrent = true,
): Promise<RevokeSessionsResult> {
  const payload = await postAuthJSON<{ data: RevokeSessionsResult }>(
    '/api/v1/auth/sessions/revoke',
    { keepCurrent },
    token,
  );
  return payload.data;
}

import { getClientAPIBaseURL } from './auth';

interface ApiError {
  error?: {
    code?: string;
    message?: string;
  };
}

interface ApiResponse<T> {
  data: T;
  meta?: {
    page?: number;
    pageSize?: number;
    total?: number;
    totalPages?: number;
  };
}

async function adminRequest<T>(
  method: string,
  path: string,
  token: string,
  body?: unknown,
): Promise<ApiResponse<T>> {
  const headers: Record<string, string> = {
    Authorization: `Bearer ${token}`,
  };
  if (body !== undefined) {
    headers['Content-Type'] = 'application/json';
  }

  const response = await fetch(`${getClientAPIBaseURL()}${path}`, {
    method,
    headers,
    body: body === undefined ? undefined : JSON.stringify(body),
  });

  const text = await response.text();
  let payload: ApiResponse<T> | ApiError;
  try {
    payload = JSON.parse(text) as ApiResponse<T> | ApiError;
  } catch {
    throw new Error(`Invalid JSON response from server: ${response.status}`);
  }

  if (!response.ok) {
    const errorPayload = payload as ApiError;
    const message = errorPayload.error?.message || `Request failed: ${response.status}`;
    throw new Error(message);
  }

  return payload as ApiResponse<T>;
}

// ── Types ──

export interface AdminStats {
  toolCount: number;
  publishedToolCount: number;
  draftToolCount: number;
  categoryCount: number;
  userCount: number;
  activeUserCount: number;
  pendingSubmissionCount: number;
  submissionCount: number;
}

export interface AdminTool {
  id: string;
  slug: string;
  name: string;
  shortDescription: string;
  longDescription: string;
  categoryId: string;
  difficulty: string;
  icon: string;
  featured: boolean;
  status: string;
  websiteUrl: string;
  githubUrl?: string;
  viewCount: number;
  favoriteCount: number;
  publishedAt?: string;
  createdAt: string;
  updatedAt: string;
}

export interface AdminCategory {
  id: string;
  slug: string;
  name: string;
  description: string;
  icon: string;
  sortOrder: number;
  isVisible: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface AdminUser {
  id: string;
  email: string;
  displayName: string;
  role: string;
  isActive: boolean;
  lastLoginAt?: string;
  createdAt: string;
}

export interface AdminSubmission {
  id: string;
  type: string;
  submittedBy?: string;
  toolId?: string;
  submitterEmail?: string;
  payload: Record<string, unknown>;
  status: string;
  reviewerId?: string;
  reviewNote?: string;
  createdAt: string;
  reviewedAt?: string;
}

export interface AdminAuditLog {
  id: string;
  userId?: string;
  action: string;
  resourceType: string;
  resourceId?: string;
  beforeData?: Record<string, unknown>;
  afterData?: Record<string, unknown>;
  createdAt: string;
}

export interface PaginationMeta {
  page: number;
  pageSize: number;
  total: number;
  totalPages: number;
}

// ── Dashboard ──

export async function getAdminStats(token: string): Promise<AdminStats> {
  const res = await adminRequest<AdminStats>('GET', '/api/v1/admin/stats', token);
  return res.data;
}

// ── Tools ──

export interface ToolFilters {
  q?: string;
  status?: string;
  category?: string;
  difficulty?: string;
  page?: number;
  pageSize?: number;
  sort?: string;
}

export async function getAdminTools(
  token: string,
  filters?: ToolFilters,
): Promise<{ data: AdminTool[]; meta: PaginationMeta }> {
  const params = new URLSearchParams();
  if (filters?.q) params.set('q', filters.q);
  if (filters?.status) params.set('status', filters.status);
  if (filters?.category) params.set('category', filters.category);
  if (filters?.difficulty) params.set('difficulty', filters.difficulty);
  if (filters?.page) params.set('page', String(filters.page));
  if (filters?.pageSize) params.set('pageSize', String(filters.pageSize));
  if (filters?.sort) params.set('sort', filters.sort);
  const suffix = params.toString() ? `?${params.toString()}` : '';
  const res = await adminRequest<AdminTool[]>('GET', `/api/v1/admin/tools${suffix}`, token);
  return { data: res.data || [], meta: res.meta as PaginationMeta };
}

export async function createTool(
  token: string,
  data: {
    slug: string;
    name: string;
    shortDescription: string;
    longDescription: string;
    categoryId: string;
    difficulty: string;
    icon: string;
    featured: boolean;
    status: string;
    websiteUrl: string;
    githubUrl?: string;
    tags?: string[];
  },
): Promise<AdminTool> {
  const res = await adminRequest<AdminTool>('POST', '/api/v1/admin/tools', token, data);
  return res.data;
}

export async function updateTool(
  token: string,
  id: string,
  data: Partial<{
    slug: string;
    name: string;
    shortDescription: string;
    longDescription: string;
    categoryId: string;
    difficulty: string;
    icon: string;
    featured: boolean;
    status: string;
    websiteUrl: string;
    githubUrl: string;
    tags: string[];
  }>,
): Promise<AdminTool> {
  const res = await adminRequest<AdminTool>('PATCH', `/api/v1/admin/tools/${id}`, token, data);
  return res.data;
}

export async function archiveTool(token: string, id: string): Promise<void> {
  await adminRequest<{ archived: boolean }>('DELETE', `/api/v1/admin/tools/${id}`, token);
}

// ── Categories ──

export async function getAdminCategories(token: string): Promise<AdminCategory[]> {
  const res = await adminRequest<AdminCategory[]>('GET', '/api/v1/admin/categories', token);
  return res.data || [];
}

export async function createCategory(
  token: string,
  data: { slug: string; name: string; description: string; icon: string; sortOrder: number; isVisible: boolean },
): Promise<AdminCategory> {
  const res = await adminRequest<AdminCategory>('POST', '/api/v1/admin/categories', token, data);
  return res.data;
}

export async function updateCategory(
  token: string,
  id: string,
  data: Partial<{ slug: string; name: string; description: string; icon: string; sortOrder: number; isVisible: boolean }>,
): Promise<AdminCategory> {
  const res = await adminRequest<AdminCategory>('PATCH', `/api/v1/admin/categories/${id}`, token, data);
  return res.data;
}

export async function deleteCategory(token: string, id: string): Promise<void> {
  await adminRequest<AdminCategory>('DELETE', `/api/v1/admin/categories/${id}`, token);
}

// ── Users ──

export async function getAdminUsers(
  token: string,
  page = 1,
  pageSize = 20,
): Promise<{ data: AdminUser[]; meta: PaginationMeta }> {
  const res = await adminRequest<AdminUser[]>(
    'GET',
    `/api/v1/admin/users?page=${page}&pageSize=${pageSize}`,
    token,
  );
  return { data: res.data || [], meta: res.meta as PaginationMeta };
}

export async function updateUserRole(
  token: string,
  id: string,
  role: string,
): Promise<AdminUser> {
  const res = await adminRequest<AdminUser>('PATCH', `/api/v1/admin/users/${id}`, token, { role });
  return res.data;
}

export async function toggleUserActive(
  token: string,
  id: string,
  active: boolean,
): Promise<void> {
  await adminRequest<{ active: boolean }>('DELETE', `/api/v1/admin/users/${id}`, token, { active });
}

// ── Submissions ──

export async function getAdminSubmissions(
  token: string,
  status?: string,
  page = 1,
  pageSize = 20,
): Promise<{ data: AdminSubmission[]; meta: PaginationMeta }> {
  const params = new URLSearchParams();
  if (status) params.set('status', status);
  params.set('page', String(page));
  params.set('pageSize', String(pageSize));
  const res = await adminRequest<AdminSubmission[]>(
    'GET',
    `/api/v1/admin/submissions?${params.toString()}`,
    token,
  );
  return { data: res.data || [], meta: res.meta as PaginationMeta };
}

export async function reviewSubmission(
  token: string,
  id: string,
  status: string,
  note?: string,
): Promise<AdminSubmission> {
  const res = await adminRequest<AdminSubmission>(
    'PATCH',
    `/api/v1/admin/submissions/${id}`,
    token,
    { status, note: note || '' },
  );
  return res.data;
}

// ── Audit Logs ──

export async function getAdminAuditLogs(
  token: string,
  page = 1,
  pageSize = 20,
): Promise<{ data: AdminAuditLog[]; meta: PaginationMeta }> {
  const res = await adminRequest<AdminAuditLog[]>(
    'GET',
    `/api/v1/admin/audit-logs?page=${page}&pageSize=${pageSize}`,
    token,
  );
  return { data: res.data || [], meta: res.meta as PaginationMeta };
}

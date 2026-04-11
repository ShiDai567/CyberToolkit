import { Category, Difficulty, Tool } from '@/types/types';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080';

interface ApiResponse<T> {
  data: T;
  meta?: {
    page?: number;
    pageSize?: number;
    total?: number;
    totalPages?: number;
  };
}

interface ApiCategory {
  id: string;
  name: string;
  description: string;
  icon: string;
  toolCount?: number;
}

interface ApiToolCard {
  id: string;
  name: string;
  description: string;
  category: {
    id: string;
    name: string;
  };
  difficulty: Difficulty;
  icon: string;
  featured?: boolean;
  tags: string[];
  links: {
    website: string;
    github?: string;
  };
}

interface ApiToolDetail {
  id: string;
  name: string;
  description: string;
  longDescription: string;
  category: {
    id: string;
    name: string;
    description: string;
  };
  difficulty: Difficulty;
  icon: string;
  featured?: boolean;
  tags: string[];
  links: Array<{
    type: string;
    label: string;
    url: string;
  }>;
  relatedTools: Array<{
    id: string;
    name: string;
  }>;
}

interface HomePayload {
  stats: {
    toolCount: number;
    categoryCount: number;
    featuredCount: number;
  };
  featuredTools: ApiToolCard[];
  categories: ApiCategory[];
}

export interface HomeData {
  stats: {
    toolCount: number;
    categoryCount: number;
    featuredCount: number;
  };
  featuredTools: Tool[];
  categories: Category[];
}

export interface ToolDetailData extends Tool {
  categoryInfo: {
    id: string;
    name: string;
    description: string;
  };
}

async function fetchAPI<T>(path: string): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    next: { revalidate: 0 },
    cache: 'no-store',
  });

  if (!response.ok) {
    throw new Error(`API request failed: ${response.status} ${path}`);
  }

  const payload = (await response.json()) as ApiResponse<T>;
  return payload.data;
}

function normalizeToolCard(tool: ApiToolCard): Tool {
  return {
    id: tool.id,
    name: tool.name,
    description: tool.description,
    longDescription: tool.description,
    category: tool.category.id,
    tags: tool.tags,
    website: tool.links.website,
    github: tool.links.github,
    difficulty: tool.difficulty,
    icon: tool.icon,
    featured: tool.featured,
  };
}

function normalizeCategory(category: ApiCategory): Category {
  return {
    id: category.id,
    name: category.name,
    description: category.description,
    icon: category.icon,
    toolCount: category.toolCount ?? 0,
  };
}

export async function getHomeData(): Promise<HomeData> {
  const data = await fetchAPI<HomePayload>('/api/v1/home');

  return {
    stats: data.stats,
    featuredTools: data.featuredTools.map(normalizeToolCard),
    categories: data.categories.map(normalizeCategory),
  };
}

export async function getCategories(): Promise<Category[]> {
  const data = await fetchAPI<ApiCategory[]>('/api/v1/categories?includeCounts=true');
  return data.map(normalizeCategory);
}

export async function getTools(query?: {
  category?: string;
  difficulty?: string;
  q?: string;
  featured?: boolean;
  page?: number;
  pageSize?: number;
  sort?: string;
}): Promise<Tool[]> {
  const params = new URLSearchParams();
  if (query?.category) params.set('category', query.category);
  if (query?.difficulty) params.set('difficulty', query.difficulty);
  if (query?.q) params.set('q', query.q);
  if (typeof query?.featured === 'boolean') params.set('featured', String(query.featured));
  if (query?.page) params.set('page', String(query.page));
  if (query?.pageSize) params.set('pageSize', String(query.pageSize));
  if (query?.sort) params.set('sort', query.sort);

  const suffix = params.toString() ? `?${params.toString()}` : '';
  const data = await fetchAPI<ApiToolCard[]>(`/api/v1/tools${suffix}`);
  return data.map(normalizeToolCard);
}

export async function getToolById(id: string): Promise<ToolDetailData | null> {
  try {
    const data = await fetchAPI<ApiToolDetail>(`/api/v1/tools/${id}`);
    const github = data.links.find((link) => link.type === 'github')?.url;
    const website = data.links.find((link) => link.type === 'website')?.url ?? '';

    return {
      id: data.id,
      name: data.name,
      description: data.description,
      longDescription: data.longDescription,
      category: data.category.id,
      tags: data.tags,
      website,
      github,
      difficulty: data.difficulty,
      icon: data.icon,
      featured: data.featured,
      categoryInfo: {
        id: data.category.id,
        name: data.category.name,
        description: data.category.description,
      },
    };
  } catch {
    return null;
  }
}

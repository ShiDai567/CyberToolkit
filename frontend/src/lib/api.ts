import { Category, Difficulty, Tool } from '@/types/types';

const EMPTY_HOME_DATA: HomeData = {
  stats: {
    toolCount: 0,
    categoryCount: 0,
    featuredCount: 0,
  },
  featuredTools: [],
  categories: [],
};

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

function normalizeBaseURL(value: string): string {
  return value.trim().replace(/\/+$/, '');
}

function buildFallbackBaseURLs(baseURL: string): string[] {
  const candidates = new Set<string>([normalizeBaseURL(baseURL)]);

  try {
    const url = new URL(baseURL);
    const port = url.port || (url.protocol === 'https:' ? '443' : '80');

    if (url.hostname === 'localhost') {
      candidates.add(`${url.protocol}//127.0.0.1:${port}`);
    }

    if (url.hostname === '127.0.0.1') {
      candidates.add(`${url.protocol}//localhost:${port}`);
    }
  } catch {
    // Ignore malformed env values here and let fetch surface the real error.
  }

  return Array.from(candidates);
}

function getAPIBaseURLs(): string[] {
  const configuredBaseURL =
    process.env.API_BASE_URL ||
    process.env.NEXT_PUBLIC_API_BASE_URL ||
    'http://localhost:8080';

  return buildFallbackBaseURLs(configuredBaseURL);
}

async function fetchAPI<T>(path: string): Promise<T> {
  const baseURLs = getAPIBaseURLs();
  let lastError: Error | null = null;

  for (const baseURL of baseURLs) {
    try {
      const response = await fetch(`${baseURL}${path}`, {
        next: { revalidate: 0 },
        cache: 'no-store',
      });

      if (!response.ok) {
        throw new Error(`API request failed: ${response.status} ${path} via ${baseURL}`);
      }

      const payload = (await response.json()) as ApiResponse<T>;
      return payload.data;
    } catch (error) {
      lastError = error instanceof Error ? error : new Error(String(error));
    }
  }

  throw new Error(
    `Unable to reach API for ${path}. Tried: ${baseURLs.join(', ')}. Last error: ${lastError?.message || 'Unknown error'}`,
  );
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
  try {
    const data = await fetchAPI<HomePayload>('/api/v1/home');

    return {
      stats: data.stats,
      featuredTools: data.featuredTools.map(normalizeToolCard),
      categories: data.categories.map(normalizeCategory),
    };
  } catch (error) {
    console.error('Failed to load home data:', error);
    return EMPTY_HOME_DATA;
  }
}

export async function getCategories(): Promise<Category[]> {
  try {
    const data = await fetchAPI<ApiCategory[]>('/api/v1/categories?includeCounts=true');
    return data.map(normalizeCategory);
  } catch (error) {
    console.error('Failed to load categories:', error);
    return [];
  }
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
  try {
    const data = await fetchAPI<ApiToolCard[]>(`/api/v1/tools${suffix}`);
    return data.map(normalizeToolCard);
  } catch (error) {
    console.error('Failed to load tools:', error);
    return [];
  }
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

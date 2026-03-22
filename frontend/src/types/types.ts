export type Difficulty = 'beginner' | 'intermediate' | 'advanced' | 'expert';

export interface Tool {
  id: string;
  name: string;
  description: string;
  longDescription: string;
  category: string;
  tags: string[];
  website: string;
  github?: string;
  difficulty: Difficulty;
  icon: string;
  featured?: boolean;
}

export interface Category {
  id: string;
  name: string;
  description: string;
  icon: string;
  toolCount: number;
}

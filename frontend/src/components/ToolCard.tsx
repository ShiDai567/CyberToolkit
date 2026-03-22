import Link from 'next/link';
import {
  ExternalLink,
  Github,
  Radar,
  ShieldAlert,
  Crosshair,
  KeyRound,
  Wifi,
  Globe,
  Microscope,
  Eye,
  Zap,
  ScanLine,
  Search,
  Atom,
  Bug,
  Database,
  Cpu,
  Activity,
  FolderSearch,
  BrainCircuit,
  Network,
  Wheat,
} from 'lucide-react';
import { Tool } from '@/types/types';
import styles from './ToolCard.module.css';

const iconMap: Record<string, React.ElementType> = {
  Radar,
  ShieldAlert,
  Crosshair,
  KeyRound,
  Wifi,
  Globe,
  Microscope,
  Eye,
  Zap,
  ScanLine,
  Search,
  Atom,
  Bug,
  Database,
  Cpu,
  Activity,
  FolderSearch,
  BrainCircuit,
  Network,
  Wheat,
};

const difficultyLabels: Record<string, string> = {
  beginner: '入门',
  intermediate: '中级',
  advanced: '高级',
  expert: '专家',
};

export function ToolCard({ tool }: { tool: Tool }) {
  const Icon = iconMap[tool.icon] || Radar;

  return (
    <div className={styles.card}>
      <div className={styles.cardHeader}>
        <div className={styles.iconWrap}>
          <Icon size={22} />
        </div>
        <div className={styles.headerInfo}>
          <Link href={`/tools/${tool.id}`} className={styles.name}>
            {tool.name}
          </Link>
          <span className={`${styles.difficulty} difficulty-${tool.difficulty}`}>
            {difficultyLabels[tool.difficulty]}
          </span>
        </div>
      </div>

      <p className={styles.description}>{tool.description}</p>

      <div className={styles.tags}>
        {tool.tags.slice(0, 3).map((tag) => (
          <span key={tag} className={styles.tag}>
            {tag}
          </span>
        ))}
      </div>

      <div className={styles.links}>
        <a
          href={tool.website}
          target="_blank"
          rel="noopener noreferrer"
          className={styles.linkBtn}
        >
          <ExternalLink size={14} />
          <span>官网</span>
        </a>
        {tool.github && (
          <a
            href={tool.github}
            target="_blank"
            rel="noopener noreferrer"
            className={styles.linkBtn}
          >
            <Github size={14} />
            <span>源码</span>
          </a>
        )}
      </div>

      <div className={styles.glow} />
    </div>
  );
}

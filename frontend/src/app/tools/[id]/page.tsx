import { notFound } from 'next/navigation';
import Link from 'next/link';
import {
  ArrowLeft,
  ExternalLink,
  Github,
  Terminal,
  Tag,
  BarChart3,
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
import { ToolCard } from '@/components/ToolCard';
import { getToolById, getTools } from '@/lib/api';
import styles from './page.module.css';

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

const difficultyConfig: Record<string, { label: string; color: string }> = {
  beginner: { label: '入门', color: '#22C55E' },
  intermediate: { label: '中级', color: '#F59E0B' },
  advanced: { label: '高级', color: '#EF4444' },
  expert: { label: '专家', color: '#A855F7' },
};

export async function generateStaticParams() {
  return [];
}

export async function generateMetadata({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const tool = await getToolById(id);

  if (!tool) {
    return { title: '未找到工具' };
  }

  return {
    title: `${tool.name} - CyberToolkit`,
    description: tool.description,
  };
}

export default async function ToolDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const tool = await getToolById(id);

  if (!tool) {
    notFound();
  }

  const Icon = iconMap[tool.icon] || Radar;
  const category = tool.categoryInfo;
  const relatedTools = (await getTools({ category: tool.category, page: 1, pageSize: 10 }))
    .filter((t) => t.id !== tool.id)
    .slice(0, 3);
  const diff = difficultyConfig[tool.difficulty];

  return (
    <div className={styles.page}>
      <div className={styles.container}>
        <Link href="/tools" className={styles.backLink}>
          <ArrowLeft size={16} />
          <span>返回工具列表</span>
        </Link>

        <div className={styles.hero}>
          <div className={styles.heroIcon}>
            <Icon size={40} />
          </div>
          <div className={styles.heroInfo}>
            <div className={styles.heroMeta}>
              <span className={styles.categoryBadge}>{category.name}</span>
              <span
                className={styles.difficultyBadge}
                style={{ color: diff.color, borderColor: `${diff.color}40` }}
              >
                {diff.label}
              </span>
            </div>
            <h1 className={styles.heroTitle}>{tool.name}</h1>
            <p className={styles.heroDesc}>{tool.description}</p>
          </div>
        </div>

        <div className={styles.contentGrid}>
          <div className={styles.mainContent}>
            <div className={styles.section}>
              <h2 className={styles.sectionTitle}>
                <Terminal size={16} /> 概述
              </h2>
              <p className={styles.bodyText}>{tool.longDescription}</p>
            </div>

            <div className={styles.section}>
              <h2 className={styles.sectionTitle}>
                <Tag size={16} /> 标签
              </h2>
              <div className={styles.tagsList}>
                {tool.tags.map((tag) => (
                  <span key={tag} className={styles.tag}>
                    {tag}
                  </span>
                ))}
              </div>
            </div>
          </div>

          <div className={styles.sidebar}>
            <div className={styles.infoCard}>
              <h3 className={styles.infoTitle}>快速信息</h3>

              <div className={styles.infoRow}>
                <span className={styles.infoLabel}>分类</span>
                <span className={styles.infoValue}>{category.name}</span>
              </div>

              <div className={styles.infoRow}>
                <span className={styles.infoLabel}>难度</span>
                <span className={styles.infoValue} style={{ color: diff.color }}>
                  <BarChart3 size={14} />
                  {diff.label}
                </span>
              </div>

              <div className={styles.infoLinks}>
                <a
                  href={tool.website}
                  target="_blank"
                  rel="noopener noreferrer"
                  className={styles.primaryLink}
                >
                  <ExternalLink size={16} />
                  <span>官方网站</span>
                </a>
                {tool.github && (
                  <a
                    href={tool.github}
                    target="_blank"
                    rel="noopener noreferrer"
                    className={styles.secondaryLink}
                  >
                    <Github size={16} />
                    <span>源码仓库</span>
                  </a>
                )}
              </div>
            </div>
          </div>
        </div>

        {relatedTools.length > 0 && (
          <div className={styles.relatedSection}>
            <h2 className={styles.relatedTitle}>
              相关<span className="neon-text">工具</span>
            </h2>
            <div className={styles.relatedGrid}>
              {relatedTools.map((t) => (
                <ToolCard key={t.id} tool={t} />
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

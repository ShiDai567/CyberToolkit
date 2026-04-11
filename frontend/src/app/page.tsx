import { HeroSection } from '@/components/HeroSection';
import { StatsBar } from '@/components/StatsBar';
import { ToolCard } from '@/components/ToolCard';
import { CategoryCard } from '@/components/CategoryCard';
import { getHomeData } from '@/lib/api';
import { ChevronRight, Terminal } from 'lucide-react';
import Link from 'next/link';
import styles from './page.module.css';

export default async function HomePage() {
  const { stats, featuredTools, categories } = await getHomeData();

  return (
    <>
      <HeroSection />
      <StatsBar
        toolCount={stats.toolCount}
        categoryCount={stats.categoryCount}
        featuredCount={stats.featuredCount}
      />

      <section className={styles.section}>
        <div className={styles.container}>
          <div className={styles.sectionHeader}>
            <div>
              <span className={styles.sectionTag}>
                <Terminal size={12} /> 精选推荐
              </span>
              <h2 className={styles.sectionTitle}>
                精选安全<span className="neon-text">工具</span>
              </h2>
              <p className={styles.sectionDesc}>汇总安全研究与实战中常用的高质量工具。</p>
            </div>
            <Link href="/tools" className={styles.viewAll}>
              查看全部
              <ChevronRight size={16} />
            </Link>
          </div>

          <div className={styles.toolsGrid}>
            {featuredTools.map((tool, i) => (
              <div
                key={tool.id}
                className={`animate-fadeInUp stagger-${i + 1}`}
                style={{ opacity: 0 }}
              >
                <ToolCard tool={tool} />
              </div>
            ))}
          </div>
        </div>
      </section>

      <section className={styles.section} id="categories">
        <div className={styles.container}>
          <div className={styles.sectionHeader}>
            <div>
              <span className={styles.sectionTag}>
                <Terminal size={12} /> 分类浏览
              </span>
              <h2 className={styles.sectionTitle}>
                按<span className="neon-text">分类</span>浏览
              </h2>
              <p className={styles.sectionDesc}>按专业方向整理，快速定位你需要的工具。</p>
            </div>
          </div>

          <div className={styles.categoriesGrid}>
            {categories.map((category, i) => (
              <div
                key={category.id}
                className={`animate-fadeInUp stagger-${i + 1}`}
                style={{ opacity: 0 }}
              >
                <CategoryCard category={category} />
              </div>
            ))}
          </div>
        </div>
      </section>

      <section className={styles.ctaSection}>
        <div className={styles.container}>
          <div className={styles.ctaCard}>
            <h2 className={styles.ctaTitle}>
              准备好扩充你的<span className="neon-text">安全工具箱</span>了吗？
            </h2>
            <p className={styles.ctaDesc}>
              从网络扫描到数字取证，把常用能力集中到一个清晰的入口里。
            </p>
            <Link href="/tools" className={styles.ctaButton}>
              探索所有工具
              <ChevronRight size={18} />
            </Link>
          </div>
        </div>
      </section>
    </>
  );
}

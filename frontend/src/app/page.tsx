import { HeroSection } from '@/components/HeroSection';
import { StatsBar } from '@/components/StatsBar';
import { ToolCard } from '@/components/ToolCard';
import { CategoryCard } from '@/components/CategoryCard';
import { featuredTools, categories } from '@/data/tools';
import { ChevronRight, Terminal } from 'lucide-react';
import Link from 'next/link';
import styles from './page.module.css';

export default function HomePage() {
  return (
    <>
      <HeroSection />
      <StatsBar />

      {/* Featured Tools Section */}
      <section className={styles.section}>
        <div className={styles.container}>
          <div className={styles.sectionHeader}>
            <div>
              <span className={styles.sectionTag}>
                <Terminal size={12} /> 精选推荐
              </span>
              <h2 className={styles.sectionTitle}>
              精选安全 <span className="neon-text">工具</span>
            </h2>
              <p className={styles.sectionDesc}>
                全球安全专家信赖的精选工具
              </p>
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

      {/* Categories Section */}
      <section className={styles.section} id="categories">
        <div className={styles.container}>
          <div className={styles.sectionHeader}>
            <div>
              <span className={styles.sectionTag}>
                <Terminal size={12} /> 分类浏览
              </span>
              <h2 className={styles.sectionTitle}>
              按 <span className="neon-text">分类</span> 浏览
            </h2>
              <p className={styles.sectionDesc}>
                按专业领域分类，快速查找所需工具
              </p>
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

      {/* CTA Section */}
      <section className={styles.ctaSection}>
        <div className={styles.container}>
          <div className={styles.ctaCard}>
            <h2 className={styles.ctaTitle}>
              准备好提升你的 <span className="neon-text">安全技能</span> 了吗？
            </h2>
            <p className={styles.ctaDesc}>
              探索我们完整的网络安全工具集合，找到最适合你的安全武器库。
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

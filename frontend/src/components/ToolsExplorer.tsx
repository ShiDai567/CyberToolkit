'use client';

import { useMemo, useState } from 'react';
import { ToolCard } from '@/components/ToolCard';
import { SearchBar } from '@/components/SearchBar';
import { Category, Difficulty, Tool } from '@/types/types';
import { Terminal, Filter } from 'lucide-react';
import styles from '@/app/tools/page.module.css';

interface ToolsExplorerProps {
  tools: Tool[];
  categories: Category[];
  initialCategory?: string;
}

export function ToolsExplorer({
  tools,
  categories,
  initialCategory = '',
}: ToolsExplorerProps) {
  const [search, setSearch] = useState('');
  const [selectedCategory, setSelectedCategory] = useState(initialCategory);
  const [selectedDifficulty, setSelectedDifficulty] = useState<Difficulty | ''>('');

  const filteredTools = useMemo(() => {
    return tools.filter((tool) => {
      const matchesSearch =
        !search ||
        tool.name.toLowerCase().includes(search.toLowerCase()) ||
        tool.description.toLowerCase().includes(search.toLowerCase()) ||
        tool.tags.some((tag) => tag.toLowerCase().includes(search.toLowerCase()));

      const matchesCategory = !selectedCategory || tool.category === selectedCategory;
      const matchesDifficulty = !selectedDifficulty || tool.difficulty === selectedDifficulty;

      return matchesSearch && matchesCategory && matchesDifficulty;
    });
  }, [search, selectedCategory, selectedDifficulty, tools]);

  const difficulties: { value: Difficulty | ''; label: string }[] = [
    { value: '', label: '全部等级' },
    { value: 'beginner', label: '入门' },
    { value: 'intermediate', label: '中级' },
    { value: 'advanced', label: '高级' },
    { value: 'expert', label: '专家' },
  ];

  return (
    <div className={styles.page}>
      <div className={styles.container}>
        <div className={styles.header}>
          <span className={styles.tag}>
            <Terminal size={12} /> 工具库
          </span>
          <h1 className={styles.title}>
            全部<span className="neon-text">工具</span>
          </h1>
          <p className={styles.desc}>浏览、搜索并按分类筛选整个工具集合。</p>
        </div>

        <div className={styles.filters}>
          <SearchBar value={search} onChange={setSearch} placeholder="搜索工具、标签、分类..." />

          <div className={styles.filterRow}>
            <div className={styles.filterGroup}>
              <Filter size={14} className={styles.filterIcon} />
              <span className={styles.filterLabel}>分类：</span>
              <div className={styles.filterBtns}>
                <button
                  className={`${styles.filterBtn} ${selectedCategory === '' ? styles.filterBtnActive : ''}`}
                  onClick={() => setSelectedCategory('')}
                >
                  全部
                </button>
                {categories.map((cat) => (
                  <button
                    key={cat.id}
                    className={`${styles.filterBtn} ${selectedCategory === cat.id ? styles.filterBtnActive : ''}`}
                    onClick={() => setSelectedCategory(cat.id === selectedCategory ? '' : cat.id)}
                  >
                    {cat.name}
                  </button>
                ))}
              </div>
            </div>

            <div className={styles.filterGroup}>
              <span className={styles.filterLabel}>等级：</span>
              <div className={styles.filterBtns}>
                {difficulties.map((d) => (
                  <button
                    key={d.value}
                    className={`${styles.filterBtn} ${selectedDifficulty === d.value ? styles.filterBtnActive : ''}`}
                    onClick={() =>
                      setSelectedDifficulty(d.value === selectedDifficulty ? '' : (d.value as Difficulty | ''))
                    }
                  >
                    {d.label}
                  </button>
                ))}
              </div>
            </div>
          </div>
        </div>

        <div className={styles.results}>
          <span className={styles.resultCount}>找到 {filteredTools.length} 个工具</span>
        </div>

        {filteredTools.length > 0 ? (
          <div className={styles.grid}>
            {filteredTools.map((tool) => (
              <ToolCard key={tool.id} tool={tool} />
            ))}
          </div>
        ) : (
          <div className={styles.empty}>
            <p className={styles.emptyText}>没有找到符合当前条件的工具。</p>
            <button
              className={styles.resetBtn}
              onClick={() => {
                setSearch('');
                setSelectedCategory('');
                setSelectedDifficulty('');
              }}
            >
              重置筛选
            </button>
          </div>
        )}
      </div>
    </div>
  );
}

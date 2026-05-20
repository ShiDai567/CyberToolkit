'use client';

import { FormEvent, useCallback, useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import {
  ArrowLeft,
  AtSign,
  BookOpen,
  Code,
  Globe,
  Layers,
  Send,
  Tag,
  Terminal,
  User,
} from 'lucide-react';
import { toast } from 'sonner';
import { useAuth } from '@/components/AuthProvider';
import { submitTool, type SubmissionPayload } from '@/lib/auth';
import { getCategories } from '@/lib/api';
import type { Category, Difficulty } from '@/types/types';
import styles from './page.module.css';

const DIFFICULTY_OPTIONS: { value: Difficulty; label: string }[] = [
  { value: 'beginner', label: '初级' },
  { value: 'intermediate', label: '中级' },
  { value: 'advanced', label: '高级' },
  { value: 'expert', label: '专家' },
];

export default function SubmitToolPage() {
  const router = useRouter();
  const { user, token, isLoading, isAuthenticated } = useAuth();

  const [categories, setCategories] = useState<Category[]>([]);
  const [submitting, setSubmitting] = useState(false);

  const [name, setName] = useState('');
  const [website, setWebsite] = useState('');
  const [github, setGithub] = useState('');
  const [category, setCategory] = useState('');
  const [difficulty, setDifficulty] = useState<Difficulty>('beginner');
  const [shortDescription, setShortDescription] = useState('');
  const [longDescription, setLongDescription] = useState('');
  const [tagsInput, setTagsInput] = useState('');

  useEffect(() => {
    getCategories().then(setCategories);
  }, []);

  const handleSubmit = useCallback(
    async (e: FormEvent) => {
      e.preventDefault();
      if (!token) return;

      const trimmedName = name.trim();
      const trimmedWebsite = website.trim();
      const trimmedShort = shortDescription.trim();
      const trimmedLong = longDescription.trim();

      if (!trimmedName) {
        toast.error('请输入工具名称');
        return;
      }
      if (!trimmedWebsite) {
        toast.error('请输入工具网站链接');
        return;
      }
      if (!category) {
        toast.error('请选择工具分类');
        return;
      }
      if (!trimmedShort) {
        toast.error('请输入简短描述');
        return;
      }
      if (trimmedShort.length > 280) {
        toast.error('简短描述不能超过 280 个字符');
        return;
      }
      if (!trimmedLong) {
        toast.error('请输入详细描述');
        return;
      }

      const tags = tagsInput
        .split(',')
        .map((t) => t.trim())
        .filter(Boolean);

      const data: SubmissionPayload = {
        name: trimmedName,
        website: trimmedWebsite,
        github: github.trim() || undefined,
        category,
        difficulty,
        shortDescription: trimmedShort,
        longDescription: trimmedLong,
        tags,
      };

      setSubmitting(true);
      try {
        await submitTool(token, data);
        toast.success('投稿成功！管理员审核后将会发布。');
        router.push('/account');
      } catch (err) {
        const text = err instanceof Error ? err.message : '投稿失败，请稍后重试';
        toast.error(text);
      } finally {
        setSubmitting(false);
      }
    },
    [token, name, website, github, category, difficulty, shortDescription, longDescription, tagsInput, router],
  );

  if (isLoading) {
    return (
      <div className={styles.empty}>
        <div className={styles.emptyCard}>
          <div className={styles.emptyTitle}>正在验证登录状态…</div>
          <p className={styles.emptyText}>请稍候，正在与认证终端建立链接。</p>
        </div>
      </div>
    );
  }

  if (!isAuthenticated || !user || !token) {
    return (
      <div className={styles.empty}>
        <div className={styles.emptyCard}>
          <div className={styles.emptyTitle}>未登录</div>
          <p className={styles.emptyText}>需要先通过身份认证才能提交工具投稿。</p>
          <Link href="/login?redirect=/tools/submit" className={styles.loginLink}>
            前往登录
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className={styles.page}>
      <div className={styles.bgGlow} aria-hidden />
      <div className={styles.bgGrid} aria-hidden />

      <div className={styles.container}>
        {/* Header */}
        <section className={styles.hero}>
          <span className={styles.heroScan} aria-hidden />

          <div className={styles.avatar}>
            <Send size={36} />
          </div>

          <div className={styles.heroMain}>
            <span className={styles.heroLabel}>
              <span className={styles.heroLabelPrompt}>{'>'}</span> SUBMIT_TOOL // NEW ENTRY
            </span>
            <h1 className={styles.heroName}>
              投稿新工具<span className={styles.heroNameAccent}>_</span>
            </h1>
            <div className={styles.heroMeta}>
              <span className={styles.heroChip}>
                <User size={12} className={styles.heroChipAccent} />
                {user.displayName}
              </span>
              <span className={styles.heroChip}>
                <AtSign size={12} className={styles.heroChipAccent} />
                {user.email}
              </span>
            </div>
          </div>

          <div className={styles.heroSide}>
            <Link href="/tools" className={styles.backLink}>
              <ArrowLeft size={14} /> 返回工具库
            </Link>
          </div>
        </section>

        {/* Form */}
        <section className={styles.panel}>
          <header className={styles.panelHeader}>
            <Terminal size={14} className={styles.panelTitleAccent} />
            TOOL_SUBMISSION // <span className={styles.panelTitleAccent}>FORM</span>
          </header>
          <div className={styles.panelBody}>
            <form className={styles.form} onSubmit={handleSubmit} noValidate>
              {/* Name */}
              <div className={styles.field}>
                <label className={styles.label} htmlFor="name">
                  <span className={styles.labelPrompt}>{'>'}</span> TOOL_NAME
                </label>
                <div className={styles.inputWrapper}>
                  <Terminal className={styles.inputIcon} />
                  <input
                    id="name"
                    className={styles.input}
                    type="text"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    placeholder="例如：Nmap、Burp Suite"
                    maxLength={150}
                    disabled={submitting}
                    required
                  />
                </div>
              </div>

              {/* Website */}
              <div className={styles.field}>
                <label className={styles.label} htmlFor="website">
                  <span className={styles.labelPrompt}>{'>'}</span> WEBSITE_URL
                </label>
                <div className={styles.inputWrapper}>
                  <Globe className={styles.inputIcon} />
                  <input
                    id="website"
                    className={styles.input}
                    type="url"
                    value={website}
                    onChange={(e) => setWebsite(e.target.value)}
                    placeholder="https://example.com"
                    disabled={submitting}
                    required
                  />
                </div>
                <span className={styles.hint}>工具的官方网站或文档链接。</span>
              </div>

              {/* GitHub (optional) */}
              <div className={styles.field}>
                <label className={styles.label} htmlFor="github">
                  <span className={styles.labelPrompt}>{'>'}</span> GITHUB_URL{' '}
                  <span style={{ opacity: 0.5 }}>(OPTIONAL)</span>
                </label>
                <div className={styles.inputWrapper}>
                  <Code className={styles.inputIcon} />
                  <input
                    id="github"
                    className={styles.input}
                    type="url"
                    value={github}
                    onChange={(e) => setGithub(e.target.value)}
                    placeholder="https://github.com/user/repo"
                    disabled={submitting}
                  />
                </div>
                <span className={styles.hint}>如果工具有开源仓库，请填写 GitHub 链接。</span>
              </div>

              {/* Category */}
              <div className={styles.field}>
                <label className={styles.label} htmlFor="category">
                  <span className={styles.labelPrompt}>{'>'}</span> CATEGORY
                </label>
                <div className={styles.inputWrapper}>
                  <Layers className={styles.inputIcon} />
                  <select
                    id="category"
                    className={styles.input}
                    value={category}
                    onChange={(e) => setCategory(e.target.value)}
                    disabled={submitting}
                    required
                    style={{ paddingLeft: 38 }}
                  >
                    <option value="">请选择分类</option>
                    {categories.map((cat) => (
                      <option key={cat.id} value={cat.id}>
                        {cat.name}
                      </option>
                    ))}
                  </select>
                </div>
              </div>

              {/* Difficulty */}
              <div className={styles.field}>
                <label className={styles.label} htmlFor="difficulty">
                  <span className={styles.labelPrompt}>{'>'}</span> DIFFICULTY
                </label>
                <div className={styles.inputWrapper}>
                  <BookOpen className={styles.inputIcon} />
                  <select
                    id="difficulty"
                    className={styles.input}
                    value={difficulty}
                    onChange={(e) => setDifficulty(e.target.value as Difficulty)}
                    disabled={submitting}
                    style={{ paddingLeft: 38 }}
                  >
                    {DIFFICULTY_OPTIONS.map((opt) => (
                      <option key={opt.value} value={opt.value}>
                        {opt.label}
                      </option>
                    ))}
                  </select>
                </div>
              </div>

              {/* Short Description */}
              <div className={styles.field}>
                <label className={styles.label} htmlFor="shortDesc">
                  <span className={styles.labelPrompt}>{'>'}</span> SHORT_DESCRIPTION
                </label>
                <div className={styles.textareaWrapper}>
                  <textarea
                    id="shortDesc"
                    className={`${styles.input} ${styles.textarea}`}
                    value={shortDescription}
                    onChange={(e) => setShortDescription(e.target.value)}
                    placeholder="一句话描述工具的主要功能"
                    maxLength={280}
                    rows={2}
                    disabled={submitting}
                    required
                    style={{ paddingLeft: 12 }}
                  />
                </div>
                <span className={styles.hint}>{shortDescription.length} / 280 字符</span>
              </div>

              {/* Long Description */}
              <div className={styles.field}>
                <label className={styles.label} htmlFor="longDesc">
                  <span className={styles.labelPrompt}>{'>'}</span> LONG_DESCRIPTION
                </label>
                <div className={styles.textareaWrapper}>
                  <textarea
                    id="longDesc"
                    className={`${styles.input} ${styles.textarea}`}
                    value={longDescription}
                    onChange={(e) => setLongDescription(e.target.value)}
                    placeholder="详细描述工具的功能、使用场景和特点"
                    rows={4}
                    disabled={submitting}
                    required
                    style={{ paddingLeft: 12 }}
                  />
                </div>
              </div>

              {/* Tags */}
              <div className={styles.field}>
                <label className={styles.label} htmlFor="tags">
                  <span className={styles.labelPrompt}>{'>'}</span> TAGS{' '}
                  <span style={{ opacity: 0.5 }}>(OPTIONAL)</span>
                </label>
                <div className={styles.inputWrapper}>
                  <Tag className={styles.inputIcon} />
                  <input
                    id="tags"
                    className={styles.input}
                    type="text"
                    value={tagsInput}
                    onChange={(e) => setTagsInput(e.target.value)}
                    placeholder="port-scan, automation, web-security（逗号分隔）"
                    disabled={submitting}
                  />
                </div>
                <span className={styles.hint}>用逗号分隔多个标签，有助于用户搜索和分类工具。</span>
              </div>

              {/* Actions */}
              <div className={styles.actions}>
                <button type="submit" className={styles.btnPrimary} disabled={submitting}>
                  {submitting ? (
                    <>
                      <span className={styles.spinner} />
                      提交中…
                    </>
                  ) : (
                    <>
                      <Send size={14} />
                      提交投稿
                    </>
                  )}
                </button>
                <button
                  type="button"
                  className={styles.btnGhost}
                  onClick={() => router.back()}
                  disabled={submitting}
                >
                  取消
                </button>
              </div>
            </form>
          </div>
        </section>
      </div>
    </div>
  );
}

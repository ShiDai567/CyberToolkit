'use client';

import { useCallback, useEffect, useMemo, useState } from 'react';
import { useAuth } from '@/components/AuthProvider';
import {
  AdminTool,
  AdminCategory,
  PaginationMeta,
  getAdminTools,
  getAdminCategories,
  createTool,
  updateTool,
  archiveTool,
} from '@/lib/admin';
import styles from './page.module.css';

/* ─── Constants ─── */

const STATUS_TABS = [
  { key: '', label: '全部' },
  { key: 'draft', label: '草稿' },
  { key: 'published', label: '已发布' },
  { key: 'archived', label: '已归档' },
] as const;

const DIFFICULTY_OPTIONS = [
  { value: 'beginner', label: '入门' },
  { value: 'intermediate', label: '中级' },
  { value: 'advanced', label: '高级' },
  { value: 'expert', label: '专家' },
] as const;

const STATUS_OPTIONS = [
  { value: 'draft', label: '草稿' },
  { value: 'published', label: '已发布' },
  { value: 'archived', label: '已归档' },
] as const;

const PAGE_SIZE = 15;

/* ─── Helpers ─── */

const statusBadgeClass: Record<string, string> = {
  draft: styles.badgeDraft,
  published: styles.badgePublished,
  archived: styles.badgeArchived,
};

const difficultyBadgeClass: Record<string, string> = {
  beginner: styles.badgeBeginner,
  intermediate: styles.badgeIntermediate,
  advanced: styles.badgeAdvanced,
  expert: styles.badgeExpert,
};

/* ─── Form Data Shape ─── */

interface ToolFormData {
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
  tags: string;
}

const emptyForm: ToolFormData = {
  slug: '',
  name: '',
  shortDescription: '',
  longDescription: '',
  categoryId: '',
  difficulty: 'beginner',
  icon: '',
  featured: false,
  status: 'draft',
  websiteUrl: '',
  githubUrl: '',
  tags: '',
};

/* ─── Page Component ─── */

export default function AdminToolsPage() {
  const { token } = useAuth();

  // List state
  const [tools, setTools] = useState<AdminTool[]>([]);
  const [meta, setMeta] = useState<PaginationMeta | null>(null);
  const [categories, setCategories] = useState<AdminCategory[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  // Filters
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [page, setPage] = useState(1);

  // Modal state
  const [modalOpen, setModalOpen] = useState(false);
  const [editingTool, setEditingTool] = useState<AdminTool | null>(null);
  const [form, setForm] = useState<ToolFormData>(emptyForm);
  const [saving, setSaving] = useState(false);
  const [modalError, setModalError] = useState('');
  const [modalSuccess, setModalSuccess] = useState('');

  // Archive confirmation
  const [archiveTarget, setArchiveTarget] = useState<AdminTool | null>(null);
  const [archiving, setArchiving] = useState(false);

  // Page-level success message
  const [pageSuccess, setPageSuccess] = useState('');

  // Category lookup map
  const categoryMap = useMemo(() => {
    const map: Record<string, string> = {};
    for (const cat of categories) {
      map[cat.id] = cat.name;
    }
    return map;
  }, [categories]);

  /* ─── Fetch data ─── */

  const fetchTools = useCallback(async () => {
    if (!token) return;
    setLoading(true);
    setError('');
    try {
      const res = await getAdminTools(token, {
        q: search || undefined,
        status: statusFilter || undefined,
        page,
        pageSize: PAGE_SIZE,
      });
      setTools(res.data);
      setMeta(res.meta);
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载工具列表失败');
    } finally {
      setLoading(false);
    }
  }, [token, search, statusFilter, page]);

  const fetchCategories = useCallback(async () => {
    if (!token) return;
    try {
      const cats = await getAdminCategories(token);
      setCategories(cats);
    } catch {
      // silent – categories are supplementary
    }
  }, [token]);

  useEffect(() => {
    fetchTools();
  }, [fetchTools]);

  useEffect(() => {
    fetchCategories();
  }, [fetchCategories]);

  // Debounced search: reset page on search change
  useEffect(() => {
    setPage(1);
  }, [search, statusFilter]);

  // Auto-dismiss page success
  useEffect(() => {
    if (!pageSuccess) return;
    const t = setTimeout(() => setPageSuccess(''), 3000);
    return () => clearTimeout(t);
  }, [pageSuccess]);

  /* ─── Modal open / close ─── */

  function openCreate() {
    setEditingTool(null);
    setForm(emptyForm);
    setModalError('');
    setModalSuccess('');
    setModalOpen(true);
  }

  function openEdit(tool: AdminTool) {
    setEditingTool(tool);
    setForm({
      slug: tool.slug,
      name: tool.name,
      shortDescription: tool.shortDescription,
      longDescription: tool.longDescription,
      categoryId: tool.categoryId,
      difficulty: tool.difficulty,
      icon: tool.icon,
      featured: tool.featured,
      status: tool.status,
      websiteUrl: tool.websiteUrl,
      githubUrl: tool.githubUrl || '',
      tags: '', // API doesn't return tags on list, user can update if needed
    });
    setModalError('');
    setModalSuccess('');
    setModalOpen(true);
  }

  function closeModal() {
    setModalOpen(false);
    setEditingTool(null);
    setModalError('');
    setModalSuccess('');
  }

  /* ─── Form helpers ─── */

  function setField<K extends keyof ToolFormData>(key: K, value: ToolFormData[K]) {
    setForm((prev) => ({ ...prev, [key]: value }));
  }

  /* ─── Save (Create / Update) ─── */

  async function handleSave() {
    if (!token) return;
    setSaving(true);
    setModalError('');
    setModalSuccess('');

    const payload = {
      slug: form.slug.trim(),
      name: form.name.trim(),
      shortDescription: form.shortDescription.trim(),
      longDescription: form.longDescription.trim(),
      categoryId: form.categoryId,
      difficulty: form.difficulty,
      icon: form.icon.trim(),
      featured: form.featured,
      status: form.status,
      websiteUrl: form.websiteUrl.trim(),
      githubUrl: form.githubUrl.trim() || undefined,
      tags: form.tags
        .split(',')
        .map((t) => t.trim())
        .filter(Boolean),
    };

    if (!payload.name || !payload.slug) {
      setModalError('名称和 Slug 为必填项');
      setSaving(false);
      return;
    }

    try {
      if (editingTool) {
        await updateTool(token, editingTool.id, payload);
        setModalSuccess('工具更新成功');
      } else {
        await createTool(token, payload as Parameters<typeof createTool>[1]);
        setModalSuccess('工具创建成功');
      }
      setTimeout(() => {
        closeModal();
        setPageSuccess(editingTool ? '工具已更新' : '工具已创建');
        fetchTools();
      }, 600);
    } catch (err) {
      setModalError(err instanceof Error ? err.message : '保存失败');
    } finally {
      setSaving(false);
    }
  }

  /* ─── Archive ─── */

  async function handleArchive() {
    if (!token || !archiveTarget) return;
    setArchiving(true);
    try {
      await archiveTool(token, archiveTarget.id);
      setArchiveTarget(null);
      setPageSuccess('工具已归档');
      fetchTools();
    } catch (err) {
      setError(err instanceof Error ? err.message : '归档失败');
    } finally {
      setArchiving(false);
    }
  }

  /* ─── Toggle featured ─── */

  async function toggleFeatured(tool: AdminTool) {
    if (!token) return;
    try {
      await updateTool(token, tool.id, { featured: !tool.featured });
      fetchTools();
    } catch {
      // silent
    }
  }

  /* ─── Pagination helpers ─── */

  const totalPages = meta?.totalPages ?? 1;

  function pageNumbers(): number[] {
    const pages: number[] = [];
    const start = Math.max(1, page - 2);
    const end = Math.min(totalPages, page + 2);
    for (let i = start; i <= end; i++) {
      pages.push(i);
    }
    return pages;
  }

  /* ─── Render ─── */

  return (
    <div className={styles.page}>
      <div className={styles.container}>
        {/* Header */}
        <div className={styles.header}>
          <h1 className={styles.headerTitle}>TOOLS // MANAGEMENT</h1>
          <button className={styles.btnPrimary} onClick={openCreate} type="button">
            + 新建工具
          </button>
        </div>

        {/* Success alert */}
        {pageSuccess && (
          <div className={`${styles.alert} ${styles.alertSuccess}`}>
            {pageSuccess}
          </div>
        )}

        {/* Error alert */}
        {error && (
          <div className={`${styles.alert} ${styles.alertError}`}>
            {error}
          </div>
        )}

        {/* Toolbar */}
        <div className={styles.toolbar}>
          <input
            className={styles.searchInput}
            type="text"
            placeholder="搜索工具名称..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
          <div className={styles.filterTabs}>
            {STATUS_TABS.map((tab) => (
              <button
                key={tab.key}
                className={`${styles.filterTab} ${statusFilter === tab.key ? styles.filterTabActive : ''}`}
                onClick={() => setStatusFilter(tab.key)}
                type="button"
              >
                {tab.label}
              </button>
            ))}
          </div>
        </div>

        {/* Table */}
        {loading ? (
          <div className={styles.loading}>正在加载...</div>
        ) : tools.length === 0 ? (
          <div className={styles.empty}>
            <div className={styles.emptyIcon}>&#128736;</div>
            <div className={styles.emptyText}>
              {search || statusFilter ? '没有找到匹配的工具' : '暂无工具数据'}
            </div>
          </div>
        ) : (
          <>
            <div className={styles.tableWrapper}>
              <table className={styles.table}>
                <thead>
                  <tr>
                    <th className={styles.th}>名称</th>
                    <th className={styles.th}>分类</th>
                    <th className={styles.th}>难度</th>
                    <th className={styles.th}>状态</th>
                    <th className={styles.th}>精选</th>
                    <th className={styles.th}>操作</th>
                  </tr>
                </thead>
                <tbody>
                  {tools.map((tool) => (
                    <tr key={tool.id} className={styles.tr}>
                      <td className={styles.td}>
                        <span className={styles.toolName}>{tool.name}</span>
                        <span className={styles.toolSlug}>{tool.slug}</span>
                      </td>
                      <td className={styles.td}>
                        {categoryMap[tool.categoryId] || tool.categoryId}
                      </td>
                      <td className={styles.td}>
                        <span
                          className={`${styles.badge} ${difficultyBadgeClass[tool.difficulty] || ''}`}
                        >
                          {tool.difficulty}
                        </span>
                      </td>
                      <td className={styles.td}>
                        <span
                          className={`${styles.badge} ${statusBadgeClass[tool.status] || ''}`}
                        >
                          {tool.status}
                        </span>
                      </td>
                      <td className={styles.td}>
                        <button
                          className={`${styles.star} ${tool.featured ? styles.starActive : ''}`}
                          onClick={() => toggleFeatured(tool)}
                          type="button"
                          title={tool.featured ? '取消精选' : '设为精选'}
                        >
                          {tool.featured ? '\u2605' : '\u2606'}
                        </button>
                      </td>
                      <td className={styles.td}>
                        <div className={styles.rowActions}>
                          <button
                            className={styles.btnGhost}
                            onClick={() => openEdit(tool)}
                            type="button"
                          >
                            编辑
                          </button>
                          {tool.status !== 'archived' && (
                            <button
                              className={styles.btnDanger}
                              onClick={() => setArchiveTarget(tool)}
                              type="button"
                            >
                              归档
                            </button>
                          )}
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            {/* Pagination */}
            {totalPages > 1 && (
              <div className={styles.pagination}>
                <button
                  className={styles.pageBtn}
                  disabled={page <= 1}
                  onClick={() => setPage((p) => Math.max(1, p - 1))}
                  type="button"
                >
                  &laquo;
                </button>
                {pageNumbers().map((p) => (
                  <button
                    key={p}
                    className={`${styles.pageBtn} ${p === page ? styles.pageBtnActive : ''}`}
                    onClick={() => setPage(p)}
                    type="button"
                  >
                    {p}
                  </button>
                ))}
                <button
                  className={styles.pageBtn}
                  disabled={page >= totalPages}
                  onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                  type="button"
                >
                  &raquo;
                </button>
                <span className={styles.pageInfo}>
                  {meta?.total ?? 0} 条记录
                </span>
              </div>
            )}
          </>
        )}
      </div>

      {/* ─── Create / Edit Modal ─── */}
      {modalOpen && (
        <div className={styles.overlay} onClick={closeModal}>
          <div className={styles.modal} onClick={(e) => e.stopPropagation()}>
            <div className={styles.modalHeader}>
              <h2 className={styles.modalTitle}>
                {editingTool ? '编辑工具' : '新建工具'}
              </h2>
              <button
                className={styles.modalClose}
                onClick={closeModal}
                type="button"
              >
                &times;
              </button>
            </div>

            {modalError && (
              <div className={`${styles.alert} ${styles.alertError}`}>
                {modalError}
              </div>
            )}
            {modalSuccess && (
              <div className={`${styles.alert} ${styles.alertSuccess}`}>
                {modalSuccess}
              </div>
            )}

            {/* Slug + Name */}
            <div className={styles.fieldRow}>
              <div className={styles.field}>
                <label className={styles.label}>Slug</label>
                <input
                  className={styles.input}
                  type="text"
                  value={form.slug}
                  onChange={(e) => setField('slug', e.target.value)}
                  placeholder="my-tool"
                />
              </div>
              <div className={styles.field}>
                <label className={styles.label}>名称</label>
                <input
                  className={styles.input}
                  type="text"
                  value={form.name}
                  onChange={(e) => setField('name', e.target.value)}
                  placeholder="工具名称"
                />
              </div>
            </div>

            {/* Short description */}
            <div className={styles.field}>
              <label className={styles.label}>简介</label>
              <input
                className={styles.input}
                type="text"
                value={form.shortDescription}
                onChange={(e) => setField('shortDescription', e.target.value)}
                placeholder="一句话描述"
              />
            </div>

            {/* Long description */}
            <div className={styles.field}>
              <label className={styles.label}>详细描述</label>
              <textarea
                className={styles.textarea}
                value={form.longDescription}
                onChange={(e) => setField('longDescription', e.target.value)}
                placeholder="详细介绍工具的功能和用途..."
              />
            </div>

            {/* Category + Difficulty */}
            <div className={styles.fieldRow}>
              <div className={styles.field}>
                <label className={styles.label}>分类</label>
                <select
                  className={styles.select}
                  value={form.categoryId}
                  onChange={(e) => setField('categoryId', e.target.value)}
                >
                  <option value="">选择分类</option>
                  {categories.map((cat) => (
                    <option key={cat.id} value={cat.id}>
                      {cat.name}
                    </option>
                  ))}
                </select>
              </div>
              <div className={styles.field}>
                <label className={styles.label}>难度</label>
                <select
                  className={styles.select}
                  value={form.difficulty}
                  onChange={(e) => setField('difficulty', e.target.value)}
                >
                  {DIFFICULTY_OPTIONS.map((opt) => (
                    <option key={opt.value} value={opt.value}>
                      {opt.label}
                    </option>
                  ))}
                </select>
              </div>
            </div>

            {/* Icon + Status */}
            <div className={styles.fieldRow}>
              <div className={styles.field}>
                <label className={styles.label}>图标</label>
                <input
                  className={styles.input}
                  type="text"
                  value={form.icon}
                  onChange={(e) => setField('icon', e.target.value)}
                  placeholder="shield, terminal, lock..."
                />
              </div>
              <div className={styles.field}>
                <label className={styles.label}>状态</label>
                <select
                  className={styles.select}
                  value={form.status}
                  onChange={(e) => setField('status', e.target.value)}
                >
                  {STATUS_OPTIONS.map((opt) => (
                    <option key={opt.value} value={opt.value}>
                      {opt.label}
                    </option>
                  ))}
                </select>
              </div>
            </div>

            {/* Website URL + GitHub URL */}
            <div className={styles.fieldRow}>
              <div className={styles.field}>
                <label className={styles.label}>网站 URL</label>
                <input
                  className={styles.input}
                  type="url"
                  value={form.websiteUrl}
                  onChange={(e) => setField('websiteUrl', e.target.value)}
                  placeholder="https://example.com"
                />
              </div>
              <div className={styles.field}>
                <label className={styles.label}>GitHub URL</label>
                <input
                  className={styles.input}
                  type="url"
                  value={form.githubUrl}
                  onChange={(e) => setField('githubUrl', e.target.value)}
                  placeholder="https://github.com/..."
                />
              </div>
            </div>

            {/* Tags */}
            <div className={styles.field}>
              <label className={styles.label}>标签 (逗号分隔)</label>
              <input
                className={styles.input}
                type="text"
                value={form.tags}
                onChange={(e) => setField('tags', e.target.value)}
                placeholder="渗透测试, 网络安全, 密码学"
              />
            </div>

            {/* Featured checkbox */}
            <div className={styles.field}>
              <div className={styles.checkboxRow}>
                <input
                  className={styles.checkbox}
                  type="checkbox"
                  id="featured"
                  checked={form.featured}
                  onChange={(e) => setField('featured', e.target.checked)}
                />
                <label className={styles.checkboxLabel} htmlFor="featured">
                  设为精选工具
                </label>
              </div>
            </div>

            {/* Actions */}
            <div className={styles.actions}>
              <button
                className={styles.btnGhost}
                onClick={closeModal}
                type="button"
                disabled={saving}
              >
                取消
              </button>
              <button
                className={styles.btnPrimary}
                onClick={handleSave}
                type="button"
                disabled={saving}
              >
                {saving ? '保存中...' : editingTool ? '更新工具' : '创建工具'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* ─── Archive Confirmation Modal ─── */}
      {archiveTarget && (
        <div className={styles.overlay} onClick={() => setArchiveTarget(null)}>
          <div className={styles.modal} onClick={(e) => e.stopPropagation()}>
            <div className={styles.modalHeader}>
              <h2 className={styles.modalTitle}>确认归档</h2>
              <button
                className={styles.modalClose}
                onClick={() => setArchiveTarget(null)}
                type="button"
              >
                &times;
              </button>
            </div>

            <p className={styles.confirmText}>
              确定要归档工具{' '}
              <span className={styles.confirmName}>{archiveTarget.name}</span>{' '}
              吗？归档后该工具将不再对用户可见。
            </p>

            <div className={styles.actions}>
              <button
                className={styles.btnGhost}
                onClick={() => setArchiveTarget(null)}
                type="button"
                disabled={archiving}
              >
                取消
              </button>
              <button
                className={styles.btnDanger}
                onClick={handleArchive}
                type="button"
                disabled={archiving}
              >
                {archiving ? '归档中...' : '确认归档'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

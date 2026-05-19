'use client';

import { useCallback, useEffect, useState } from 'react';
import { toast } from 'sonner';
import { useAuth } from '@/components/AuthProvider';
import {
  AdminCategory,
  getAdminCategories,
  createCategory,
  updateCategory,
  deleteCategory,
} from '@/lib/admin';
import styles from './page.module.css';

interface CategoryForm {
  slug: string;
  name: string;
  description: string;
  icon: string;
  sortOrder: number;
  isVisible: boolean;
}

const emptyForm: CategoryForm = {
  slug: '',
  name: '',
  description: '',
  icon: '',
  sortOrder: 0,
  isVisible: true,
};

export default function AdminCategoriesPage() {
  const { token } = useAuth();
  const [categories, setCategories] = useState<AdminCategory[]>([]);
  const [loading, setLoading] = useState(true);

  // Modal state
  const [modalOpen, setModalOpen] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [form, setForm] = useState<CategoryForm>(emptyForm);
  const [saving, setSaving] = useState(false);

  // Track toggling / deleting
  const [togglingId, setTogglingId] = useState<string | null>(null);
  const [deletingId, setDeletingId] = useState<string | null>(null);

  const fetchCategories = useCallback(async () => {
    if (!token) return;
    setLoading(true);
    try {
      const data = await getAdminCategories(token);
      setCategories(data);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : '加载分类失败');
    } finally {
      setLoading(false);
    }
  }, [token]);

  useEffect(() => {
    void fetchCategories();
  }, [fetchCategories]);

  /* ── Visibility toggle ── */
  const handleToggleVisibility = async (cat: AdminCategory) => {
    if (!token) return;
    setTogglingId(cat.id);
    try {
      const updated = await updateCategory(token, cat.id, {
        isVisible: !cat.isVisible,
      });
      setCategories((prev) =>
        prev.map((c) => (c.id === cat.id ? updated : c)),
      );
      toast.success(`已${updated.isVisible ? '显示' : '隐藏'} "${cat.name}"`);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : '更新可见性失败');
    } finally {
      setTogglingId(null);
    }
  };

  /* ── Open modal ── */
  const openCreate = () => {
    setEditingId(null);
    setForm(emptyForm);
    setModalOpen(true);
  };

  const openEdit = (cat: AdminCategory) => {
    setEditingId(cat.id);
    setForm({
      slug: cat.slug,
      name: cat.name,
      description: cat.description,
      icon: cat.icon,
      sortOrder: cat.sortOrder,
      isVisible: cat.isVisible,
    });
    setModalOpen(true);
  };

  const closeModal = () => {
    setModalOpen(false);
    setEditingId(null);
    setForm(emptyForm);
  };

  /* ── Save (create or update) ── */
  const handleSave = async () => {
    if (!token) return;
    setSaving(true);
    try {
      if (editingId) {
        const updated = await updateCategory(token, editingId, form);
        setCategories((prev) =>
          prev.map((c) => (c.id === editingId ? updated : c)),
        );
        toast.success(`分类 "${updated.name}" 已更新`);
      } else {
        const created = await createCategory(token, form);
        setCategories((prev) => [...prev, created]);
        toast.success(`分类 "${created.name}" 已创建`);
      }
      closeModal();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : '保存失败');
    } finally {
      setSaving(false);
    }
  };

  /* ── Delete (soft delete = hide) ── */
  const handleDelete = async (cat: AdminCategory) => {
    if (!token) return;
    if (!window.confirm(`确定要删除分类 "${cat.name}" 吗？`)) return;
    setDeletingId(cat.id);
    try {
      await deleteCategory(token, cat.id);
      setCategories((prev) => prev.filter((c) => c.id !== cat.id));
      toast.success(`分类 "${cat.name}" 已删除`);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : '删除失败');
    } finally {
      setDeletingId(null);
    }
  };

  /* ── Render ── */

  if (loading) {
    return (
      <div className={styles.page}>
        <div className={styles.loadingWrapper}>
          <span className={styles.spinner} />
        </div>
      </div>
    );
  }

  return (
    <div className={styles.page}>
      {/* Header */}
      <div className={styles.header}>
        <h1 className={styles.headerTitle}>&gt; CATEGORIES // MANAGEMENT</h1>
        <div className={styles.toolbar}>
          <button className={styles.btnPrimary} onClick={openCreate}>
            + 新建分类
          </button>
        </div>
      </div>

      {/* Table */}
      {categories.length === 0 ? (
        <div className={styles.empty}>暂无分类数据</div>
      ) : (
        <table className={styles.table}>
          <thead>
            <tr>
              <th className={styles.th}>名称</th>
              <th className={styles.th}>Slug</th>
              <th className={styles.th}>图标</th>
              <th className={styles.th}>排序</th>
              <th className={styles.th}>可见性</th>
              <th className={styles.th}>操作</th>
            </tr>
          </thead>
          <tbody>
            {categories.map((cat) => (
              <tr key={cat.id} className={styles.tr}>
                <td className={styles.td}>{cat.name}</td>
                <td className={styles.td}>
                  <code>{cat.slug}</code>
                </td>
                <td className={styles.td}>{cat.icon || '—'}</td>
                <td className={styles.td}>{cat.sortOrder}</td>
                <td className={styles.td}>
                  <button
                    className={styles.visibility}
                    onClick={() => handleToggleVisibility(cat)}
                    disabled={togglingId === cat.id}
                  >
                    {togglingId === cat.id ? (
                      <span className={styles.spinner} />
                    ) : (
                      <span
                        className={`${styles.visibilityDot} ${
                          cat.isVisible
                            ? styles.visibilityDotActive
                            : styles.visibilityDotInactive
                        }`}
                      />
                    )}
                    {cat.isVisible ? '可见' : '隐藏'}
                  </button>
                </td>
                <td className={styles.td}>
                  <div style={{ display: 'flex', gap: 8 }}>
                    <button
                      className={styles.btnGhost}
                      onClick={() => openEdit(cat)}
                    >
                      编辑
                    </button>
                    <button
                      className={styles.btnDanger}
                      onClick={() => handleDelete(cat)}
                      disabled={deletingId === cat.id}
                    >
                      {deletingId === cat.id ? (
                        <span className={styles.spinner} />
                      ) : (
                        '删除'
                      )}
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}

      {/* Modal */}
      {modalOpen && (
        <div className={styles.overlay} onClick={closeModal}>
          <div className={styles.modal} onClick={(e) => e.stopPropagation()}>
            <div className={styles.modalHeader}>
              {editingId ? '> 编辑分类' : '> 新建分类'}
            </div>

            <div className={styles.modalBody}>
              <div className={styles.field}>
                <label className={styles.label}>Slug</label>
                <input
                  className={styles.input}
                  type="text"
                  value={form.slug}
                  onChange={(e) =>
                    setForm((f) => ({ ...f, slug: e.target.value }))
                  }
                  placeholder="category-slug"
                />
              </div>

              <div className={styles.field}>
                <label className={styles.label}>名称</label>
                <input
                  className={styles.input}
                  type="text"
                  value={form.name}
                  onChange={(e) =>
                    setForm((f) => ({ ...f, name: e.target.value }))
                  }
                  placeholder="分类名称"
                />
              </div>

              <div className={styles.field}>
                <label className={styles.label}>描述</label>
                <textarea
                  className={styles.textarea}
                  value={form.description}
                  onChange={(e) =>
                    setForm((f) => ({ ...f, description: e.target.value }))
                  }
                  placeholder="分类描述..."
                />
              </div>

              <div className={styles.field}>
                <label className={styles.label}>图标</label>
                <input
                  className={styles.input}
                  type="text"
                  value={form.icon}
                  onChange={(e) =>
                    setForm((f) => ({ ...f, icon: e.target.value }))
                  }
                  placeholder="emoji 或图标名称"
                />
              </div>

              <div className={styles.field}>
                <label className={styles.label}>排序</label>
                <input
                  className={styles.input}
                  type="number"
                  value={form.sortOrder}
                  onChange={(e) =>
                    setForm((f) => ({
                      ...f,
                      sortOrder: parseInt(e.target.value, 10) || 0,
                    }))
                  }
                />
              </div>

              <label className={styles.checkbox}>
                <input
                  type="checkbox"
                  checked={form.isVisible}
                  onChange={(e) =>
                    setForm((f) => ({ ...f, isVisible: e.target.checked }))
                  }
                />
                可见
              </label>
            </div>

            <div className={styles.actions}>
              <button
                className={styles.btnGhost}
                onClick={closeModal}
                disabled={saving}
              >
                取消
              </button>
              <button
                className={styles.btnPrimary}
                onClick={handleSave}
                disabled={saving || !form.slug || !form.name}
              >
                {saving ? <span className={styles.spinner} /> : null}
                {editingId ? '保存修改' : '创建分类'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

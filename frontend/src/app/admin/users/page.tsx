'use client';

import { useCallback, useEffect, useState } from 'react';
import { useAuth } from '@/components/AuthProvider';
import {
  AdminUser,
  PaginationMeta,
  getAdminUsers,
  updateUserRole,
  toggleUserActive,
} from '@/lib/admin';
import { toast } from 'sonner';
import styles from './page.module.css';

const PAGE_SIZE = 20;

function getRoleBadgeClass(role: string): string {
  switch (role) {
    case 'admin':
      return styles.roleBadgeAdmin;
    case 'editor':
      return styles.roleBadgeEditor;
    default:
      return styles.roleBadgeViewer;
  }
}

function formatDate(dateStr?: string): string {
  if (!dateStr) return '—';
  const d = new Date(dateStr);
  return d.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  });
}

export default function AdminUsersPage() {
  const { token, user: currentUser } = useAuth();
  const [users, setUsers] = useState<AdminUser[]>([]);
  const [meta, setMeta] = useState<PaginationMeta | null>(null);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);

  // Track in-flight operations
  const [updatingRoleId, setUpdatingRoleId] = useState<string | null>(null);
  const [togglingActiveId, setTogglingActiveId] = useState<string | null>(null);

  const fetchUsers = useCallback(async () => {
    if (!token) return;
    setLoading(true);
    try {
      const res = await getAdminUsers(token, page, PAGE_SIZE);
      setUsers(res.data);
      setMeta(res.meta);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : '加载用户列表失败');
    } finally {
      setLoading(false);
    }
  }, [token, page]);

  useEffect(() => {
    void fetchUsers();
  }, [fetchUsers]);

  /* ── Role change ── */
  const handleRoleChange = async (u: AdminUser, newRole: string) => {
    if (!token || newRole === u.role) return;
    setUpdatingRoleId(u.id);
    try {
      const updated = await updateUserRole(token, u.id, newRole);
      setUsers((prev) => prev.map((x) => (x.id === u.id ? updated : x)));
      toast.success(`用户 "${u.displayName}" 角色已更新为 ${newRole}`);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : '更新角色失败');
    } finally {
      setUpdatingRoleId(null);
    }
  };

  /* ── Active toggle ── */
  const handleToggleActive = async (u: AdminUser) => {
    if (!token) return;
    setTogglingActiveId(u.id);
    try {
      await toggleUserActive(token, u.id, !u.isActive);
      setUsers((prev) =>
        prev.map((x) =>
          x.id === u.id ? { ...x, isActive: !x.isActive } : x,
        ),
      );
      toast.success(
        `用户 "${u.displayName}" 已${u.isActive ? '停用' : '启用'}`,
      );
    } catch (err) {
      toast.error(err instanceof Error ? err.message : '切换状态失败');
    } finally {
      setTogglingActiveId(null);
    }
  };

  /* ── Pagination helpers ── */
  const totalPages = meta?.totalPages ?? 1;

  const pageNumbers = (): number[] => {
    const pages: number[] = [];
    const start = Math.max(1, page - 2);
    const end = Math.min(totalPages, page + 2);
    for (let i = start; i <= end; i++) {
      pages.push(i);
    }
    return pages;
  };

  /* ── Self-check: prevent modifying own account ── */
  const isSelf = (u: AdminUser) => currentUser?.id === u.id;

  /* ── Render ── */

  if (loading && users.length === 0) {
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
        <h1 className={styles.headerTitle}>&gt; USERS // MANAGEMENT</h1>
      </div>

      {/* Table */}
      {users.length === 0 ? (
        <div className={styles.empty}>暂无用户数据</div>
      ) : (
        <table className={styles.table}>
          <thead>
            <tr>
              <th className={styles.th}>用户名</th>
              <th className={styles.th}>邮箱</th>
              <th className={styles.th}>角色</th>
              <th className={styles.th}>状态</th>
              <th className={styles.th}>最后登录</th>
              <th className={styles.th}>操作</th>
            </tr>
          </thead>
          <tbody>
            {users.map((u) => (
              <tr key={u.id} className={styles.tr}>
                <td className={styles.td}>
                  {u.displayName}
                  {isSelf(u) && (
                    <span
                      style={{
                        marginLeft: 6,
                        fontSize: '0.68rem',
                        color: 'var(--color-text-dim)',
                      }}
                    >
                      (你)
                    </span>
                  )}
                </td>
                <td className={styles.td}>{u.email}</td>
                <td className={styles.td}>
                  <span
                    className={`${styles.roleBadge} ${getRoleBadgeClass(u.role)}`}
                  >
                    {u.role}
                  </span>
                </td>
                <td className={styles.td}>
                  <span
                    className={`${styles.statusDot} ${
                      u.isActive
                        ? styles.statusDotActive
                        : styles.statusDotInactive
                    }`}
                  />
                  {u.isActive ? '活跃' : '停用'}
                </td>
                <td className={styles.td}>{formatDate(u.lastLoginAt)}</td>
                <td className={styles.td}>
                  <div className={styles.actionCell}>
                    {/* Role dropdown */}
                    <select
                      className={styles.roleSelect}
                      value={u.role}
                      onChange={(e) => handleRoleChange(u, e.target.value)}
                      disabled={isSelf(u) || updatingRoleId === u.id}
                      title={isSelf(u) ? '不能修改自己的角色' : '更改角色'}
                    >
                      <option value="admin">admin</option>
                      <option value="editor">editor</option>
                      <option value="viewer">viewer</option>
                    </select>

                    {updatingRoleId === u.id && (
                      <span className={styles.spinner} />
                    )}

                    {/* Active/Inactive toggle */}
                    <button
                      className={styles.btnGhost}
                      onClick={() => handleToggleActive(u)}
                      disabled={isSelf(u) || togglingActiveId === u.id}
                      title={
                        isSelf(u)
                          ? '不能停用自己的账户'
                          : u.isActive
                            ? '停用此用户'
                            : '启用此用户'
                      }
                    >
                      {togglingActiveId === u.id ? (
                        <span className={styles.spinner} />
                      ) : u.isActive ? (
                        '停用'
                      ) : (
                        '启用'
                      )}
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}

      {/* Pagination */}
      {totalPages > 1 && (
        <div className={styles.pagination}>
          <button
            className={styles.pageBtn}
            onClick={() => setPage((p) => Math.max(1, p - 1))}
            disabled={page === 1}
          >
            &lt;
          </button>

          {pageNumbers().map((n) => (
            <button
              key={n}
              className={`${styles.pageBtn} ${n === page ? styles.pageBtnActive : ''}`}
              onClick={() => setPage(n)}
            >
              {n}
            </button>
          ))}

          <span className={styles.pageInfo}>
            {page} / {totalPages}
          </span>

          <button
            className={styles.pageBtn}
            onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
            disabled={page === totalPages}
          >
            &gt;
          </button>
        </div>
      )}
    </div>
  );
}

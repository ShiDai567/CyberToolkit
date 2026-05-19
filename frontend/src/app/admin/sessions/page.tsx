'use client';

import { useCallback, useEffect, useState } from 'react';
import {
  AlertCircle,
  CheckCircle2,
  Globe,
  Monitor,
  Trash2,
  Clock,
  User,
  MapPin,
} from 'lucide-react';
import { useAuth } from '@/components/AuthProvider';
import {
  AdminSession,
  getAdminSessions,
  PaginationMeta,
  revokeAdminSession,
} from '@/lib/admin';
import styles from './page.module.css';

function formatDate(value?: string | null): string {
  if (!value) return '—';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return '—';
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  });
}

function formatBrowser(ua: string): string {
  if (!ua) return '未知';
  if (ua.includes('Edg/')) {
    const match = ua.match(/Edg\/([\d.]+)/);
    return `Edge ${match?.[1] ?? ''}`;
  }
  if (ua.includes('Chrome/')) {
    const match = ua.match(/Chrome\/([\d.]+)/);
    return `Chrome ${match?.[1] ?? ''}`;
  }
  if (ua.includes('Firefox/')) {
    const match = ua.match(/Firefox\/([\d.]+)/);
    return `Firefox ${match?.[1] ?? ''}`;
  }
  if (ua.includes('Safari/') && !ua.includes('Chrome')) {
    return 'Safari';
  }
  return '其他浏览器';
}

function formatOS(ua: string): string {
  if (!ua) return '';
  if (ua.includes('Windows')) return 'Windows';
  if (ua.includes('Macintosh') || ua.includes('Mac OS')) return 'macOS';
  if (ua.includes('Linux')) return 'Linux';
  if (ua.includes('Android')) return 'Android';
  if (ua.includes('iPhone') || ua.includes('iPad')) return 'iOS';
  return '';
}

function maskToken(token: string): string {
  if (token.length <= 12) return token;
  return `${token.slice(0, 6)}…${token.slice(-4)}`;
}

const ROLE_LABELS: Record<string, string> = {
  admin: '管理员',
  editor: '编辑者',
  viewer: '访客',
};

export default function AdminSessionsPage() {
  const { token } = useAuth();

  const [sessions, setSessions] = useState<AdminSession[]>([]);
  const [meta, setMeta] = useState<PaginationMeta>({
    page: 1,
    pageSize: 20,
    total: 0,
    totalPages: 0,
  });
  const [loading, setLoading] = useState(true);
  const [busyToken, setBusyToken] = useState<string | null>(null);
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

  const fetchSessions = useCallback(async (page = meta.page) => {
    if (!token) return;
    setLoading(true);
    try {
      const result = await getAdminSessions(token, page, meta.pageSize);
      setSessions(result.data);
      setMeta(result.meta);
    } catch {
      setMessage({ type: 'error', text: '加载会话列表失败' });
    } finally {
      setLoading(false);
    }
  }, [token, meta.pageSize, meta.page]);

  useEffect(() => {
    fetchSessions();
  }, [fetchSessions]);

  const handleRevoke = async (accessToken: string) => {
    setMessage(null);
    setBusyToken(accessToken);
    try {
      await revokeAdminSession(token!, accessToken);
      setMessage({ type: 'success', text: '会话已强制下线' });
      fetchSessions();
    } catch {
      setMessage({ type: 'error', text: '操作失败，请稍后重试' });
    } finally {
      setBusyToken(null);
    }
  };

  const handleRevokeAll = async () => {
    setMessage(null);
    if (!confirm('确定要强制所有其他会话下线吗？')) return;
    setBusyToken('__all__');
    try {
      await Promise.all(sessions.map((s) => revokeAdminSession(token!, s.accessToken)));
      setMessage({ type: 'success', text: '所有会话已强制下线' });
      fetchSessions();
    } catch {
      setMessage({ type: 'error', text: '操作失败，请稍后重试' });
    } finally {
      setBusyToken(null);
    }
  };

  return (
    <div className={styles.page}>
      <div className={styles.header}>
        <div>
          <h1 className={styles.headerTitle}>{'>'} 会话管理 // ACTIVE_SESSIONS</h1>
          <p className={styles.headerDesc}>
            当前共有 {meta.total} 个活跃会话
          </p>
        </div>
        <div style={{ display: 'flex', gap: 8 }}>
          <button
            type="button"
            className={styles.btnGhost}
            onClick={() => fetchSessions()}
            disabled={loading}
          >
            刷新
          </button>
          <button
            type="button"
            className={styles.btnDanger}
            onClick={handleRevokeAll}
            disabled={busyToken !== null || sessions.length === 0}
          >
            {busyToken === '__all__' ? '处理中…' : '全部下线'}
          </button>
        </div>
      </div>

      {message && (
        <div
          className={`${styles.alert} ${
            message.type === 'success' ? styles.alertSuccess : styles.alertError
          }`}
        >
          {message.type === 'success' ? <CheckCircle2 size={14} /> : <AlertCircle size={14} />}
          <span>{message.text}</span>
        </div>
      )}

      {loading ? (
        <div className={styles.loadingRow}>
          <div className={styles.spinner} />
          <span>正在加载会话列表…</span>
        </div>
      ) : sessions.length === 0 ? (
        <div className={styles.empty}>
          <Monitor size={32} />
          <p>暂无活跃会话</p>
        </div>
      ) : (
        <div className={styles.tableWrapper}>
          <table className={styles.table}>
            <thead>
              <tr>
                <th className={styles.th}>用户</th>
                <th className={styles.th}>设备</th>
                <th className={styles.th}>IP 地址</th>
                <th className={styles.th}>会话 ID</th>
                <th className={styles.th}>创建时间</th>
                <th className={styles.th}>最后活跃</th>
                <th className={styles.th}>过期时间</th>
                <th className={styles.th}>操作</th>
              </tr>
            </thead>
            <tbody>
              {sessions.map((session) => {
                const os = formatOS(session.userAgent);
                const browser = formatBrowser(session.userAgent);
                const isCurrentSession = session.accessToken === token;

                return (
                  <tr key={session.accessToken} className={styles.tr}>
                    <td className={styles.td}>
                      <div className={styles.userCell}>
                        <div className={styles.userAvatar}>
                          <User size={14} />
                        </div>
                        <div>
                          <div className={styles.userName}>{session.userDisplayName}</div>
                          <div className={styles.userEmail}>{session.userEmail}</div>
                          <span className={styles.roleBadge}>{ROLE_LABELS[session.userRole] ?? session.userRole}</span>
                        </div>
                      </div>
                    </td>
                    <td className={styles.td}>
                      <div className={styles.deviceCell}>
                        <Monitor size={14} className={styles.deviceIcon} />
                        <div>
                          <div className={styles.deviceName}>{browser}</div>
                          {os && <div className={styles.deviceOS}>{os}</div>}
                        </div>
                      </div>
                    </td>
                    <td className={styles.td}>
                      <div className={styles.ipCell}>
                        <MapPin size={12} />
                        {session.ipAddress || '—'}
                      </div>
                    </td>
                    <td className={`${styles.td} ${styles.tdMono}`}>
                      {maskToken(session.accessToken)}
                    </td>
                    <td className={`${styles.td} ${styles.tdSmall}`}>{formatDate(session.createdAt)}</td>
                    <td className={`${styles.td} ${styles.tdSmall}`}>
                      <div className={styles.lastActive}>
                        <Clock size={12} />
                        {formatDate(session.lastActiveAt)}
                      </div>
                    </td>
                    <td className={`${styles.td} ${styles.tdSmall}`}>{formatDate(session.expiresAt)}</td>
                    <td className={styles.td}>
                      <button
                        type="button"
                        className={styles.revokeBtn}
                        onClick={() => handleRevoke(session.accessToken)}
                        disabled={busyToken !== null || isCurrentSession}
                        title={isCurrentSession ? '当前会话' : '强制下线'}
                      >
                        <Trash2 size={14} />
                        {isCurrentSession ? '当前' : '下线'}
                      </button>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}

      {meta.totalPages > 1 && (
        <div className={styles.pagination}>
          <button
            type="button"
            className={styles.pageBtn}
            disabled={meta.page <= 1}
            onClick={() => fetchSessions(meta.page - 1)}
          >
            上一页
          </button>
          <span className={styles.pageInfo}>
            第 {meta.page} / {meta.totalPages} 页（共 {meta.total} 条）
          </span>
          <button
            type="button"
            className={styles.pageBtn}
            disabled={meta.page >= meta.totalPages}
            onClick={() => fetchSessions(meta.page + 1)}
          >
            下一页
          </button>
        </div>
      )}
    </div>
  );
}

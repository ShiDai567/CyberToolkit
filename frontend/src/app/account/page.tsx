'use client';

import { FormEvent, useCallback, useEffect, useMemo, useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import {
  Clock,
  Lock,
  LogOut,
  Mail,
  MapPin,
  Monitor,
  Save,
  ShieldCheck,
  Trash2,
  User as UserIcon,
  UserCog,
} from 'lucide-react';
import { toast } from 'sonner';
import { useAuth } from '@/components/AuthProvider';
import {
  changePassword,
  listUserSessions,
  revokeOtherSessions,
  revokeUserSession,
  updateProfile,
  type UserSession,
} from '@/lib/auth';
import styles from './page.module.css';

type TabId = 'profile' | 'security' | 'sessions';

const ROLE_LABELS: Record<string, string> = {
  admin: '管理员',
  editor: '编辑者',
  viewer: '访客',
};

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
  });
}

function getInitial(name: string): string {
  const trimmed = name.trim();
  if (!trimmed) return '?';
  return trimmed.charAt(0).toUpperCase();
}

function parseUA(ua: string): string {
  if (!ua) return '未知设备';
  let browser = 'Unknown';
  let os = 'Unknown';
  if (/Edg\//.test(ua)) browser = 'Edge';
  else if (/Chrome\//.test(ua) && !/Chromium\//.test(ua)) browser = 'Chrome';
  else if (/Firefox\//.test(ua)) browser = 'Firefox';
  else if (/Safari\//.test(ua) && !/Chrome\//.test(ua)) browser = 'Safari';
  if (/Windows NT/.test(ua)) os = 'Windows';
  else if (/Macintosh/.test(ua)) os = 'macOS';
  else if (/Android/.test(ua)) os = 'Android';
  else if (/iPhone|iPad/.test(ua)) os = 'iOS';
  else if (/Linux/.test(ua)) os = 'Linux';
  return `${browser} · ${os}`;
}

export default function AccountPage() {
  const router = useRouter();
  const { user, token, isLoading, isAuthenticated, signOut, updateUser } = useAuth();

  const [activeTab, setActiveTab] = useState<TabId>('profile');

  // Profile form
  const [displayName, setDisplayName] = useState('');
  const [profileSaving, setProfileSaving] = useState(false);

  // Password form
  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [passwordSaving, setPasswordSaving] = useState(false);

  // Session actions
  const [sessionBusy, setSessionBusy] = useState<'revoke' | 'logout' | null>(null);
  const [sessions, setSessions] = useState<UserSession[]>([]);
  const [sessionsLoading, setSessionsLoading] = useState(false);
  const [revokingToken, setRevokingToken] = useState<string | null>(null);

  const fetchSessions = useCallback(async () => {
    if (!token) return;
    setSessionsLoading(true);
    try {
      const data = await listUserSessions(token);
      setSessions(data);
    } catch {
      // silently ignore
    } finally {
      setSessionsLoading(false);
    }
  }, [token]);

  useEffect(() => {
    if (user) {
      setDisplayName(user.displayName);
    }
  }, [user]);

  useEffect(() => {
    if (activeTab === 'sessions') {
      fetchSessions();
    }
  }, [activeTab, fetchSessions]);

  const initial = useMemo(() => (user ? getInitial(user.displayName) : '?'), [user]);
  const roleLabel = user ? ROLE_LABELS[user.role] ?? user.role : '';
  const profileDirty = user ? displayName.trim() !== user.displayName : false;

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
          <p className={styles.emptyText}>需要先通过身份认证才能访问个人资料。</p>
          <Link href="/login?redirect=/account" className={styles.loginLink}>
            前往登录
          </Link>
        </div>
      </div>
    );
  }

  const handleProfileSubmit = async (e: FormEvent) => {
    e.preventDefault();

    const trimmed = displayName.trim();
    if (trimmed.length < 2 || trimmed.length > 32) {
      toast.error('昵称需在 2-32 个字符之间');
      return;
    }
    if (trimmed === user.displayName) {
      toast.info('新昵称与当前一致，无需保存');
      return;
    }

    setProfileSaving(true);
    try {
      const updated = await updateProfile(token, trimmed);
      updateUser(updated);
      setDisplayName(updated.displayName);
      toast.success('昵称已更新');
    } catch (err) {
      const text = err instanceof Error ? err.message : '更新失败，请稍后重试';
      toast.error(text);
    } finally {
      setProfileSaving(false);
    }
  };

  const handlePasswordSubmit = async (e: FormEvent) => {
    e.preventDefault();

    if (!currentPassword) {
      toast.error('请输入当前密码');
      return;
    }
    if (newPassword.length < 6) {
      toast.error('新密码至少需要 6 位字符');
      return;
    }
    if (newPassword !== confirmPassword) {
      toast.error('两次输入的新密码不一致');
      return;
    }
    if (newPassword === currentPassword) {
      toast.error('新密码不能与当前密码相同');
      return;
    }

    setPasswordSaving(true);
    try {
      const result = await changePassword(token, currentPassword, newPassword);
      setCurrentPassword('');
      setNewPassword('');
      setConfirmPassword('');
      const tail = result.revokedSessions > 0 ? ` 已强制下线 ${result.revokedSessions} 个其他设备会话。` : '';
      toast.success(`密码更新成功${tail}`);
    } catch (err) {
      const text = err instanceof Error ? err.message : '密码更新失败，请稍后重试';
      toast.error(text);
    } finally {
      setPasswordSaving(false);
    }
  };

  const handleRevokeOthers = async () => {
    setSessionBusy('revoke');
    try {
      const result = await revokeOtherSessions(token, true);
      if (result.revokedSessions === 0) {
        toast.info('当前没有其他设备登录。');
      } else {
        toast.success(`已退出 ${result.revokedSessions} 个其他设备的登录会话。`);
      }
      await fetchSessions();
    } catch (err) {
      const text = err instanceof Error ? err.message : '操作失败，请稍后重试';
      toast.error(text);
    } finally {
      setSessionBusy(null);
    }
  };

  const handleRevokeOne = async (accessToken: string) => {
    if (!token) return;
    setRevokingToken(accessToken);
    try {
      await revokeUserSession(token, accessToken);
      await fetchSessions();
    } catch (err) {
      const text = err instanceof Error ? err.message : '撤销失败，请稍后重试';
      toast.error(text);
    } finally {
      setRevokingToken(null);
    }
  };

  const handleSignOut = async () => {
    setSessionBusy('logout');
    try {
      await signOut();
      router.replace('/login');
    } finally {
      setSessionBusy(null);
    }
  };

  const tokenPreview = `${token.slice(0, 6)}…${token.slice(-4)}`;

  return (
    <div className={styles.page}>
      <div className={styles.bgGlow} aria-hidden />
      <div className={styles.bgGrid} aria-hidden />

      <div className={styles.container}>
        {/* Hero */}
        <section className={styles.hero}>
          <span className={styles.heroScan} aria-hidden />

          <div className={styles.avatar}>{initial}</div>

          <div className={styles.heroMain}>
            <span className={styles.heroLabel}>
              <span className={styles.heroLabelPrompt}>{'>'}</span> USER_PROFILE // IDENTITY
            </span>
            <h1 className={styles.heroName}>
              {user.displayName}
              <span className={styles.heroNameAccent}>_</span>
            </h1>
            <div className={styles.heroMeta}>
              <span className={styles.heroChip}>
                <Mail size={12} className={styles.heroChipAccent} />
                {user.email}
              </span>
              <span className={styles.heroChip}>
                <ShieldCheck size={12} className={styles.heroChipAccent} />
                ROLE · {roleLabel}
              </span>
              <span className={styles.heroChip}>
                <UserCog size={12} className={styles.heroChipAccent} />
                JOINED · {formatDate(user.createdAt)}
              </span>
            </div>
          </div>

          <div className={styles.heroSide}>
            <span className={styles.statusPill}>
              <span className={styles.statusDot} />
              ONLINE
            </span>
            <span className={styles.uidLine}>UID · {user.id.slice(0, 8)}</span>
            <span className={styles.uidLine}>SESSION · {tokenPreview}</span>
          </div>
        </section>

        {/* Tabs */}
        <div className={styles.tabs} role="tablist">
          <button
            type="button"
            role="tab"
            aria-selected={activeTab === 'profile'}
            className={`${styles.tab} ${activeTab === 'profile' ? styles.tabActive : ''}`}
            onClick={() => setActiveTab('profile')}
          >
            <UserIcon size={14} />
            基本资料
          </button>
          <button
            type="button"
            role="tab"
            aria-selected={activeTab === 'security'}
            className={`${styles.tab} ${activeTab === 'security' ? styles.tabActive : ''}`}
            onClick={() => setActiveTab('security')}
          >
            <Lock size={14} />
            安全设置
          </button>
          <button
            type="button"
            role="tab"
            aria-selected={activeTab === 'sessions'}
            className={`${styles.tab} ${activeTab === 'sessions' ? styles.tabActive : ''}`}
            onClick={() => setActiveTab('sessions')}
          >
            <Monitor size={14} />
            会话管理
          </button>
        </div>

        {/* Profile Panel */}
        {activeTab === 'profile' && (
          <section className={styles.panel}>
            <header className={styles.panelHeader}>
              <UserIcon size={14} className={styles.panelTitleAccent} />
              ACCOUNT_INFO // <span className={styles.panelTitleAccent}>BASIC</span>
            </header>
            <div className={styles.panelBody}>
              <div className={styles.row}>
                <span className={styles.rowLabel}>
                  <span className={styles.rowLabelPrompt}>{'>'}</span> EMAIL
                </span>
                <span className={styles.rowValue}>{user.email}</span>
              </div>
              <div className={styles.row}>
                <span className={styles.rowLabel}>
                  <span className={styles.rowLabelPrompt}>{'>'}</span> ROLE
                </span>
                <span className={styles.rowValue}>{roleLabel}</span>
              </div>
              <div className={styles.row}>
                <span className={styles.rowLabel}>
                  <span className={styles.rowLabelPrompt}>{'>'}</span> USER_ID
                </span>
                <span className={`${styles.rowValue} ${styles.rowValueDim}`}>{user.id}</span>
              </div>
              <div className={styles.row}>
                <span className={styles.rowLabel}>
                  <span className={styles.rowLabelPrompt}>{'>'}</span> CREATED_AT
                </span>
                <span className={styles.rowValue}>{formatDate(user.createdAt)}</span>
              </div>

              <form className={styles.form} onSubmit={handleProfileSubmit} style={{ marginTop: 24 }} noValidate>
                <div className={styles.field}>
                  <label className={styles.label} htmlFor="displayName">
                    <span className={styles.labelPrompt}>{'>'}</span> DISPLAY_NAME
                  </label>
                  <div className={styles.inputWrapper}>
                    <UserIcon className={styles.inputIcon} />
                    <input
                      id="displayName"
                      className={styles.input}
                      type="text"
                      value={displayName}
                      onChange={(e) => {
                        setDisplayName(e.target.value);
                      }}
                      placeholder="输入新昵称"
                      maxLength={32}
                      disabled={profileSaving}
                    />
                  </div>
                  <span className={styles.hint}>2-32 个字符。修改后会立即对全站生效。</span>
                </div>

                <div className={styles.actions}>
                  <button
                    type="submit"
                    className={styles.btnPrimary}
                    disabled={profileSaving || !profileDirty}
                  >
                    {profileSaving ? (
                      <>
                        <span className={styles.spinner} />
                        保存中…
                      </>
                    ) : (
                      <>
                        <Save size={14} />
                        保存修改
                      </>
                    )}
                  </button>
                  <button
                    type="button"
                    className={styles.btnGhost}
                    disabled={profileSaving || !profileDirty}
                    onClick={() => {
                      setDisplayName(user.displayName);
                    }}
                  >
                    重置
                  </button>
                </div>
              </form>
            </div>
          </section>
        )}

        {/* Security Panel */}
        {activeTab === 'security' && (
          <section className={styles.panel}>
            <header className={styles.panelHeader}>
              <Lock size={14} className={styles.panelTitleAccent} />
              SECURITY // <span className={styles.panelTitleAccent}>PASSWORD</span>
            </header>
            <div className={styles.panelBody}>
              <form className={styles.form} onSubmit={handlePasswordSubmit} noValidate>
                <div className={styles.field}>
                  <label className={styles.label} htmlFor="currentPassword">
                    <span className={styles.labelPrompt}>{'>'}</span> CURRENT_PASSWORD
                  </label>
                  <div className={styles.inputWrapper}>
                    <Lock className={styles.inputIcon} />
                    <input
                      id="currentPassword"
                      className={styles.input}
                      type="password"
                      value={currentPassword}
                      autoComplete="current-password"
                      onChange={(e) => {
                        setCurrentPassword(e.target.value);
                      }}
                      disabled={passwordSaving}
                    />
                  </div>
                </div>

                <div className={styles.field}>
                  <label className={styles.label} htmlFor="newPassword">
                    <span className={styles.labelPrompt}>{'>'}</span> NEW_PASSWORD
                  </label>
                  <div className={styles.inputWrapper}>
                    <Lock className={styles.inputIcon} />
                    <input
                      id="newPassword"
                      className={styles.input}
                      type="password"
                      value={newPassword}
                      autoComplete="new-password"
                      onChange={(e) => {
                        setNewPassword(e.target.value);
                      }}
                      disabled={passwordSaving}
                    />
                  </div>
                  <span className={styles.hint}>至少 6 位字符，建议包含字母、数字与符号。</span>
                </div>

                <div className={styles.field}>
                  <label className={styles.label} htmlFor="confirmPassword">
                    <span className={styles.labelPrompt}>{'>'}</span> CONFIRM_PASSWORD
                  </label>
                  <div className={styles.inputWrapper}>
                    <Lock className={styles.inputIcon} />
                    <input
                      id="confirmPassword"
                      className={styles.input}
                      type="password"
                      value={confirmPassword}
                      autoComplete="new-password"
                      onChange={(e) => {
                        setConfirmPassword(e.target.value);
                      }}
                      disabled={passwordSaving}
                    />
                  </div>
                </div>

                <div className={styles.actions}>
                  <button type="submit" className={styles.btnPrimary} disabled={passwordSaving}>
                    {passwordSaving ? (
                      <>
                        <span className={styles.spinner} />
                        更新中…
                      </>
                    ) : (
                      <>
                        <ShieldCheck size={14} />
                        更新密码
                      </>
                    )}
                  </button>
                </div>
              </form>

              <div className={styles.sessionWarning} style={{ marginTop: 22 }}>
                <strong>提示：</strong>修改密码成功后，其他设备的登录会话将被自动注销，仅保留当前会话。
              </div>
            </div>
          </section>
        )}

        {/* Sessions Panel */}
        {activeTab === 'sessions' && (
          <section className={styles.panel}>
            <header className={styles.panelHeader}>
              <Monitor size={14} className={styles.panelTitleAccent} />
              SESSIONS // <span className={styles.panelTitleAccent}>ACTIVE</span>
            </header>
            <div className={styles.panelBody}>
              {sessionsLoading ? (
                <div className={styles.sessionsLoading}>
                  <span className={styles.spinner} style={{ borderTopColor: 'var(--color-cta)' }} />
                  <span>加载会话列表…</span>
                </div>
              ) : sessions.length === 0 ? (
                <div className={styles.sessionsEmpty}>暂无活跃会话数据</div>
              ) : (
                <div className={styles.sessionList}>
                  {sessions.map((sess) => (
                    <div
                      key={sess.accessToken}
                      className={`${styles.sessionCard} ${sess.isCurrent ? styles.sessionCardCurrent : ''}`}
                    >
                      <div className={styles.sessionCardLeft}>
                        <div className={styles.sessionIcon}>
                          <Monitor size={18} />
                        </div>
                        <div>
                          <div className={styles.sessionTitle}>
                            {parseUA(sess.userAgent)}
                            {sess.isCurrent && (
                              <span className={styles.currentBadge}>当前</span>
                            )}
                          </div>
                          <div className={styles.sessionMeta}>
                            {sess.ipAddress && (
                              <span><MapPin size={10} /> {sess.ipAddress}</span>
                            )}
                            <span><Clock size={10} /> 活跃 {formatDate(sess.lastActiveAt)}</span>
                            <span className={styles.sessionToken}>{sess.tokenPreview}</span>
                          </div>
                        </div>
                      </div>
                      {!sess.isCurrent && (
                        <button
                          type="button"
                          className={styles.btnRevokeSmall}
                          disabled={revokingToken === sess.accessToken || sessionBusy !== null}
                          onClick={() => handleRevokeOne(sess.accessToken)}
                        >
                          {revokingToken === sess.accessToken ? (
                            <span className={styles.spinner} style={{ borderTopColor: '#fecaca' }} />
                          ) : (
                            <Trash2 size={13} />
                          )}
                          撤销
                        </button>
                      )}
                      {sess.isCurrent && (
                        <span className={styles.statusPill}>
                          <span className={styles.statusDot} />
                          ACTIVE
                        </span>
                      )}
                    </div>
                  ))}
                </div>
              )}

              <div className={styles.actions} style={{ marginTop: 18 }}>
                <button
                  type="button"
                  className={styles.btnGhost}
                  onClick={handleRevokeOthers}
                  disabled={sessionBusy !== null || revokingToken !== null}
                >
                  {sessionBusy === 'revoke' ? (
                    <>
                      <span className={styles.spinner} />
                      处理中…
                    </>
                  ) : (
                    <>
                      <ShieldCheck size={14} />
                      退出其他设备
                    </>
                  )}
                </button>
                <button
                  type="button"
                  className={styles.btnDanger}
                  onClick={handleSignOut}
                  disabled={sessionBusy !== null || revokingToken !== null}
                >
                  {sessionBusy === 'logout' ? (
                    <>
                      <span className={`${styles.spinner} ${styles.spinnerDanger}`} />
                      退出中…
                    </>
                  ) : (
                    <>
                      <LogOut size={14} />
                      退出登录
                    </>
                  )}
                </button>
              </div>

              <div className={styles.sessionWarning}>
                出于安全考虑，请勿在公共设备上保持登录。如发现异常活动，建议立即修改密码。
              </div>
            </div>
          </section>
        )}
      </div>
    </div>
  );
}

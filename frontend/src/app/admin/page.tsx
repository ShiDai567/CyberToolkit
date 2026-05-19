'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import {
  Wrench,
  FolderTree,
  Users,
  Inbox,
} from 'lucide-react';
import { toast } from 'sonner';
import { useAuth } from '@/components/AuthProvider';
import { getAdminStats, type AdminStats } from '@/lib/admin';
import styles from './page.module.css';

const STAT_CARDS = [
  {
    key: 'toolCount' as const,
    label: '工具总数',
    icon: Wrench,
    sub: (s: AdminStats) => `已发布 ${s.publishedToolCount} / 草稿 ${s.draftToolCount}`,
  },
  {
    key: 'categoryCount' as const,
    label: '分类数',
    icon: FolderTree,
    sub: () => '工具分类总数',
  },
  {
    key: 'userCount' as const,
    label: '用户数',
    icon: Users,
    sub: (s: AdminStats) => `活跃用户 ${s.activeUserCount}`,
  },
  {
    key: 'pendingSubmissionCount' as const,
    label: '待审核投稿',
    icon: Inbox,
    sub: (s: AdminStats) => `共 ${s.submissionCount} 条投稿`,
  },
];

const QUICK_ACTIONS = [
  { href: '/admin/tools', label: '管理工具', icon: Wrench },
  { href: '/admin/categories', label: '管理分类', icon: FolderTree },
  { href: '/admin/users', label: '管理用户', icon: Users },
  { href: '/admin/submissions', label: '审核投稿', icon: Inbox },
];

export default function AdminDashboardPage() {
  const { token } = useAuth();
  const [stats, setStats] = useState<AdminStats | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!token) return;

    let cancelled = false;

    async function fetchStats() {
      try {
        const data = await getAdminStats(token!);
        if (!cancelled) {
          setStats(data);
        }
      } catch (err) {
        if (!cancelled) {
          toast.error(err instanceof Error ? err.message : '获取统计数据失败');
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    }

    void fetchStats();

    return () => {
      cancelled = true;
    };
  }, [token]);

  return (
    <div>
      {/* Terminal header */}
      <div className={styles.header}>
        <h1 className={styles.headerPrompt}>
          <span className={styles.headerPromptChar}>&gt; </span>
          DASHBOARD // OVERVIEW
        </h1>
        <p className={styles.headerDesc}>系统运行状态总览</p>
      </div>

      {/* Loading skeleton */}
      {loading && (
        <div className={styles.skeleton}>
          {Array.from({ length: 4 }).map((_, i) => (
            <div key={i} className={styles.skeletonCard} />
          ))}
        </div>
      )}

      {/* Stats grid */}
      {!loading && stats && (
        <div className={styles.grid}>
          {STAT_CARDS.map(({ key, label, icon: Icon, sub }) => (
            <div key={key} className={styles.card}>
              <Icon size={24} className={styles.cardIcon} />
              <div className={styles.cardValue}>{stats[key]}</div>
              <div className={styles.cardLabel}>{label}</div>
              <div className={styles.cardSub}>{sub(stats)}</div>
            </div>
          ))}
        </div>
      )}

      {/* Quick actions */}
      {!loading && stats && (
        <>
          <div className={styles.quickActionsHeader}>
            <span className={styles.headerPromptChar}>&gt; </span>
            QUICK_ACTIONS // 快捷操作
          </div>
          <div className={styles.quickActions}>
            {QUICK_ACTIONS.map(({ href, label, icon: Icon }) => (
              <Link key={href} href={href} className={styles.quickAction}>
                <Icon size={18} className={styles.quickActionIcon} />
                <span>{label}</span>
              </Link>
            ))}
          </div>
        </>
      )}
    </div>
  );
}

'use client';

import { useEffect, type ReactNode } from 'react';
import Link from 'next/link';
import { usePathname, useRouter } from 'next/navigation';
import {
  Shield,
  LayoutDashboard,
  Wrench,
  FolderTree,
  Users,
  Inbox,
  ArrowLeft,
  User,
} from 'lucide-react';
import { useAuth } from '@/components/AuthProvider';
import styles from './layout.module.css';

const NAV_ITEMS = [
  { href: '/admin', label: '仪表盘', icon: LayoutDashboard, exact: true },
  { href: '/admin/tools', label: '工具管理', icon: Wrench, exact: false },
  { href: '/admin/categories', label: '分类管理', icon: FolderTree, exact: false },
  { href: '/admin/users', label: '用户管理', icon: Users, exact: false },
  { href: '/admin/submissions', label: '投稿审核', icon: Inbox, exact: false },
];

export default function AdminLayout({ children }: { children: ReactNode }) {
  const { user, isLoading, isAuthenticated, signOut } = useAuth();
  const pathname = usePathname();
  const router = useRouter();

  useEffect(() => {
    if (!isLoading && (!isAuthenticated || !user || user.role !== 'admin')) {
      router.replace('/');
    }
  }, [isLoading, isAuthenticated, user, router]);

  if (isLoading) {
    return (
      <div className={styles.loadingWrapper}>
        <div className={styles.loadingContent}>
          <div className={styles.loadingSpinner} />
          <div className={styles.loadingText}>&gt; AUTHENTICATING...</div>
        </div>
      </div>
    );
  }

  if (!isAuthenticated || !user || user.role !== 'admin') {
    return (
      <div className={styles.loadingWrapper}>
        <div className={styles.loadingContent}>
          <div className={styles.loadingText}>&gt; ACCESS_DENIED // REDIRECTING...</div>
        </div>
      </div>
    );
  }

  return (
    <div className={styles.layout}>
      {/* Sidebar */}
      <aside className={styles.sidebar}>
        <div className={styles.logo}>
          <Shield size={24} className={styles.logoIcon} />
          <div className={styles.logoText}>
            CyberToolkit
            <span className={styles.logoBadge}>ADMIN</span>
          </div>
        </div>

        <nav className={styles.nav}>
          {NAV_ITEMS.map(({ href, label, icon: Icon, exact }) => {
            const isActive = exact
              ? pathname === href
              : pathname.startsWith(href);
            return (
              <Link
                key={href}
                href={href}
                className={`${styles.navLink} ${isActive ? styles.navLinkActive : ''}`}
              >
                <Icon size={18} className={styles.navIcon} />
                <span>{label}</span>
              </Link>
            );
          })}
        </nav>

        <div className={styles.sidebarFooter}>
          <Link href="/" className={styles.backLink}>
            <ArrowLeft size={16} />
            <span>返回前台</span>
          </Link>
        </div>
      </aside>

      {/* Content */}
      <div className={styles.content}>
        <header className={styles.topBar}>
          <div className={styles.userInfo}>
            <User size={16} />
            <span>{user.displayName}</span>
            <span className={styles.userRole}>{user.role}</span>
          </div>
          <button
            type="button"
            className={styles.logoutBtn}
            onClick={async () => {
              await signOut();
              router.replace('/login');
            }}
          >
            退出登录
          </button>
        </header>

        <main className={styles.main}>{children}</main>
      </div>
    </div>
  );
}

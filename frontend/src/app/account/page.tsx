'use client';

import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/components/AuthProvider';
import styles from '../admin/page.module.css';

export default function AccountPage() {
  const router = useRouter();
  const { user, isLoading, isAuthenticated, signOut } = useAuth();

  if (isLoading) {
    return (
      <div className={styles.empty}>
        <div className={styles.emptyCard}>正在验证登录状态...</div>
      </div>
    );
  }

  if (!isAuthenticated || !user) {
    return (
      <div className={styles.empty}>
        <div className={styles.emptyCard}>
          <h1>未登录</h1>
          <p className={styles.emptyText}>需要先登录账户才能访问个人中心。</p>
          <Link href="/login?redirect=/account" className={styles.loginLink}>
            前往登录
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className={styles.page}>
      <div className={styles.container}>
        <div className={styles.header}>
          <div>
            <h1 className={styles.title}>
              个人<span className="neon-text">中心</span>
            </h1>
            <p className={styles.desc}>当前页面用于展示登录后的用户身份和角色信息。</p>
          </div>
          <button
            className={styles.logout}
            onClick={async () => {
              await signOut();
              router.replace('/login');
            }}
            type="button"
          >
            退出登录
          </button>
        </div>

        <div className={styles.grid}>
          <div className={styles.card}>
            <div className={styles.cardTitle}>当前用户</div>
            <div className={styles.cardValue}>{user.displayName}</div>
          </div>
          <div className={styles.card}>
            <div className={styles.cardTitle}>邮箱</div>
            <div className={styles.cardValue}>{user.email}</div>
          </div>
          <div className={styles.card}>
            <div className={styles.cardTitle}>角色</div>
            <div className={styles.cardValue}>{user.role}</div>
          </div>
        </div>
      </div>
    </div>
  );
}

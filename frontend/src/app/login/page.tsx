'use client';

import Link from 'next/link';
import { useEffect, useState } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { LockKeyhole, Shield, UserRound, PencilRuler } from 'lucide-react';
import { useAuth } from '@/components/AuthProvider';
import { getClientAPIBaseURL, type AuthSession } from '@/lib/auth';
import styles from './page.module.css';

export default function LoginPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { signIn, isAuthenticated, isLoading } = useAuth();
  const [email, setEmail] = useState('viewer@cybertoolkit.local');
  const [password, setPassword] = useState('viewer123456');
  const [error, setError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  useEffect(() => {
    if (!isLoading && isAuthenticated) {
      router.replace(searchParams.get('redirect') || '/account');
    }
  }, [isAuthenticated, isLoading, router, searchParams]);

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError('');
    setIsSubmitting(true);

    try {
      const response = await fetch(`${getClientAPIBaseURL()}/api/v1/auth/login`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ email, password }),
      });

      const payload = (await response.json()) as
        | { data: AuthSession }
        | { error?: { message?: string } };

      if (!response.ok || !('data' in payload)) {
        throw new Error(payload.error?.message || '登录失败');
      }

      signIn(payload.data);
      const redirect = searchParams.get('redirect') || '/account';
      router.replace(redirect);
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : '登录失败');
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <div className={styles.page}>
      <div className={styles.card}>
        <span className={styles.eyebrow}>
          <Shield size={14} />
          Unified Access
        </span>
        <h1 className={styles.title}>
          账户<span className="neon-text">登录</span>
        </h1>
        <p className={styles.desc}>统一登录入口，适用于普通用户、编辑和管理员角色。</p>

        <div className={styles.rolePanel}>
          <div className={styles.roleCard}>
            <div className={styles.roleTitle}>
              <UserRound size={14} />
              普通用户
            </div>
            <div className={styles.roleDesc}>浏览工具、提交建议、维护个人资料。</div>
          </div>
          <div className={styles.roleCard}>
            <div className={styles.roleTitle}>
              <PencilRuler size={14} />
              编辑
            </div>
            <div className={styles.roleDesc}>维护工具信息、分类和标签内容。</div>
          </div>
          <div className={styles.roleCard}>
            <div className={styles.roleTitle}>
              <Shield size={14} />
              管理员
            </div>
            <div className={styles.roleDesc}>访问后台控制台和管理接口。</div>
          </div>
        </div>

        <form className={styles.form} onSubmit={handleSubmit}>
          <div className={styles.field}>
            <label htmlFor="email" className={styles.label}>
              邮箱
            </label>
            <input
              id="email"
              className={styles.input}
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              autoComplete="username"
              required
            />
          </div>

          <div className={styles.field}>
            <label htmlFor="password" className={styles.label}>
              密码
            </label>
            <input
              id="password"
              className={styles.input}
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              autoComplete="current-password"
              required
            />
          </div>

          {error && <div className={styles.error}>{error}</div>}

          <div className={styles.actions}>
            <button className={styles.submit} type="submit" disabled={isSubmitting}>
              <LockKeyhole size={16} />
              {isSubmitting ? '登录中...' : '登录账户'}
            </button>
          </div>
        </form>

        <p className={styles.footer}>
          还没有账户？
          <Link href="/register" className={styles.footerLink}>
            立即注册
          </Link>
        </p>

        <p className={styles.hint}>
          演示账号：`viewer@cybertoolkit.local / viewer123456`、`editor@cybertoolkit.local / editor123456`、`admin@cybertoolkit.local / admin123456`
        </p>
      </div>
    </div>
  );
}

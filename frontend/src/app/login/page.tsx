'use client';

import { useState } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { LockKeyhole, Shield } from 'lucide-react';
import { useAuth } from '@/components/AuthProvider';
import { getClientAPIBaseURL, type AuthSession } from '@/lib/auth';
import styles from './page.module.css';

export default function LoginPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { signIn, isAuthenticated } = useAuth();
  const [email, setEmail] = useState('admin@cybertoolkit.local');
  const [password, setPassword] = useState('admin123456');
  const [error, setError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError('');
    setIsSubmitting(true);

    try {
      const response = await fetch(`${getClientAPIBaseURL()}/api/v1/admin/auth/login`, {
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
      const redirect = searchParams.get('redirect') || '/admin';
      router.replace(redirect);
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : '登录失败');
    } finally {
      setIsSubmitting(false);
    }
  }

  if (isAuthenticated) {
    router.replace('/admin');
  }

  return (
    <div className={styles.page}>
      <div className={styles.card}>
        <span className={styles.eyebrow}>
          <Shield size={14} />
          Admin Access
        </span>
        <h1 className={styles.title}>
          控制台<span className="neon-text">登录</span>
        </h1>
        <p className={styles.desc}>登录后可进入后台接口和后续管理页面。</p>

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
              <LockKeyhole size={16} style={{ marginRight: 8, verticalAlign: 'text-bottom' }} />
              {isSubmitting ? '登录中...' : '登录后台'}
            </button>
          </div>
        </form>

        <p className={styles.hint}>
          当前默认开发账号为 `admin@cybertoolkit.local / admin123456`。
        </p>
      </div>
    </div>
  );
}

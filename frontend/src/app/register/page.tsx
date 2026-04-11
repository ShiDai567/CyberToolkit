'use client';

import Link from 'next/link';
import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { UserPlus, Shield } from 'lucide-react';
import { useAuth } from '@/components/AuthProvider';
import { getClientAPIBaseURL, type AuthSession } from '@/lib/auth';
import loginStyles from '../login/page.module.css';

export default function RegisterPage() {
  const router = useRouter();
  const { signIn, isAuthenticated, isLoading } = useAuth();
  const [displayName, setDisplayName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [error, setError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  useEffect(() => {
    if (!isLoading && isAuthenticated) {
      router.replace('/account');
    }
  }, [isAuthenticated, isLoading, router]);

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError('');

    if (password !== confirmPassword) {
      setError('两次输入的密码不一致');
      return;
    }

    setIsSubmitting(true);
    try {
      const response = await fetch(`${getClientAPIBaseURL()}/api/v1/auth/register`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ email, password, displayName }),
      });

      const payload = (await response.json()) as
        | { data: AuthSession }
        | { error?: { message?: string } };

      if (!response.ok || !('data' in payload)) {
        throw new Error(payload.error?.message || '注册失败');
      }

      signIn(payload.data);
      router.replace('/account');
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : '注册失败');
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <div className={loginStyles.page}>
      <div className={loginStyles.card}>
        <span className={loginStyles.eyebrow}>
          <Shield size={14} />
          Create Account
        </span>
        <h1 className={loginStyles.title}>
          注册<span className="neon-text">账户</span>
        </h1>
        <p className={loginStyles.desc}>新用户默认注册为普通浏览角色，管理员和编辑角色由系统分配。</p>

        <form className={loginStyles.form} onSubmit={handleSubmit}>
          <div className={loginStyles.field}>
            <label htmlFor="displayName" className={loginStyles.label}>
              昵称
            </label>
            <input
              id="displayName"
              className={loginStyles.input}
              type="text"
              value={displayName}
              onChange={(e) => setDisplayName(e.target.value)}
              autoComplete="nickname"
              required
            />
          </div>

          <div className={loginStyles.field}>
            <label htmlFor="email" className={loginStyles.label}>
              邮箱
            </label>
            <input
              id="email"
              className={loginStyles.input}
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              autoComplete="email"
              required
            />
          </div>

          <div className={loginStyles.field}>
            <label htmlFor="password" className={loginStyles.label}>
              密码
            </label>
            <input
              id="password"
              className={loginStyles.input}
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              autoComplete="new-password"
              required
            />
          </div>

          <div className={loginStyles.field}>
            <label htmlFor="confirmPassword" className={loginStyles.label}>
              确认密码
            </label>
            <input
              id="confirmPassword"
              className={loginStyles.input}
              type="password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              autoComplete="new-password"
              required
            />
          </div>

          {error && <div className={loginStyles.error}>{error}</div>}

          <div className={loginStyles.actions}>
            <button className={loginStyles.submit} type="submit" disabled={isSubmitting}>
              <UserPlus size={16} />
              {isSubmitting ? '注册中...' : '创建账户'}
            </button>
          </div>
        </form>

        <p className={loginStyles.footer}>
          已有账户？
          <Link href="/login" className={loginStyles.footerLink}>
            去登录
          </Link>
        </p>
      </div>
    </div>
  );
}

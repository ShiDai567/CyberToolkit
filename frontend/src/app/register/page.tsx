'use client';

import { useEffect, useRef, useState, FormEvent } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { Mail, Lock, User, IdCard, Shield, AlertCircle } from 'lucide-react';
import { useAuth } from '@/components/AuthProvider';
import { register } from '@/lib/auth';
import styles from './page.module.css';

export default function RegisterPage() {
  const router = useRouter();
  const { signIn, isAuthenticated } = useAuth();
  const canvasRef = useRef<HTMLCanvasElement>(null);

  const [username, setUsername] = useState('');
  const [displayName, setDisplayName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [error, setError] = useState('');
  const [fieldErrors, setFieldErrors] = useState<{ username?: string; displayName?: string; email?: string; password?: string; confirmPassword?: string }>({});
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (isAuthenticated) {
      router.replace('/');
    }
  }, [isAuthenticated, router]);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    const resize = () => {
      canvas.width = window.innerWidth;
      canvas.height = window.innerHeight;
    };
    resize();
    window.addEventListener('resize', resize);

    const chars = '01アイウエオカキクケコ@#$%&*{}[]<>/\\|';
    const fontSize = 14;
    const columns = Math.floor(canvas.width / fontSize);
    const drops: number[] = new Array(columns).fill(1).map(() => Math.random() * -50);

    const draw = () => {
      ctx.fillStyle = 'rgba(15, 23, 42, 0.06)';
      ctx.fillRect(0, 0, canvas.width, canvas.height);

      ctx.font = `${fontSize}px "Fira Code", monospace`;

      for (let i = 0; i < drops.length; i++) {
        const brightness = Math.random();
        if (brightness > 0.5) {
          ctx.fillStyle = `rgba(34, 197, 94, ${0.08 + brightness * 0.12})`;
        } else {
          ctx.fillStyle = `rgba(34, 197, 94, 0.05)`;
        }
        const text = chars[Math.floor(Math.random() * chars.length)];
        ctx.fillText(text, i * fontSize, drops[i] * fontSize);

        if (drops[i] * fontSize > canvas.height && Math.random() > 0.975) {
          drops[i] = 0;
        }
        drops[i]++;
      }
    };

    const interval = setInterval(draw, 55);
    return () => {
      clearInterval(interval);
      window.removeEventListener('resize', resize);
    };
  }, []);

  const validate = (): boolean => {
    const errors: { username?: string; displayName?: string; email?: string; password?: string; confirmPassword?: string } = {};
    if (!username.trim()) {
      errors.username = '请输入用户名';
    } else if (!/^[a-zA-Z0-9_]{3,30}$/.test(username.trim())) {
      errors.username = '3-30个字符，仅支持字母、数字和下划线';
    }
    if (!displayName.trim()) {
      errors.displayName = '请输入昵称';
    } else if (displayName.trim().length < 2) {
      errors.displayName = '昵称至少2个字符';
    } else if (displayName.trim().length > 32) {
      errors.displayName = '昵称不超过32个字符';
    }
    if (!email.trim()) {
      errors.email = '请输入邮箱地址';
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
      errors.email = '邮箱格式不正确';
    }
    if (!password) {
      errors.password = '请输入密码';
    } else if (password.length < 6) {
      errors.password = '密码至少6个字符';
    }
    if (!confirmPassword) {
      errors.confirmPassword = '请确认密码';
    } else if (password !== confirmPassword) {
      errors.confirmPassword = '两次密码不一致';
    }
    setFieldErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError('');

    if (!validate()) return;

    setLoading(true);
    try {
      const session = await register(username.trim(), email, password, confirmPassword, displayName.trim());
      signIn(session);
      router.replace('/');
    } catch (err) {
      const message = err instanceof Error ? err.message : '注册失败，请重试';
      setError(message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className={styles.page}>
      <canvas ref={canvasRef} className={styles.canvas} />
      <div className={styles.overlay} />

      <div className={styles.card}>
        <div className={styles.cardGlow} />

        <div className={styles.terminalHeader}>
          <span className={styles.dot} style={{ background: '#EF4444' }} />
          <span className={styles.dot} style={{ background: '#F59E0B' }} />
          <span className={styles.dot} style={{ background: '#22C55E' }} />
          <span className={styles.terminalTitle}>AUTH_TERMINAL // REGISTER</span>
          <span className={styles.terminalStatus}>
            <span className={styles.statusDot} />
            SECURE
          </span>
        </div>

        <div className={styles.terminalBody}>
          <h1 className={styles.heading}>
            身份<span className={styles.headingAccent}>注册</span>
          </h1>
          <p className={styles.subheading}>
            创建新身份以获取安全工具库访问权限
          </p>

          {error && (
            <div className={styles.globalError}>
              <AlertCircle size={14} style={{ verticalAlign: 'middle', marginRight: 6 }} />
              {error}
            </div>
          )}

          <form className={styles.form} onSubmit={handleSubmit} noValidate>
            <div className={styles.field}>
              <label className={styles.label}>
                <span className={styles.labelPrompt}>{'>'}</span> USERNAME
              </label>
              <div className={styles.inputWrapper}>
                <IdCard className={styles.inputIcon} />
                <input
                  type="text"
                  className={`${styles.input} ${fieldErrors.username ? styles.inputError : ''}`}
                  placeholder="你的用户ID"
                  value={username}
                  onChange={(e) => { setUsername(e.target.value); setFieldErrors((f) => ({ ...f, username: undefined })); }}
                  autoComplete="username"
                  autoFocus
                />
              </div>
              <span className={styles.errorMsg}>{fieldErrors.username}</span>
            </div>

            <div className={styles.field}>
              <label className={styles.label}>
                <span className={styles.labelPrompt}>{'>'}</span> NICKNAME
              </label>
              <div className={styles.inputWrapper}>
                <User className={styles.inputIcon} />
                <input
                  type="text"
                  className={`${styles.input} ${fieldErrors.displayName ? styles.inputError : ''}`}
                  placeholder="你的昵称"
                  value={displayName}
                  onChange={(e) => { setDisplayName(e.target.value); setFieldErrors((f) => ({ ...f, displayName: undefined })); }}
                  autoComplete="name"
                />
              </div>
              <span className={styles.errorMsg}>{fieldErrors.displayName}</span>
            </div>

            <div className={styles.field}>
              <label className={styles.label}>
                <span className={styles.labelPrompt}>{'>'}</span> EMAIL
              </label>
              <div className={styles.inputWrapper}>
                <Mail className={styles.inputIcon} />
                <input
                  type="email"
                  className={`${styles.input} ${fieldErrors.email ? styles.inputError : ''}`}
                  placeholder="user@example.com"
                  value={email}
                  onChange={(e) => { setEmail(e.target.value); setFieldErrors((f) => ({ ...f, email: undefined })); }}
                  autoComplete="email"
                />
              </div>
              <span className={styles.errorMsg}>{fieldErrors.email}</span>
            </div>

            <div className={styles.field}>
              <label className={styles.label}>
                <span className={styles.labelPrompt}>{'>'}</span> PASSWORD
              </label>
              <div className={styles.inputWrapper}>
                <Lock className={styles.inputIcon} />
                <input
                  type="password"
                  className={`${styles.input} ${fieldErrors.password ? styles.inputError : ''}`}
                  placeholder="••••••••"
                  value={password}
                  onChange={(e) => { setPassword(e.target.value); setFieldErrors((f) => ({ ...f, password: undefined })); }}
                  autoComplete="new-password"
                />
              </div>
              <span className={styles.errorMsg}>{fieldErrors.password}</span>
            </div>

            <div className={styles.field}>
              <label className={styles.label}>
                <span className={styles.labelPrompt}>{'>'}</span> CONFIRM_PASSWORD
              </label>
              <div className={styles.inputWrapper}>
                <Lock className={styles.inputIcon} />
                <input
                  type="password"
                  className={`${styles.input} ${fieldErrors.confirmPassword ? styles.inputError : ''}`}
                  placeholder="••••••••"
                  value={confirmPassword}
                  onChange={(e) => { setConfirmPassword(e.target.value); setFieldErrors((f) => ({ ...f, confirmPassword: undefined })); }}
                  autoComplete="new-password"
                />
              </div>
              <span className={styles.errorMsg}>{fieldErrors.confirmPassword}</span>
            </div>

            <button
              type="submit"
              className={styles.submitBtn}
              disabled={loading}
            >
              {loading ? (
                <>
                  <span className={styles.spinner} />
                  创建身份中...
                </>
              ) : (
                '注册身份'
              )}
            </button>
          </form>

          <div className={styles.divider}>
            <span className={styles.dividerLine} />
            <span className={styles.dividerText}>OR</span>
            <span className={styles.dividerLine} />
          </div>

          <p className={styles.footer}>
            已有身份？{' '}
            <Link href="/login" className={styles.footerLink}>
              登录系统
            </Link>
          </p>

          <div className={styles.secureNote}>
            <Shield size={10} />
            <span>ENCRYPTED CONNECTION // TLS 1.3</span>
          </div>
        </div>
      </div>
    </div>
  );
}

'use client';

import Link from 'next/link';
import { useState } from 'react';
import { Shield, Terminal, Menu, X, LogIn, LogOut, UserRound, LayoutDashboard } from 'lucide-react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/components/AuthProvider';
import styles from './Navbar.module.css';

export function Navbar() {
  const [menuOpen, setMenuOpen] = useState(false);
  const router = useRouter();
  const { user, isLoading, isAuthenticated, signOut } = useAuth();

  return (
    <nav className={styles.nav}>
      <div className={styles.inner}>
        <Link href="/" className={styles.logo}>
          <Shield className={styles.logoIcon} />
          <span className={styles.logoText}>
            Cyber<span className={styles.logoAccent}>Toolkit</span>
          </span>
          <span className={styles.cursor}>_</span>
        </Link>

        <div className={`${styles.links} ${menuOpen ? styles.linksOpen : ''}`}>
          <Link href="/" className={styles.link} onClick={() => setMenuOpen(false)}>
            <Terminal size={14} />
            <span>首页</span>
          </Link>
          <Link href="/tools" className={styles.link} onClick={() => setMenuOpen(false)}>
            <Terminal size={14} />
            <span>工具库</span>
          </Link>
          <Link href="/tools" className={styles.ctaBtn} onClick={() => setMenuOpen(false)}>
            探索工具
          </Link>

          {!isLoading && !isAuthenticated && (
            <>
              <Link href="/login" className={styles.link} onClick={() => setMenuOpen(false)}>
                <LogIn size={14} />
                <span>登录</span>
              </Link>
              <Link href="/register" className={styles.link} onClick={() => setMenuOpen(false)}>
                <UserRound size={14} />
                <span>注册</span>
              </Link>
            </>
          )}

          {!isLoading && isAuthenticated && (
            <>
              <Link href="/account" className={styles.link} onClick={() => setMenuOpen(false)}>
                <UserRound size={14} />
                <span>{user?.displayName || '个人中心'}</span>
              </Link>
              {user?.role === 'admin' && (
                <Link href="/admin" className={styles.link} onClick={() => setMenuOpen(false)}>
                  <LayoutDashboard size={14} />
                  <span>控制台</span>
                </Link>
              )}
              <button
                type="button"
                className={styles.ghostBtn}
                onClick={async () => {
                  await signOut();
                  setMenuOpen(false);
                  router.push('/login');
                }}
              >
                <LogOut size={14} />
                <span>退出</span>
              </button>
            </>
          )}
        </div>

        <button
          className={styles.menuBtn}
          onClick={() => setMenuOpen(!menuOpen)}
          aria-label="Toggle menu"
        >
          {menuOpen ? <X size={24} /> : <Menu size={24} />}
        </button>
      </div>
    </nav>
  );
}

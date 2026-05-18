'use client';

import { useState } from 'react';
import Link from 'next/link';
import { Shield, Home, Terminal, LogOut, Menu, X, User } from 'lucide-react';
import { useAuth } from '@/components/AuthProvider';
import styles from './Navbar.module.css';

export function Navbar() {
  const { user, isAuthenticated, signOut } = useAuth();
  const [menuOpen, setMenuOpen] = useState(false);

  const handleSignOut = async () => {
    await signOut();
    setMenuOpen(false);
  };

  return (
    <nav className={styles.nav}>
      <div className={styles.inner}>
        <Link href="/" className={styles.logo}>
          <Shield className={styles.logoIcon} />
          <span className={styles.logoText}>
            Cyber<span className={styles.logoAccent}>Toolkit</span>
            <span className={styles.cursor}>_</span>
          </span>
        </Link>

        <div className={`${styles.links} ${menuOpen ? styles.linksOpen : ''}`}>
          <Link href="/" className={styles.link} onClick={() => setMenuOpen(false)}>
            <Home size={14} /> 首页
          </Link>
          <Link href="/tools" className={styles.link} onClick={() => setMenuOpen(false)}>
            <Terminal size={14} /> 工具库
          </Link>

          {isAuthenticated && user ? (
            <>
              <Link
                href="/account"
                className={styles.link}
                onClick={() => setMenuOpen(false)}
              >
                <User size={14} /> {user.displayName}
              </Link>
              <button className={styles.ghostBtn} onClick={handleSignOut}>
                <LogOut size={14} /> 退出
              </button>
            </>
          ) : (
            <>
              <Link href="/login" className={styles.link} onClick={() => setMenuOpen(false)}>
                登录
              </Link>
              <Link href="/register" className={styles.ctaBtn} onClick={() => setMenuOpen(false)}>
                注册
              </Link>
            </>
          )}
        </div>

        <button
          className={styles.menuBtn}
          onClick={() => setMenuOpen(!menuOpen)}
          aria-label={menuOpen ? '关闭菜单' : '打开菜单'}
        >
          {menuOpen ? <X size={20} /> : <Menu size={20} />}
        </button>
      </div>
    </nav>
  );
}

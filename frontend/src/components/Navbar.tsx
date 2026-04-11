'use client';

import Link from 'next/link';
import { useState } from 'react';
import { Shield, Terminal, Menu, X } from 'lucide-react';
import styles from './Navbar.module.css';

export function Navbar() {
  const [menuOpen, setMenuOpen] = useState(false);

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

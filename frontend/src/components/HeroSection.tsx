'use client';

import { useEffect, useRef, useState } from 'react';
import { ChevronRight, Shield } from 'lucide-react';
import Link from 'next/link';
import styles from './HeroSection.module.css';

const TYPING_TEXTS = [
  'Network Scanning & Enumeration',
  'Penetration Testing Frameworks',
  'Digital Forensics & OSINT',
  'Vulnerability Assessment Tools',
  'Password Cracking Utilities',
];

export function HeroSection() {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [typedText, setTypedText] = useState('');
  const [textIndex, setTextIndex] = useState(0);
  const [charIndex, setCharIndex] = useState(0);
  const [isDeleting, setIsDeleting] = useState(false);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    const resize = () => {
      canvas.width = window.innerWidth;
      canvas.height = 700;
    };
    resize();
    window.addEventListener('resize', resize);

    const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789@#$%^&*(){}[]|;:<>?/~';
    const fontSize = 14;
    const columns = Math.floor(canvas.width / fontSize);
    const drops: number[] = new Array(columns).fill(1).map(() => Math.random() * -50);

    const draw = () => {
      ctx.fillStyle = 'rgba(15, 23, 42, 0.05)';
      ctx.fillRect(0, 0, canvas.width, canvas.height);

      ctx.fillStyle = 'rgba(34, 197, 94, 0.15)';
      ctx.font = `${fontSize}px "Fira Code", monospace`;

      for (let i = 0; i < drops.length; i++) {
        const text = chars[Math.floor(Math.random() * chars.length)];
        const x = i * fontSize;
        const y = drops[i] * fontSize;

        ctx.fillText(text, x, y);

        if (y > canvas.height && Math.random() > 0.975) {
          drops[i] = 0;
        }
        drops[i]++;
      }
    };

    const interval = setInterval(draw, 50);
    return () => {
      clearInterval(interval);
      window.removeEventListener('resize', resize);
    };
  }, []);

  const pauseTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    const currentText = TYPING_TEXTS[textIndex];
    const speed = isDeleting ? 30 : 70;

    const timer = setTimeout(() => {
      if (!isDeleting) {
        const nextIndex = Math.min(currentText.length, charIndex + 1);
        setTypedText(currentText.substring(0, nextIndex));
        setCharIndex(nextIndex);

        if (nextIndex === currentText.length) {
          pauseTimerRef.current = setTimeout(() => setIsDeleting(true), 2000);
        }
      } else {
        const nextIndex = Math.max(0, charIndex - 1);
        setTypedText(currentText.substring(0, nextIndex));
        setCharIndex(nextIndex);

        if (nextIndex === 0) {
          setIsDeleting(false);
          setTextIndex((prev) => (prev + 1) % TYPING_TEXTS.length);
        }
      }
    }, speed);

    return () => {
      clearTimeout(timer);
      if (pauseTimerRef.current) {
        clearTimeout(pauseTimerRef.current);
      }
    };
  }, [charIndex, isDeleting, textIndex]);

  return (
    <section className={styles.hero}>
      <canvas ref={canvasRef} className={styles.canvas} />
      <div className={styles.overlay} />

      <div className={styles.content}>
        <div className={styles.badge}>
          <Shield size={14} />
          <span>安全工具库 v2.0</span>
        </div>

        <h1 className={styles.title}>
          <span className={styles.titleLine}>Your Ultimate</span>
          <span className={`${styles.titleLine} neon-text`}>Cybersecurity</span>
          <span className={styles.titleLine}>Toolkit</span>
        </h1>

        <div className={styles.terminal}>
          <div className={styles.terminalHeader}>
            <span className={styles.dot} style={{ background: '#EF4444' }} />
            <span className={styles.dot} style={{ background: '#F59E0B' }} />
            <span className={styles.dot} style={{ background: '#22C55E' }} />
            <span className={styles.terminalTitle}>Terminal</span>
          </div>
          <div className={styles.terminalBody}>
            <span className={styles.prompt}>root@cyber:~$</span>
            <span className={styles.typed}>{typedText}</span>
            <span className={styles.blinkCursor}>█</span>
          </div>
        </div>

        <p className={styles.subtitle}>
          探索 20+ 款精选网络安全工具，从扫描、评估到取证与情报收集，统一整理到一个入口里。
        </p>

        <div className={styles.actions}>
          <Link href="/tools" className={styles.ctaButton}>
            <span>探索工具</span>
            <ChevronRight size={18} />
          </Link>
          <a href="#categories" className={styles.secondaryButton}>
            浏览分类
          </a>
        </div>
      </div>
    </section>
  );
}

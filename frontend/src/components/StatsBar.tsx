'use client';

import { useEffect, useState, useRef } from 'react';
import { Wrench, Layers, Users, Shield } from 'lucide-react';
import styles from './StatsBar.module.css';

interface StatItemProps {
  icon: React.ElementType;
  value: number;
  label: string;
  suffix?: string;
}

function StatItem({ icon: Icon, value, label, suffix = '' }: StatItemProps) {
  const [count, setCount] = useState(0);
  const ref = useRef<HTMLDivElement>(null);
  const [inView, setInView] = useState(false);

  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) setInView(true);
      },
      { threshold: 0.3 }
    );

    if (ref.current) observer.observe(ref.current);
    return () => observer.disconnect();
  }, []);

  useEffect(() => {
    if (!inView) return;

    let start = 0;
    const end = value;
    const duration = 1500;
    const step = Math.ceil(end / (duration / 16));

    const timer = setInterval(() => {
      start += step;
      if (start >= end) {
        setCount(end);
        clearInterval(timer);
      } else {
        setCount(start);
      }
    }, 16);

    return () => clearInterval(timer);
  }, [inView, value]);

  return (
    <div className={styles.stat} ref={ref}>
      <Icon size={20} className={styles.statIcon} />
      <span className={styles.statValue}>
        {count}
        {suffix}
      </span>
      <span className={styles.statLabel}>{label}</span>
    </div>
  );
}

export function StatsBar() {
  return (
    <section className={styles.section}>
      <div className={styles.inner}>
        <StatItem icon={Wrench} value={20} label="安全工具" suffix="+" />
        <div className={styles.separator} />
        <StatItem icon={Layers} value={8} label="工具分类" />
        <div className={styles.separator} />
        <StatItem icon={Users} value={50} label="贡献者" suffix="K+" />
        <div className={styles.separator} />
        <StatItem icon={Shield} value={99} label="开源" suffix="%" />
      </div>
    </section>
  );
}

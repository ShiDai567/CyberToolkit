'use client';

import { Search } from 'lucide-react';
import styles from './SearchBar.module.css';

interface SearchBarProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
}

export function SearchBar({ value, onChange, placeholder = '搜索工具...' }: SearchBarProps) {
  return (
    <div className={styles.wrapper}>
      <span className={styles.prompt}>&gt;_</span>
      <input
        type="text"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        className={styles.input}
        aria-label="Search tools"
      />
      <Search size={16} className={styles.icon} />
    </div>
  );
}

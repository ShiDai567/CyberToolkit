import {
  Radar,
  ShieldAlert,
  Crosshair,
  KeyRound,
  Wifi,
  Globe,
  Microscope,
  Eye,
} from 'lucide-react';
import { Category } from '@/types/types';
import styles from './CategoryCard.module.css';

const iconMap: Record<string, React.ElementType> = {
  Radar,
  ShieldAlert,
  Crosshair,
  KeyRound,
  Wifi,
  Globe,
  Microscope,
  Eye,
};

export function CategoryCard({ category }: { category: Category }) {
  const Icon = iconMap[category.icon] || Radar;

  return (
    <a href={`/tools?category=${category.id}`} className={styles.card}>
      <div className={styles.iconWrap}>
        <Icon size={28} />
      </div>
      <h3 className={styles.name}>{category.name}</h3>
      <p className={styles.description}>{category.description}</p>
      <span className={styles.count}>{category.toolCount} 个工具</span>
      <div className={styles.glow} />
    </a>
  );
}

import Link from 'next/link';
import { Shield, Github, ExternalLink } from 'lucide-react';
import styles from './Footer.module.css';

export function Footer() {
  return (
    <footer className={styles.footer}>
      <div className={styles.divider} />
      <div className={styles.inner}>
        <div className={styles.ascii}>
          <pre className={styles.asciiArt}>{`
  +----------------------------------------+
  | CYBERTOOLKIT // 安全工具库              |
  | ====================================== |
  | [状态: 运行中]                          |
  | [模式: 防御]                            |
  +----------------------------------------+`}</pre>
        </div>

        <div className={styles.columns}>
          <div className={styles.column}>
            <div className={styles.brand}>
              <Shield size={20} className={styles.brandIcon} />
              <span className={styles.brandName}>CyberToolkit</span>
            </div>
            <p className={styles.desc}>为安全研究人员和爱好者整理的网络安全工具导航站。</p>
          </div>

          <div className={styles.column}>
            <h4 className={styles.colTitle}>&gt; 快速链接</h4>
            <Link href="/" className={styles.footerLink}>
              <ExternalLink size={12} /> 首页
            </Link>
            <Link href="/tools" className={styles.footerLink}>
              <ExternalLink size={12} /> 全部工具
            </Link>
          </div>

          <div className={styles.column}>
            <h4 className={styles.colTitle}>&gt; 资源</h4>
            <a href="https://github.com" target="_blank" rel="noopener noreferrer" className={styles.footerLink}>
              <Github size={12} /> GitHub
            </a>
            <a href="https://owasp.org" target="_blank" rel="noopener noreferrer" className={styles.footerLink}>
              <ExternalLink size={12} /> OWASP
            </a>
          </div>
        </div>

        <div className={styles.bottom}>
          <span className={styles.copy}>&copy; {new Date().getFullYear()} CyberToolkit. 保留所有权利。</span>
          <span className={styles.terminal}>
            root@cybertoolkit:~$ <span className={styles.blink}>█</span>
          </span>
        </div>
      </div>
    </footer>
  );
}

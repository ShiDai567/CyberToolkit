import { Shield, Github, ExternalLink } from 'lucide-react';
import styles from './Footer.module.css';

export function Footer() {
  return (
    <footer className={styles.footer}>
      <div className={styles.divider} />
      <div className={styles.inner}>
        <div className={styles.ascii}>
          <pre className={styles.asciiArt}>{`
  ╔══════════════════════════════════════╗
  ║  CYBERTOOLKIT // 安全武器库          ║
  ║  ================================   ║
  ║  [状态: 运行中]                      ║
  ║  [模式: 防御]                        ║
  ╚══════════════════════════════════════╝`}</pre>
        </div>

        <div className={styles.columns}>
          <div className={styles.column}>
            <div className={styles.brand}>
              <Shield size={20} className={styles.brandIcon} />
              <span className={styles.brandName}>CyberToolkit</span>
            </div>
            <p className={styles.desc}>
              为专业人员和爱好者精心策划的强大网络安全工具集合。
            </p>
          </div>

          <div className={styles.column}>
            <h4 className={styles.colTitle}>&gt; 快速链接</h4>
            <a href="/" className={styles.footerLink}>
              <ExternalLink size={12} /> 首页
            </a>
            <a href="/tools" className={styles.footerLink}>
              <ExternalLink size={12} /> 所有工具
            </a>
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
          <span className={styles.copy}>
            &copy; {new Date().getFullYear()} CyberToolkit. 保留所有权利。
          </span>
          <span className={styles.terminal}>
            root@cybertoolkit:~$ <span className={styles.blink}>█</span>
          </span>
        </div>
      </div>
    </footer>
  );
}

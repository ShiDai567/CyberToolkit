# CyberToolkit — 网络安全工具集

<p align="center">
  <strong>🛡️ 精心策划的网络安全工具集合平台</strong>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Next.js-16.2.1-black?logo=next.js" alt="Next.js">
  <img src="https://img.shields.io/badge/TypeScript-5.x-blue?logo=typescript" alt="TypeScript">
  <img src="https://img.shields.io/badge/Tailwind_CSS-4.x-06B6D4?logo=tailwindcss" alt="Tailwind CSS">
  <img src="https://img.shields.io/badge/License-MIT-green" alt="License">
</p>

---

## 📖 项目简介

CyberToolkit 是一个 **赛博朋克风格** 的网络安全工具集合网站，为安全研究人员、渗透测试人员和安全爱好者提供一站式工具索引平台。

### 特色

- **赛博朋克 UI** — 深色背景 + 霓虹绿配色 + 终端字体 + 矩阵雨动画
- **20+ 安全工具** — 覆盖 8 大安全领域
- **搜索与筛选** — 按名称、分类、难度等级快速定位
- **工具详情页** — 完整描述、标签、外部链接
- **响应式设计** — 支持桌面端和移动端

### 工具分类

| 分类 | 示例工具 |
|------|----------|
| 网络扫描 | Nmap, Masscan, RustScan |
| 漏洞评估 | OpenVAS, Nikto, Nuclei |
| 渗透测试 | Metasploit, Burp Suite, SQLMap |
| 密码破解 | John the Ripper, Hashcat |
| 无线安全 | Aircrack-ng, Wireshark |
| Web 安全 | OWASP ZAP, Gobuster |
| 数字取证 | Autopsy, Volatility |
| 开源情报 | Maltego, theHarvester, Shodan |

---

## 🚀 快速开始

### 环境要求

- **Node.js** >= 18.x
- **npm** >= 9.x

### 安装与运行

```bash
# 克隆项目
git clone <your-repo-url>
cd cybersecurity_tools

# 安装依赖
cd frontend
npm install

# 启动开发服务器
npm run dev
```

打开浏览器访问 [http://localhost:3000](http://localhost:3000)

### 生产构建

```bash
cd frontend
npm run build
npm start
```

---

## 🏗️ 技术栈

| 技术 | 用途 |
|------|------|
| [Next.js 16](https://nextjs.org) | React 框架 (App Router + Turbopack) |
| [TypeScript](https://typescriptlang.org) | 类型安全 |
| [Tailwind CSS 4](https://tailwindcss.com) | 实用工具样式 |
| [Lucide React](https://lucide.dev) | SVG 图标库 |
| CSS Modules | 组件级样式隔离 |

---

## 📁 项目结构

```
frontend/
├── src/
│   ├── app/                    # Next.js App Router 页面
│   │   ├── layout.tsx          # 根布局
│   │   ├── page.tsx            # 首页
│   │   ├── globals.css         # 全局样式 & 设计系统
│   │   └── tools/
│   │       ├── page.tsx        # 工具列表页
│   │       └── [id]/
│   │           └── page.tsx    # 工具详情页
│   ├── components/             # 可复用组件
│   │   ├── Navbar.tsx          # 导航栏
│   │   ├── Footer.tsx          # 页脚
│   │   ├── HeroSection.tsx     # 首页 Hero 区域
│   │   ├── ToolCard.tsx        # 工具卡片
│   │   ├── CategoryCard.tsx    # 分类卡片
│   │   ├── StatsBar.tsx        # 统计数据栏
│   │   └── SearchBar.tsx       # 搜索栏
│   ├── data/
│   │   └── tools.ts            # 工具数据集
│   └── types/
│       └── types.ts            # TypeScript 类型定义
├── next.config.ts              # Next.js 配置
├── package.json
└── tsconfig.json
```

---

## 🎨 设计系统

项目基于 **Cyberpunk UI** 设计风格：

| 设计元素 | 值 |
|---------|-----|
| 背景色 | `#0F172A` |
| 主色调 | `#1E293B` |
| 强调色 | `#22C55E` (霓虹绿) |
| 标题字体 | Share Tech Mono |
| 正文字体 | Fira Code |
| 视觉效果 | 霓虹发光、扫描线、矩阵雨、毛玻璃 |

---

## 📄 License

[MIT](LICENSE)

import { Tool, Category } from '@/types/types';

export const categories: Category[] = [
  {
    id: 'network-scanning',
    name: '网络扫描',
    description: '发现主机、端口、服务与基础网络暴露面',
    icon: 'Radar',
    toolCount: 3,
  },
  {
    id: 'vulnerability-assessment',
    name: '漏洞评估',
    description: '识别已知漏洞、配置缺陷与风险项',
    icon: 'ShieldAlert',
    toolCount: 3,
  },
  {
    id: 'penetration-testing',
    name: '渗透测试',
    description: '用于验证攻击链路与安全防护效果',
    icon: 'Crosshair',
    toolCount: 3,
  },
  {
    id: 'password-cracking',
    name: '密码破解',
    description: '评估口令强度与哈希恢复能力',
    icon: 'KeyRound',
    toolCount: 2,
  },
  {
    id: 'wireless-security',
    name: '无线安全',
    description: '面向 Wi-Fi 环境的审计、抓包与分析',
    icon: 'Wifi',
    toolCount: 2,
  },
  {
    id: 'web-security',
    name: 'Web 安全',
    description: '聚焦 Web 应用扫描、枚举与测试',
    icon: 'Globe',
    toolCount: 2,
  },
  {
    id: 'forensics',
    name: '数字取证',
    description: '面向磁盘、内存与证据提取分析',
    icon: 'Microscope',
    toolCount: 2,
  },
  {
    id: 'osint',
    name: '开源情报',
    description: '从公开数据源收集情报与关联信息',
    icon: 'Eye',
    toolCount: 3,
  },
];

export const tools: Tool[] = [
  {
    id: 'nmap',
    name: 'Nmap',
    description: '网络发现与安全审计的行业标准工具',
    longDescription:
      'Nmap 是经典的网络扫描工具，可用于主机发现、端口扫描、服务识别、操作系统探测和脚本化安全审计。它既适合基础信息收集，也适合日常渗透测试和资产盘点。',
    category: 'network-scanning',
    tags: ['端口扫描', '主机发现', '服务识别', '操作系统探测'],
    website: 'https://nmap.org',
    github: 'https://github.com/nmap/nmap',
    difficulty: 'intermediate',
    icon: 'Radar',
    featured: true,
  },
  {
    id: 'masscan',
    name: 'Masscan',
    description: '面向大规模资产探测的高速端口扫描器',
    longDescription:
      'Masscan 以极高吞吐量著称，适合大网段或互联网范围的快速端口发现。它更偏向先快扫再细查的工作流，常与 Nmap 等工具配合使用。',
    category: 'network-scanning',
    tags: ['高速扫描', '大规模探测', '端口发现'],
    website: 'https://github.com/robertdavidgraham/masscan',
    github: 'https://github.com/robertdavidgraham/masscan',
    difficulty: 'advanced',
    icon: 'Zap',
  },
  {
    id: 'rustscan',
    name: 'RustScan',
    description: '结合高速扫描与 Nmap 深度探测的现代工具',
    longDescription:
      'RustScan 用 Rust 编写，端口发现速度很快，并支持把结果直接交给 Nmap 做进一步服务识别，适合需要效率与细节兼顾的场景。',
    category: 'network-scanning',
    tags: ['Rust', '快速扫描', 'Nmap 集成'],
    website: 'https://rustscan.github.io/RustScan/',
    github: 'https://github.com/RustScan/RustScan',
    difficulty: 'beginner',
    icon: 'ScanLine',
  },
  {
    id: 'openvas',
    name: 'OpenVAS',
    description: '功能完整的开源漏洞扫描与评估平台',
    longDescription:
      'OpenVAS 提供较全面的漏洞检测能力，覆盖网络服务、系统配置与已知风险项，适合做周期性安全检查与基线评估。',
    category: 'vulnerability-assessment',
    tags: ['漏洞扫描', '合规检查', '安全评估'],
    website: 'https://www.openvas.org',
    github: 'https://github.com/greenbone/openvas-scanner',
    difficulty: 'intermediate',
    icon: 'ShieldAlert',
    featured: true,
  },
  {
    id: 'nikto',
    name: 'Nikto',
    description: '用于 Web 服务器风险检查的轻量级扫描器',
    longDescription:
      'Nikto 主要面向 Web 服务配置问题、危险文件、过期组件和常见弱点检查，适合初步侦察和快速基线审计。',
    category: 'vulnerability-assessment',
    tags: ['Web 扫描', '服务审计', '配置检查'],
    website: 'https://cirt.net/Nikto2',
    github: 'https://github.com/sullo/nikto',
    difficulty: 'beginner',
    icon: 'Search',
  },
  {
    id: 'nuclei',
    name: 'Nuclei',
    description: '基于模板的高效漏洞扫描与验证工具',
    longDescription:
      'Nuclei 通过 YAML 模板快速检测 CVE、错误配置、暴露面和指纹信息，适合批量资产巡检，也适合团队沉淀自定义规则。',
    category: 'vulnerability-assessment',
    tags: ['模板扫描', 'CVE', '自动化', 'YAML'],
    website: 'https://nuclei.projectdiscovery.io',
    github: 'https://github.com/projectdiscovery/nuclei',
    difficulty: 'intermediate',
    icon: 'Atom',
    featured: true,
  },
  {
    id: 'metasploit',
    name: 'Metasploit Framework',
    description: '被广泛使用的渗透测试与漏洞利用框架',
    longDescription:
      'Metasploit 提供漏洞利用、后渗透、载荷和辅助模块，适合验证漏洞可利用性、演练攻击链以及构建测试流程。',
    category: 'penetration-testing',
    tags: ['漏洞利用', '载荷', '后渗透', '框架'],
    website: 'https://www.metasploit.com',
    github: 'https://github.com/rapid7/metasploit-framework',
    difficulty: 'advanced',
    icon: 'Crosshair',
    featured: true,
  },
  {
    id: 'burpsuite',
    name: 'Burp Suite',
    description: 'Web 应用安全测试中的核心平台',
    longDescription:
      'Burp Suite 集成代理、重放、扫描和流量分析能力，适合手工测试与自动化扫描结合的 Web 安全工作流。',
    category: 'penetration-testing',
    tags: ['Web 测试', '代理', '扫描', '流量分析'],
    website: 'https://portswigger.net/burp',
    difficulty: 'intermediate',
    icon: 'Bug',
    featured: true,
  },
  {
    id: 'sqlmap',
    name: 'SQLMap',
    description: '自动化 SQL 注入检测与利用工具',
    longDescription:
      'SQLMap 用于检测 SQL 注入、识别数据库类型并自动化执行数据提取等操作，是 Web 渗透测试中非常常见的专用工具。',
    category: 'penetration-testing',
    tags: ['SQL 注入', '数据库', '自动化', '漏洞利用'],
    website: 'https://sqlmap.org',
    github: 'https://github.com/sqlmapproject/sqlmap',
    difficulty: 'intermediate',
    icon: 'Database',
  },
  {
    id: 'john-the-ripper',
    name: 'John the Ripper',
    description: '支持多种哈希格式的密码恢复工具',
    longDescription:
      'John the Ripper 适合做口令强度评估和哈希破解测试，支持字典、规则和暴力等多种方式，是经典密码审计工具之一。',
    category: 'password-cracking',
    tags: ['哈希破解', '字典攻击', '规则攻击', '密码审计'],
    website: 'https://www.openwall.com/john/',
    github: 'https://github.com/openwall/john',
    difficulty: 'intermediate',
    icon: 'KeyRound',
    featured: true,
  },
  {
    id: 'hashcat',
    name: 'Hashcat',
    description: '面向高性能密码恢复的主流工具',
    longDescription:
      'Hashcat 利用 GPU 提供极高的破解性能，支持大量哈希算法与攻击模式，适合需要性能和可扩展性的口令审计场景。',
    category: 'password-cracking',
    tags: ['GPU 破解', '哈希恢复', '规则攻击', '高性能'],
    website: 'https://hashcat.net/hashcat/',
    github: 'https://github.com/hashcat/hashcat',
    difficulty: 'advanced',
    icon: 'Cpu',
  },
  {
    id: 'aircrack-ng',
    name: 'Aircrack-ng',
    description: '用于 Wi-Fi 审计、抓包和密钥分析的工具套件',
    longDescription:
      'Aircrack-ng 聚焦无线网络安全，覆盖监听、注入、抓包和密钥恢复等能力，适合无线环境测试与协议研究。',
    category: 'wireless-security',
    tags: ['Wi-Fi', 'WPA', '抓包', '无线审计'],
    website: 'https://www.aircrack-ng.org',
    github: 'https://github.com/aircrack-ng/aircrack-ng',
    difficulty: 'advanced',
    icon: 'Wifi',
  },
  {
    id: 'wireshark',
    name: 'Wireshark',
    description: '最常用的网络协议分析工具之一',
    longDescription:
      'Wireshark 支持海量协议解析，可用于实时抓包、离线分析与故障排查，是网络安全与协议研究中的基础工具。',
    category: 'wireless-security',
    tags: ['协议分析', '抓包', '流量分析', '网络取证'],
    website: 'https://www.wireshark.org',
    github: 'https://gitlab.com/wireshark/wireshark',
    difficulty: 'intermediate',
    icon: 'Activity',
    featured: true,
  },
  {
    id: 'owasp-zap',
    name: 'OWASP ZAP',
    description: '开源的 Web 应用安全扫描平台',
    longDescription:
      'OWASP ZAP 提供自动化扫描、代理和手工测试能力，适合开发阶段安全检查、教学演示以及常规 Web 漏洞发现。',
    category: 'web-security',
    tags: ['Web 扫描', '代理', '自动化测试', 'OWASP'],
    website: 'https://www.zaproxy.org',
    github: 'https://github.com/zaproxy/zaproxy',
    difficulty: 'beginner',
    icon: 'Globe',
  },
  {
    id: 'dirbuster',
    name: 'Gobuster',
    description: '适合目录、子域和虚拟主机枚举的 Go 工具',
    longDescription:
      'Gobuster 以简单、快速著称，常用于目录爆破、DNS 枚举和虚拟主机探测，是 Web 侦察阶段的常见工具。',
    category: 'web-security',
    tags: ['目录枚举', 'DNS', '子域发现'],
    website: 'https://github.com/OJ/gobuster',
    github: 'https://github.com/OJ/gobuster',
    difficulty: 'beginner',
    icon: 'FolderSearch',
  },
  {
    id: 'autopsy',
    name: 'Autopsy',
    description: '面向磁盘与终端介质分析的数字取证平台',
    longDescription:
      'Autopsy 是图形化数字取证平台，适合做文件恢复、时间线分析、关键证据提取和案例整理。',
    category: 'forensics',
    tags: ['磁盘取证', '文件恢复', '证据分析', '时间线'],
    website: 'https://www.autopsy.com',
    github: 'https://github.com/sleuthkit/autopsy',
    difficulty: 'intermediate',
    icon: 'Microscope',
  },
  {
    id: 'volatility',
    name: 'Volatility',
    description: '面向内存样本分析的高级取证框架',
    longDescription:
      'Volatility 主要用于 RAM 镜像分析，可提取进程、网络连接、注入痕迹和恶意行为线索，是应急响应中的关键工具。',
    category: 'forensics',
    tags: ['内存取证', '应急响应', '恶意代码分析', 'RAM'],
    website: 'https://www.volatilityfoundation.org',
    github: 'https://github.com/volatilityfoundation/volatility3',
    difficulty: 'expert',
    icon: 'BrainCircuit',
  },
  {
    id: 'maltego',
    name: 'Maltego',
    description: '用于实体关联分析与可视化的 OSINT 工具',
    longDescription:
      'Maltego 擅长把人、域名、组织、邮箱等公开信息建立关联图谱，适合调查分析、关系梳理和情报可视化。',
    category: 'osint',
    tags: ['关联分析', '可视化', '情报收集', '图谱'],
    website: 'https://www.maltego.com',
    difficulty: 'advanced',
    icon: 'Network',
  },
  {
    id: 'theharvester',
    name: 'theHarvester',
    description: '用于邮箱、子域与人员信息收集的侦察工具',
    longDescription:
      'theHarvester 从多个公开来源聚合邮箱、域名、子域和组织信息，适合前期信息收集与攻击面梳理。',
    category: 'osint',
    tags: ['邮箱收集', '子域枚举', '侦察', '公开信息'],
    website: 'https://github.com/laramies/theHarvester',
    github: 'https://github.com/laramies/theHarvester',
    difficulty: 'beginner',
    icon: 'Wheat',
  },
  {
    id: 'shodan',
    name: 'Shodan',
    description: '面向互联网暴露设备与服务的搜索引擎',
    longDescription:
      'Shodan 可以搜索互联网可见设备、服务和指纹信息，适合做资产暴露面分析、快速侦察和风险排查。',
    category: 'osint',
    tags: ['设备搜索', '资产暴露', '服务发现', '互联网侦察'],
    website: 'https://www.shodan.io',
    difficulty: 'beginner',
    icon: 'Eye',
    featured: true,
  },
];

export const featuredTools = tools.filter((t) => t.featured);

export function getToolsByCategory(categoryId: string): Tool[] {
  return tools.filter((t) => t.category === categoryId);
}

export function getToolById(id: string): Tool | undefined {
  return tools.find((t) => t.id === id);
}

export function searchTools(query: string): Tool[] {
  const q = query.toLowerCase();
  return tools.filter(
    (t) =>
      t.name.toLowerCase().includes(q) ||
      t.description.toLowerCase().includes(q) ||
      t.tags.some((tag) => tag.toLowerCase().includes(q)) ||
      t.category.toLowerCase().includes(q)
  );
}

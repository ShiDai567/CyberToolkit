import type { Metadata } from 'next';
import './globals.css';
import { Navbar } from '@/components/Navbar';
import { AuthProvider } from '@/components/AuthProvider';

export const metadata: Metadata = {
  title: 'CyberToolkit - 网络安全工具集',
  description:
    '探索精心整理的网络安全工具集合，覆盖网络扫描、渗透测试、数字取证与开源情报等方向。',
  keywords: ['网络安全', '安全工具', '渗透测试', '网络扫描', '数字取证', '开源情报'],
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="zh-CN" data-scroll-behavior="smooth">
      <body className="scanline-overlay grid-bg">
        <AuthProvider>
          <Navbar />
          <main style={{ paddingTop: '80px' }}>{children}</main>
        </AuthProvider>
      </body>
    </html>
  );
}

import type { Metadata } from 'next';
import { Toaster } from 'sonner';
import './globals.css';
import { Navbar } from '@/components/Navbar';
import { AuthProvider } from '@/components/AuthProvider';
import NavbarWrapper from '@/components/NavbarWrapper';

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
    <html lang="zh-CN" data-scroll-behavior="smooth" suppressHydrationWarning>
      <body className="scanline-overlay grid-bg">
        <AuthProvider>
          <NavbarWrapper />
          {children}
          <Toaster
            position="top-center"
            richColors
            closeButton={false}
            theme="dark"
          />
        </AuthProvider>
      </body>
    </html>
  );
}

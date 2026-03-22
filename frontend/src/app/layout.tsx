import type { Metadata } from "next";
import "./globals.css";
import { Navbar } from "@/components/Navbar";
import { Footer } from "@/components/Footer";

export const metadata: Metadata = {
  title: "CyberToolkit — 网络安全工具集",
  description:
    "探索精心策划的强大网络安全工具集合。涵盖网络扫描、渗透测试、数字取证和开源情报等领域。",
  keywords: [
    "网络安全",
    "安全工具",
    "渗透测试",
    "网络扫描",
    "黑客工具",
    "信息安全",
  ],
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="zh-CN" data-scroll-behavior="smooth">
      <body className="scanline-overlay grid-bg">
        <Navbar />
        <main style={{ minHeight: "100vh", paddingTop: "80px" }}>
          {children}
        </main>
        <Footer />
      </body>
    </html>
  );
}

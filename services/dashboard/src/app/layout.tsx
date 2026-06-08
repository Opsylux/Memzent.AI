import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";


const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: {
    default: "Memzent — AI Agent Memory & Semantic Proxy",
    template: "%s | Memzent",
  },
  description: "The intelligent semantic proxy for AI agents. Intercept and optimize LLM traffic with entity-aware caching, semantic routing, RBAC, and GPU avoidance — saving 80%+ on LLM costs.",
  keywords: ["AI agent memory", "semantic cache", "LLM proxy", "MCP tools", "GPU avoidance", "entity extraction", "RBAC", "multi-tenant AI", "agentic infrastructure"],
  metadataBase: new URL("https://app.memzent.ai"),
  openGraph: {
    type: "website",
    siteName: "Memzent",
    title: "Memzent — AI Agent Memory & Semantic Proxy",
    description: "Entity-aware semantic caching, multi-LLM routing, and RBAC for autonomous AI agents. Pay-as-you-go with 80%+ cost savings.",
    url: "https://app.memzent.ai",
  },
  twitter: {
    card: "summary_large_image",
    title: "Memzent — AI Agent Memory Layer",
    description: "Semantic caching, RBAC, and multi-LLM routing for autonomous AI agents.",
  },
  robots: {
    index: true,
    follow: true,
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className="dark" suppressHydrationWarning>
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased min-h-screen bg-memzent-dark text-white selection:bg-memzent-glow selection:text-black`}
        suppressHydrationWarning
      >
        {children}
      </body>
    </html>
  );
}

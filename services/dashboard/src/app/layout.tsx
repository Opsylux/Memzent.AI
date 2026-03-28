import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";
import { AuraSidebar } from "@/components/aura-sidebar";
import { AuraTopNav } from "@/components/aura-top-nav";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "Aura Command Center",
  description: "Enterprise-grade observability for AI agentic infrastructure.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className="dark">
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased min-h-screen bg-aura-dark text-white`}
      >
        <div className="flex min-h-screen">
          <AuraSidebar />
          <div className="flex-1 ml-[320px] transition-all duration-300">
             <AuraTopNav />
             <main className="mt-40 p-10 max-w-7xl mx-auto min-h-screen">
              {children}
             </main>
          </div>
        </div>
      </body>
    </html>
  );
}

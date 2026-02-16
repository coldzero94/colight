import type { Metadata } from "next";
import { Inter, Geist } from "next/font/google";
import { Analytics } from "@vercel/analytics/react";
import { SpeedInsights } from "@vercel/speed-insights/next";
import { Toaster } from "sonner";
import { Providers } from "@/components/providers";
import "./globals.css";

const inter = Inter({
  subsets: ["latin"],
  display: "swap",
  variable: "--font-inter",
});

const geist = Geist({
  subsets: ["latin"],
  display: "swap",
  variable: "--font-geist",
});

export const metadata: Metadata = {
  title: "Colight",
  description: "AI 자소서 코칭 플랫폼",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="ko" className={`${inter.variable} ${geist.variable}`}>
      <body className="font-sans">
        <Providers>
          {children}
          <Toaster position="top-center" richColors />
        </Providers>
        <Analytics />
        <SpeedInsights />
      </body>
    </html>
  );
}

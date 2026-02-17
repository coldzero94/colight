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

const siteUrl =
  process.env.NEXT_PUBLIC_SITE_URL ??
  process.env.NEXT_PUBLIC_APP_URL ??
  "https://colight.vercel.app";

export const metadata: Metadata = {
  metadataBase: new URL(siteUrl),
  title: {
    default: "Colight",
    template: "%s | Colight",
  },
  description:
    "Colight helps you turn experiences into stronger stories with AI-powered resume coaching.",
  openGraph: {
    type: "website",
    url: "/",
    siteName: "Colight",
    title: "Colight | Together We Light",
    description:
      "AI-powered resume coaching platform for better applications, interviews, and growth.",
    images: [
      {
        url: "/opengraph-image",
        width: 1200,
        height: 630,
        alt: "Colight social preview card",
      },
    ],
  },
  twitter: {
    card: "summary_large_image",
    title: "Colight | Together We Light",
    description:
      "AI-powered resume coaching platform for better applications, interviews, and growth.",
    images: ["/twitter-image"],
  },
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

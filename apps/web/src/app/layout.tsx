import type { Metadata } from "next";
import "./globals.css";

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
    <html lang="ko">
      <body>{children}</body>
    </html>
  );
}

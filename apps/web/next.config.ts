import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  reactStrictMode: true,
  experimental: {
    optimizePackageImports: [
      "@tanstack/react-query",
      "sonner",
      "react-hook-form",
      "@hookform/resolvers",
    ],
  },
};

export default nextConfig;

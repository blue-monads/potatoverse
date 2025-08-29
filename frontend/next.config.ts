import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  reactStrictMode: true,
  basePath: "/z/pages",
  distDir: "output/build",
  images: {
    unoptimized: true,
  },
};

export default nextConfig;

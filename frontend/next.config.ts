import type { NextConfig } from "next";

const isProduction = process.env.NODE_ENV === "production";

const nextConfig: NextConfig = {
  reactStrictMode: true,
  basePath: "/z/pages",
  output: "export",
  distDir: isProduction ? "output/build" : ".next",
  images: {
    unoptimized: true,
  },
};

export default nextConfig;

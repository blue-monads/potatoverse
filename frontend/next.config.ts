import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  reactStrictMode: true,
  basePath: "/z/pages",
  output: "export",
  distDir: "output/build",
};

export default nextConfig;

import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  allowedDevOrigins: ["127.0.0.1", "localhost"],
  async rewrites() {
    return [
      {
        source: "/health",
        destination: "http://127.0.0.1:43148/health",
      },
      {
        source: "/export/:id",
        destination: "http://127.0.0.1:43148/export/:id",
      },
    ];
  },
};

export default nextConfig;

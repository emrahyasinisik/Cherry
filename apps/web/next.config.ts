import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  allowedDevOrigins: ["127.0.0.1", "localhost"],
  async rewrites() {
    return [
      {
        source: "/graphql",
        destination: "http://127.0.0.1:43148/graphql",
      },
      {
        source: "/health",
        destination: "http://127.0.0.1:43148/health",
      },
    ];
  },
};

export default nextConfig;

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
        source: "/graphql",
        destination: "http://127.0.0.1:43148/graphql",
      },
      {
        source: "/export/:id",
        destination: "http://127.0.0.1:43148/export/:id",
      },
      {
        source: "/colab/:path*",
        destination: "http://127.0.0.1:43148/colab/:path*",
      },
      {
        source: "/oauth/decision",
        destination: "http://127.0.0.1:43148/oauth/decision",
      },
      {
        source: "/oauth/provider/callback",
        destination: "http://127.0.0.1:43148/oauth/provider/callback",
      },
    ];
  },
};

export default nextConfig;

import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "standalone",
  async rewrites() {
    return [
      {
        source: "/api/users/:path*",
        destination: "http://localhost:8081/api/users/:path*",
      },
      {
        source: "/api/teams/:path*",
        destination: "http://localhost:8081/api/teams/:path*",
      },
      {
        source: "/api/projects/:path*",
        destination: "http://localhost:8082/api/projects/:path*",
      },
    ];
  },
};

export default nextConfig;

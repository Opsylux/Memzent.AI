import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  outputFileTracingIncludes: {
    "/blog": ["./src/content/blog/**/*"],
    "/blog/[slug]": ["./src/content/blog/**/*"],
  },
};

export default nextConfig;

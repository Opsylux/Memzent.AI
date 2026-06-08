import type { MetadataRoute } from "next";

export default function robots(): MetadataRoute.Robots {
  return {
    rules: [
      {
        userAgent: "*",
        allow: ["/docs", "/blog", "/login"],
        disallow: ["/api/", "/admin/", "/settings", "/keys", "/billing", "/audit", "/playground"],
      },
    ],
    sitemap: "https://app.memzent.ai/sitemap.xml",
  };
}

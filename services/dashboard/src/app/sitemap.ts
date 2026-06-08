import type { MetadataRoute } from "next";
import { getAllPosts } from "@/lib/blog";

const BASE_URL = "https://app.memzent.ai";

const docPages = [
  "",
  "/architecture",
  "/semantic-proxy",
  "/caching",
  "/quickstart",
  "/auth",
  "/first-request",
  "/api-reference",
  "/providers",
  "/sessions",
  "/webhooks",
  "/errors",
  "/rbac",
  "/permissions",
  "/tool-registry",
  "/spend-limits",
  "/entity-extraction",
  "/cache-layers",
  "/offline-learning",
  "/gpu-analytics",
];

export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
  const staticPages: MetadataRoute.Sitemap = [
    { url: BASE_URL, lastModified: new Date(), changeFrequency: "weekly", priority: 1.0 },
    { url: `${BASE_URL}/login`, lastModified: new Date(), changeFrequency: "monthly", priority: 0.3 },
    { url: `${BASE_URL}/blog`, lastModified: new Date(), changeFrequency: "weekly", priority: 0.8 },
  ];

  // Docs pages
  const docsEntries: MetadataRoute.Sitemap = docPages.map((page) => ({
    url: `${BASE_URL}/docs${page}`,
    lastModified: new Date(),
    changeFrequency: "monthly" as const,
    priority: 0.7,
  }));

  // Blog posts
  let blogEntries: MetadataRoute.Sitemap = [];
  try {
    const posts = await getAllPosts();
    blogEntries = posts.map((post) => ({
      url: `${BASE_URL}/blog/${post.slug}`,
      lastModified: new Date(post.date),
      changeFrequency: "monthly" as const,
      priority: 0.6,
    }));
  } catch {
    // Blog posts may not be available during build
  }

  return [...staticPages, ...docsEntries, ...blogEntries];
}

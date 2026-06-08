import fs from "fs";
import path from "path";
import { createClient } from "@/lib/supabase-server";
import type { BlogPost } from "@/lib/blog-types";

// Try multiple possible content directory paths
function getContentDir(): string {
  const candidates = [
    path.join(process.cwd(), "src/content/blog"),
    path.join(__dirname, "../../content/blog"),
    path.join(process.cwd(), "services/dashboard/src/content/blog"),
  ];
  for (const dir of candidates) {
    if (fs.existsSync(dir)) return dir;
  }
  return candidates[0];
}

function estimateReadingTime(content: string): number {
  const words = content.split(/\s+/).length;
  return Math.max(1, Math.ceil(words / 200));
}

/**
 * Load blog posts from MDX/markdown files in src/content/blog/
 */
function getFilePosts(): BlogPost[] {
  const contentDir = getContentDir();
  if (!fs.existsSync(contentDir)) return [];

  try {
    const files = fs.readdirSync(contentDir).filter((f) => f.endsWith(".md") || f.endsWith(".mdx"));
    const posts: BlogPost[] = [];

    for (const file of files) {
      const raw = fs.readFileSync(path.join(contentDir, file), "utf-8");
      const { meta, content } = parseFrontmatter(raw);
      if (!meta.title || !meta.published_at) continue;

      posts.push({
        slug: file.replace(/\.(md|mdx)$/, ""),
        title: meta.title,
        description: meta.description || "",
        content,
        author: meta.author || "Memzent Team",
        author_avatar: meta.author_avatar,
        cover_image: meta.cover_image,
        category: meta.category || "engineering",
        tags: meta.tags ? meta.tags.split(",").map((t: string) => t.trim()) : [],
        published_at: meta.published_at,
        updated_at: meta.updated_at,
        reading_time: estimateReadingTime(content),
        source: "mdx",
      });
    }

    return posts;
  } catch {
    return [];
  }
}

/**
 * Load blog posts from Supabase blog_posts table
 */
async function getDBPosts(): Promise<BlogPost[]> {
  try {
    const supabase = await createClient();
    const { data, error } = await supabase
      .from("blog_posts")
      .select("*")
      .eq("published", true)
      .order("published_at", { ascending: false });

    if (error || !data) return [];

    return data.map((row: any) => ({
      slug: row.slug,
      title: row.title,
      description: row.description || "",
      content: row.content,
      author: row.author || "Memzent Team",
      author_avatar: row.author_avatar,
      cover_image: row.cover_image,
      category: row.category || "blog",
      tags: row.tags || [],
      published_at: row.published_at,
      updated_at: row.updated_at,
      reading_time: estimateReadingTime(row.content || ""),
      source: "database",
    }));
  } catch {
    return [];
  }
}

/**
 * Get all blog posts from both sources, sorted by date
 */
export async function getAllPosts(): Promise<BlogPost[]> {
  const [filePosts, dbPosts] = await Promise.all([
    Promise.resolve(getFilePosts()),
    getDBPosts(),
  ]);

  const all = [...filePosts, ...dbPosts];
  all.sort((a, b) => new Date(b.published_at).getTime() - new Date(a.published_at).getTime());
  return all;
}

/**
 * Get a single post by slug
 */
export async function getPostBySlug(slug: string): Promise<BlogPost | null> {
  const posts = await getAllPosts();
  return posts.find((p) => p.slug === slug) || null;
}

/**
 * Simple frontmatter parser (no external deps)
 */
function parseFrontmatter(raw: string): { meta: Record<string, any>; content: string } {
  const meta: Record<string, any> = {};
  if (!raw.startsWith("---")) return { meta, content: raw };

  const end = raw.indexOf("---", 3);
  if (end === -1) return { meta, content: raw };

  const frontmatter = raw.slice(3, end).trim();
  const content = raw.slice(end + 3).trim();

  for (const line of frontmatter.split("\n")) {
    const idx = line.indexOf(":");
    if (idx === -1) continue;
    const key = line.slice(0, idx).trim();
    const value = line.slice(idx + 1).trim().replace(/^["']|["']$/g, "");
    meta[key] = value;
  }

  return { meta, content };
}

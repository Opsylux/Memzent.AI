export interface BlogPost {
  slug: string;
  title: string;
  description: string;
  content: string;
  author: string;
  author_avatar?: string;
  cover_image?: string;
  category: "engineering" | "use-case" | "announcement" | "tutorial";
  tags: string[];
  published_at: string;
  updated_at?: string;
  reading_time?: number;
  source: "mdx" | "database";
}

export const BLOG_CATEGORIES = {
  engineering: { label: "Engineering", color: "text-blue-400", bg: "bg-blue-500/10" },
  "use-case": { label: "Use Case", color: "text-green-400", bg: "bg-green-500/10" },
  announcement: { label: "Announcement", color: "text-purple-400", bg: "bg-purple-500/10" },
  tutorial: { label: "Tutorial", color: "text-memzent-glow", bg: "bg-memzent-glow/10" },
} as const;

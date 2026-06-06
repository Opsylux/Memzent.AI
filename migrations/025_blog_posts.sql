-- Blog posts table for the Memzent dashboard blog
-- Supports hybrid content: MDX files in repo + database-managed posts

CREATE TABLE IF NOT EXISTS blog_posts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  slug TEXT UNIQUE NOT NULL,
  title TEXT NOT NULL,
  description TEXT,
  content TEXT NOT NULL DEFAULT '',
  author TEXT NOT NULL DEFAULT 'Memzent Team',
  author_avatar TEXT,
  cover_image TEXT,
  category TEXT NOT NULL DEFAULT 'engineering' CHECK (category IN ('engineering', 'use-case', 'announcement', 'tutorial')),
  tags TEXT[] DEFAULT '{}',
  published BOOLEAN NOT NULL DEFAULT false,
  published_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Index for fast listing queries
CREATE INDEX IF NOT EXISTS idx_blog_posts_published ON blog_posts (published, published_at DESC);
CREATE INDEX IF NOT EXISTS idx_blog_posts_category ON blog_posts (category) WHERE published = true;
CREATE INDEX IF NOT EXISTS idx_blog_posts_slug ON blog_posts (slug);

-- RLS: Public read for published posts, write restricted to authenticated users
ALTER TABLE blog_posts ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Public can read published blog posts"
  ON blog_posts FOR SELECT
  USING (published = true);

CREATE POLICY "Authenticated users can manage blog posts"
  ON blog_posts FOR ALL
  USING (auth.role() = 'authenticated');

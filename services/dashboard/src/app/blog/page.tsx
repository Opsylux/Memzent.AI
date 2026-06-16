"use client";

import { useEffect, useState } from "react";
import { supabase } from "@/lib/supabase";
import { BLOG_CATEGORIES } from "@/lib/blog-types";
import Link from "next/link";
import { Plus, Pencil, Trash2, Eye, EyeOff, ExternalLink, Search } from "lucide-react";

interface BlogPostRow {
  id: string;
  slug: string;
  title: string;
  description: string;
  category: string;
  published: boolean;
  published_at: string | null;
  updated_at: string;
  author: string;
  cover_image: string | null;
}

export default function BlogPage() {
  const [posts, setPosts] = useState<BlogPostRow[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [filter, setFilter] = useState<"all" | "published" | "draft">("all");
  const [deleting, setDeleting] = useState<string | null>(null);

  useEffect(() => {
    const load = async () => {
      setLoading(true);
      let query = supabase
        .from("blog_posts")
        .select("id, slug, title, description, category, published, published_at, updated_at, author, cover_image")
        .order("updated_at", { ascending: false });

      if (filter === "published") query = query.eq("published", true);
      if (filter === "draft") query = query.eq("published", false);

      const { data } = await query;
      setPosts(data || []);
      setLoading(false);
    };
    load();
  }, [filter]);

  const filtered = posts.filter((p) =>
    p.title.toLowerCase().includes(search.toLowerCase()) ||
    p.slug.toLowerCase().includes(search.toLowerCase())
  );

  const handleDelete = async (slug: string) => {
    if (!confirm(`Delete "${slug}" permanently?`)) return;
    setDeleting(slug);
    await supabase.from("blog_posts").delete().eq("slug", slug);
    setPosts((prev) => prev.filter((p) => p.slug !== slug));
    setDeleting(null);
  };

  const togglePublish = async (post: BlogPostRow) => {
    const newState = !post.published;
    await supabase.from("blog_posts").update({
      published: newState,
      published_at: newState ? new Date().toISOString() : post.published_at,
      updated_at: new Date().toISOString(),
    }).eq("slug", post.slug);
    setPosts((prev) =>
      prev.map((p) => p.slug === post.slug ? { ...p, published: newState, updated_at: new Date().toISOString() } : p)
    );
  };

  return (
    <div className="min-h-screen bg-[#030507]">
      <div className="border-b border-white/5 bg-black/40">
        <div className="max-w-6xl mx-auto px-6 py-6 flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-black tracking-tight">Blog Posts</h1>
            <p className="text-xs text-white/40 mt-1">
              {posts.length} post{posts.length !== 1 ? "s" : ""} · {posts.filter((p) => p.published).length} published · {posts.filter((p) => !p.published).length} draft{posts.filter((p) => !p.published).length !== 1 ? "s" : ""}
            </p>
          </div>
          <Link
            href="/blog/admin"
            className="flex items-center gap-2 px-4 py-2 rounded-lg text-xs font-black bg-memzent-glow text-black hover:scale-105 transition-all"
          >
            <Plus size={14} />
            New Post
          </Link>
        </div>
      </div>

      <div className="max-w-6xl mx-auto px-6 py-6 space-y-4">
        {/* Filters */}
        <div className="flex items-center gap-4">
          <div className="relative flex-1 max-w-sm">
            <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-white/20" />
            <input
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Search posts..."
              className="w-full pl-9 pr-4 py-2 rounded-lg bg-white/[0.03] border border-white/10 text-sm text-white placeholder:text-white/20 focus:outline-none focus:border-memzent-glow/40"
            />
          </div>
          <div className="flex gap-1">
            {(["all", "published", "draft"] as const).map((f) => (
              <button
                key={f}
                onClick={() => setFilter(f)}
                className={`px-3 py-1.5 rounded-lg text-[10px] font-black uppercase tracking-widest transition-all ${
                  filter === f
                    ? "bg-memzent-glow/10 text-memzent-glow border border-memzent-glow/20"
                    : "text-white/40 border border-white/5 hover:border-white/10"
                }`}
              >
                {f}
              </button>
            ))}
          </div>
        </div>

        {/* Post List */}
        {loading ? (
          <div className="text-center py-20 text-white/30 text-sm">Loading posts...</div>
        ) : filtered.length === 0 ? (
          <div className="text-center py-20">
            <p className="text-white/30 text-sm">No posts found.</p>
            <Link href="/blog/admin" className="text-memzent-glow text-xs font-bold hover:underline mt-2 inline-block">
              Create your first post →
            </Link>
          </div>
        ) : (
          <div className="space-y-2">
            {filtered.map((post) => {
              const cat = BLOG_CATEGORIES[post.category] || BLOG_CATEGORIES.engineering;
              return (
                <div
                  key={post.slug}
                  className="flex items-center gap-4 p-4 rounded-xl bg-white/[0.02] border border-white/5 hover:border-white/10 transition-all group"
                >
                  {/* Cover thumbnail */}
                  <div className="w-16 h-16 rounded-lg bg-white/5 flex-shrink-0 overflow-hidden">
                    {post.cover_image ? (
                      <img src={post.cover_image} alt="" className="w-full h-full object-cover" />
                    ) : (
                      <div className="w-full h-full flex items-center justify-center text-white/10 text-[10px] font-bold">
                        NO IMG
                      </div>
                    )}
                  </div>

                  {/* Info */}
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-1">
                      <span className={`text-[9px] font-black uppercase tracking-widest px-1.5 py-0.5 rounded ${cat.bg} ${cat.color}`}>
                        {cat.label}
                      </span>
                      {post.published ? (
                        <span className="text-[9px] font-bold text-green-400/60 flex items-center gap-1">
                          <Eye size={8} /> Published
                        </span>
                      ) : (
                        <span className="text-[9px] font-bold text-yellow-400/60 flex items-center gap-1">
                          <EyeOff size={8} /> Draft
                        </span>
                      )}
                    </div>
                    <h3 className="text-sm font-bold truncate">{post.title}</h3>
                    <p className="text-[10px] text-white/30 truncate">{post.description || "No description"}</p>
                    <p className="text-[9px] text-white/20 mt-1">
                      {post.author} · Updated {new Date(post.updated_at).toLocaleDateString()}
                    </p>
                  </div>

                  {/* Actions */}
                  <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                    <Link
                      href={`/blog/admin?slug=${post.slug}`}
                      className="p-2 rounded-lg hover:bg-white/5 text-white/40 hover:text-white transition-all"
                      title="Edit"
                    >
                      <Pencil size={14} />
                    </Link>
                    <button
                      onClick={() => togglePublish(post)}
                      className="p-2 rounded-lg hover:bg-white/5 text-white/40 hover:text-white transition-all"
                      title={post.published ? "Unpublish" : "Publish"}
                    >
                      {post.published ? <EyeOff size={14} /> : <Eye size={14} />}
                    </button>
                    <a
                      href={`https://memzent.ai/blog/${post.slug}`}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="p-2 rounded-lg hover:bg-white/5 text-white/40 hover:text-white transition-all"
                      title="View on site"
                    >
                      <ExternalLink size={14} />
                    </a>
                    <button
                      onClick={() => handleDelete(post.slug)}
                      disabled={deleting === post.slug}
                      className="p-2 rounded-lg hover:bg-red-500/10 text-white/40 hover:text-red-400 transition-all disabled:opacity-50"
                      title="Delete"
                    >
                      <Trash2 size={14} />
                    </button>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}

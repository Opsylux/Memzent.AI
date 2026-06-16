"use client";

import { useState } from "react";
import { supabase } from "@/lib/supabase";
import { BLOG_CATEGORIES } from "@/lib/blog-types";
import { Save, Eye, ArrowLeft } from "lucide-react";
import Link from "next/link";

export default function BlogAdminPage() {
  const [title, setTitle] = useState("");
  const [slug, setSlug] = useState("");
  const [description, setDescription] = useState("");
  const [content, setContent] = useState("");
  const [category, setCategory] = useState<string>("engineering");
  const [tags, setTags] = useState("");
  const [author, setAuthor] = useState("Memzent Team");
  const [coverImage, setCoverImage] = useState("");
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);
  const [error, setError] = useState("");

  const generateSlug = (text: string) => {
    return text
      .toLowerCase()
      .replace(/[^a-z0-9]+/g, "-")
      .replace(/^-|-$/g, "")
      .slice(0, 128);
  };

  const safeBlogHref = (s: string): string => {
    const sanitized = generateSlug(s);
    return sanitized ? `/blog/${sanitized}` : "#";
  };

  const handleTitleChange = (val: string) => {
    setTitle(val);
    if (!slug || slug === generateSlug(title)) {
      setSlug(generateSlug(val));
    }
  };

  const handlePublish = async () => {
    if (!title || !slug || !content) {
      setError("Title, slug, and content are required.");
      return;
    }

    setSaving(true);
    setError("");
    setSaved(false);

    const { error: dbError } = await supabase.from("blog_posts").upsert(
      {
        slug,
        title,
        description,
        content,
        category,
        tags: tags.split(",").map((t) => t.trim()).filter(Boolean),
        author,
        cover_image: coverImage || null,
        published: true,
        published_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      },
      { onConflict: "slug" }
    );

    setSaving(false);

    if (dbError) {
      setError(dbError.message);
    } else {
      setSaved(true);
      setTimeout(() => setSaved(false), 3000);
    }
  };

  const handleSaveDraft = async () => {
    if (!title || !slug) {
      setError("Title and slug are required.");
      return;
    }

    setSaving(true);
    setError("");

    const { error: dbError } = await supabase.from("blog_posts").upsert(
      {
        slug,
        title,
        description,
        content,
        category,
        tags: tags.split(",").map((t) => t.trim()).filter(Boolean),
        author,
        cover_image: coverImage || null,
        published: false,
        updated_at: new Date().toISOString(),
      },
      { onConflict: "slug" }
    );

    setSaving(false);
    if (dbError) setError(dbError.message);
    else { setSaved(true); setTimeout(() => setSaved(false), 3000); }
  };

  return (
    <div className="min-h-screen bg-[#030507]">
      <div className="border-b border-white/5 bg-black/40">
        <div className="max-w-4xl mx-auto px-6 py-4 flex items-center justify-between">
          <Link href="/blog" className="flex items-center gap-2 text-xs text-white/40 hover:text-memzent-glow transition-colors font-bold">
            <ArrowLeft size={12} />
            Back to Blog
          </Link>
          <div className="flex items-center gap-2">
            <button
              onClick={handleSaveDraft}
              disabled={saving}
              className="flex items-center gap-2 px-3 py-1.5 rounded-lg text-xs font-bold border border-white/10 text-white/50 hover:text-white hover:border-white/20 transition-all disabled:opacity-50"
            >
              Save Draft
            </button>
            <button
              onClick={handlePublish}
              disabled={saving}
              className="flex items-center gap-2 px-4 py-1.5 rounded-lg text-xs font-black bg-memzent-glow text-black hover:scale-105 transition-all disabled:opacity-50"
            >
              <Save size={12} />
              {saving ? "Publishing..." : "Publish"}
            </button>
          </div>
        </div>
      </div>

      <div className="max-w-4xl mx-auto px-6 py-8 space-y-6">
        <h1 className="text-2xl font-black tracking-tight">New Blog Post</h1>

        {error && (
          <div className="p-3 rounded-lg border border-red-500/20 bg-red-500/5 text-xs text-red-400">
            {error}
          </div>
        )}
        {saved && (
          <div className="p-3 rounded-lg border border-green-500/20 bg-green-500/5 text-xs text-green-400">
            Post saved successfully!
          </div>
        )}

        {/* Title */}
        <div className="space-y-2">
          <label className="text-[10px] font-black uppercase tracking-widest text-white/40">Title</label>
          <input
            type="text"
            value={title}
            onChange={(e) => handleTitleChange(e.target.value)}
            placeholder="How Memzent Reduced Our LLM Costs by 80%"
            className="w-full px-4 py-3 rounded-xl bg-white/[0.03] border border-white/10 text-sm text-white placeholder:text-white/20 focus:outline-none focus:border-memzent-glow/40"
          />
        </div>

        {/* Slug */}
        <div className="space-y-2">
          <label className="text-[10px] font-black uppercase tracking-widest text-white/40">Slug</label>
          <input
            type="text"
            value={slug}
            onChange={(e) => setSlug(e.target.value)}
            placeholder="how-memzent-reduced-llm-costs"
            className="w-full px-4 py-3 rounded-xl bg-white/[0.03] border border-white/10 text-sm font-mono text-white/70 placeholder:text-white/20 focus:outline-none focus:border-memzent-glow/40"
          />
        </div>

        {/* Description */}
        <div className="space-y-2">
          <label className="text-[10px] font-black uppercase tracking-widest text-white/40">Description</label>
          <textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="A short summary for the blog listing card..."
            rows={2}
            className="w-full px-4 py-3 rounded-xl bg-white/[0.03] border border-white/10 text-sm text-white placeholder:text-white/20 focus:outline-none focus:border-memzent-glow/40 resize-none"
          />
        </div>

        {/* Category + Tags */}
        <div className="grid grid-cols-2 gap-4">
          <div className="space-y-2">
            <label className="text-[10px] font-black uppercase tracking-widest text-white/40">Category</label>
            <select
              value={category}
              onChange={(e) => setCategory(e.target.value)}
              className="w-full px-4 py-3 rounded-xl bg-white/[0.03] border border-white/10 text-sm text-white focus:outline-none focus:border-memzent-glow/40"
            >
              {Object.entries(BLOG_CATEGORIES).map(([key, val]) => (
                <option key={key} value={key} className="bg-black">{val.label}</option>
              ))}
            </select>
          </div>
          <div className="space-y-2">
            <label className="text-[10px] font-black uppercase tracking-widest text-white/40">Tags (comma-separated)</label>
            <input
              type="text"
              value={tags}
              onChange={(e) => setTags(e.target.value)}
              placeholder="caching, performance, cost"
              className="w-full px-4 py-3 rounded-xl bg-white/[0.03] border border-white/10 text-sm text-white placeholder:text-white/20 focus:outline-none focus:border-memzent-glow/40"
            />
          </div>
        </div>

        {/* Author + Cover */}
        <div className="grid grid-cols-2 gap-4">
          <div className="space-y-2">
            <label className="text-[10px] font-black uppercase tracking-widest text-white/40">Author</label>
            <input
              type="text"
              value={author}
              onChange={(e) => setAuthor(e.target.value)}
              className="w-full px-4 py-3 rounded-xl bg-white/[0.03] border border-white/10 text-sm text-white focus:outline-none focus:border-memzent-glow/40"
            />
          </div>
          <div className="space-y-2">
            <label className="text-[10px] font-black uppercase tracking-widest text-white/40">Cover Image URL</label>
            <input
              type="text"
              value={coverImage}
              onChange={(e) => setCoverImage(e.target.value)}
              placeholder="https://..."
              className="w-full px-4 py-3 rounded-xl bg-white/[0.03] border border-white/10 text-sm text-white placeholder:text-white/20 focus:outline-none focus:border-memzent-glow/40"
            />
          </div>
        </div>

        {/* Content */}
        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <label className="text-[10px] font-black uppercase tracking-widest text-white/40">Content (Markdown)</label>
            <Link
              href={safeBlogHref(slug)}
              target="_blank"
              className="flex items-center gap-1 text-[10px] text-white/30 hover:text-memzent-glow transition-colors"
            >
              <Eye size={10} />
              Preview
            </Link>
          </div>
          <textarea
            value={content}
            onChange={(e) => setContent(e.target.value)}
            placeholder="Write your blog post in Markdown..."
            rows={20}
            className="w-full px-4 py-3 rounded-xl bg-white/[0.03] border border-white/10 text-sm font-mono text-white/70 placeholder:text-white/20 focus:outline-none focus:border-memzent-glow/40 resize-y leading-relaxed"
          />
        </div>
      </div>
    </div>
  );
}

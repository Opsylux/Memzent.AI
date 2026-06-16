"use client";

import { useState, useEffect, useCallback } from "react";
import { useSearchParams } from "next/navigation";
import { supabase } from "@/lib/supabase";
import { BLOG_CATEGORIES } from "@/lib/blog-types";
import { Save, Eye, ArrowLeft, Upload, Image, X, Loader2 } from "lucide-react";
import Link from "next/link";

function renderPreviewMarkdown(md: string): string {
  const codeBlocks: string[] = [];
  let html = md.replace(/```(\w*)\n([\s\S]*?)```/g, (_m, lang, code) => {
    const escaped = code.replace(/</g, "&lt;").replace(/>/g, "&gt;");
    const idx = codeBlocks.length;
    codeBlocks.push(
      `<div class="my-4 rounded-lg overflow-hidden border border-white/10 bg-black/60">` +
        (lang ? `<div class="px-3 py-1 border-b border-white/10 bg-white/[0.02]"><span class="text-[10px] font-bold uppercase tracking-widest text-white/20">${lang}</span></div>` : "") +
        `<pre class="p-3 overflow-x-auto text-[13px] font-mono leading-relaxed text-slate-300"><code>${escaped}</code></pre></div>`
    );
    return `%%CB_${idx}%%`;
  });

  html = html.replace(/!\[([^\]]*)\]\(([^)]+)\)/g,
    '<figure class="my-4"><img src="$2" alt="$1" class="rounded-lg border border-white/10 max-w-full" /><figcaption class="text-[10px] text-white/30 text-center mt-1">$1</figcaption></figure>');
  html = html.replace(/^### (.+)$/gm, '<h3 class="text-base font-bold mt-6 mb-2">$1</h3>');
  html = html.replace(/^## (.+)$/gm, '<h2 class="text-lg font-bold mt-8 mb-3">$1</h2>');
  html = html.replace(/^# (.+)$/gm, '<h1 class="text-xl font-bold mt-8 mb-3">$1</h1>');
  html = html.replace(/^> (.+)$/gm, '<blockquote class="border-l-2 border-memzent-glow/30 pl-4 my-3 text-white/40 italic text-sm">$1</blockquote>');
  html = html.replace(/\*\*(.+?)\*\*/g, '<strong class="text-white">$1</strong>');
  html = html.replace(/\*(.+?)\*/g, '<em>$1</em>');
  html = html.replace(/`([^`]+)`/g, '<code class="px-1.5 py-0.5 rounded bg-white/10 text-[12px] font-mono text-memzent-glow/80">$1</code>');
  html = html.replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2" class="text-memzent-glow underline">$1</a>');
  html = html.replace(/^- (.+)$/gm, '<li class="ml-4 list-disc text-sm text-white/60">$1</li>');
  html = html.replace(/^(\d+)\. (.+)$/gm, '<li class="ml-4 list-decimal text-sm text-white/60">$2</li>');
  html = html.replace(/^---$/gm, '<hr class="border-white/10 my-6" />');
  html = html.replace(/\n\n/g, '</p><p class="text-sm text-white/60 leading-relaxed my-2">');
  html = html.replace(/\n/g, '<br/>');
  codeBlocks.forEach((block, idx) => { html = html.replace(`%%CB_${idx}%%`, block); });
  return `<p class="text-sm text-white/60 leading-relaxed">${html}</p>`;
}

export default function BlogAdminPage() {
  const searchParams = useSearchParams();
  const editSlug = searchParams.get("slug");

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
  const [tab, setTab] = useState<"write" | "preview">("write");
  const [uploading, setUploading] = useState(false);
  const [isEdit, setIsEdit] = useState(false);

  // Load existing post when editing
  useEffect(() => {
    if (!editSlug) return;
    (async () => {
      const { data } = await supabase
        .from("blog_posts")
        .select("*")
        .eq("slug", editSlug)
        .single();
      if (data) {
        setTitle(data.title || "");
        setSlug(data.slug || "");
        setDescription(data.description || "");
        setContent(data.content || "");
        setCategory(data.category || "engineering");
        setTags((data.tags || []).join(", "));
        setAuthor(data.author || "Memzent Team");
        setCoverImage(data.cover_image || "");
        setIsEdit(true);
      }
    })();
  }, [editSlug]);

  const generateSlug = (text: string) => {
    return text
      .toLowerCase()
      .replace(/[^a-z0-9]+/g, "-")
      .replace(/^-|-$/g, "")
      .slice(0, 128);
  };

  const handleTitleChange = (val: string) => {
    setTitle(val);
    if (!isEdit && (!slug || slug === generateSlug(title))) {
      setSlug(generateSlug(val));
    }
  };

  const handleImageUpload = useCallback(async (file: File) => {
    if (!file.type.startsWith("image/")) {
      setError("Only image files are allowed.");
      return null;
    }
    if (file.size > 5 * 1024 * 1024) {
      setError("Image must be under 5MB.");
      return null;
    }

    setUploading(true);
    setError("");

    const ext = file.name.split(".").pop() || "png";
    const path = `blog/${Date.now()}-${Math.random().toString(36).slice(2, 8)}.${ext}`;

    const { error: uploadError } = await supabase.storage
      .from("blog-images")
      .upload(path, file, { cacheControl: "31536000", upsert: false });

    setUploading(false);

    if (uploadError) {
      setError(`Upload failed: ${uploadError.message}`);
      return null;
    }

    const { data: urlData } = supabase.storage.from("blog-images").getPublicUrl(path);
    return urlData.publicUrl;
  }, []);

  const handleCoverUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    const url = await handleImageUpload(file);
    if (url) setCoverImage(url);
  };

  const handleInlineImageUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    const url = await handleImageUpload(file);
    if (url) {
      const markdown = `\n![${file.name.replace(/\.[^.]+$/, "")}](${url})\n`;
      setContent((prev) => prev + markdown);
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
    if (dbError) { setError(dbError.message); }
    else { setSaved(true); setIsEdit(true); setTimeout(() => setSaved(false), 3000); }
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
    else { setSaved(true); setIsEdit(true); setTimeout(() => setSaved(false), 3000); }
  };

  return (
    <div className="min-h-screen bg-[#030507]">
      <div className="border-b border-white/5 bg-black/40">
        <div className="max-w-4xl mx-auto px-6 py-4 flex items-center justify-between">
          <Link href="/blog" className="flex items-center gap-2 text-xs text-white/40 hover:text-memzent-glow transition-colors font-bold">
            <ArrowLeft size={12} />
            All Posts
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
              {saving ? "Saving..." : isEdit ? "Update & Publish" : "Publish"}
            </button>
          </div>
        </div>
      </div>

      <div className="max-w-4xl mx-auto px-6 py-8 space-y-6">
        <h1 className="text-2xl font-black tracking-tight">
          {isEdit ? "Edit Post" : "New Blog Post"}
        </h1>

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

        {/* Author + Cover Image */}
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
            <label className="text-[10px] font-black uppercase tracking-widest text-white/40">Cover Image</label>
            <div className="flex gap-2">
              <input
                type="text"
                value={coverImage}
                onChange={(e) => setCoverImage(e.target.value)}
                placeholder="URL or upload →"
                className="flex-1 px-4 py-3 rounded-xl bg-white/[0.03] border border-white/10 text-sm text-white placeholder:text-white/20 focus:outline-none focus:border-memzent-glow/40"
              />
              <label className="flex items-center gap-1 px-3 py-2 rounded-xl bg-white/5 border border-white/10 text-[10px] font-bold text-white/50 hover:text-white hover:bg-white/10 cursor-pointer transition-all">
                {uploading ? <Loader2 size={12} className="animate-spin" /> : <Upload size={12} />}
                <input type="file" accept="image/*" onChange={handleCoverUpload} className="hidden" />
              </label>
            </div>
            {coverImage && (
              <div className="relative mt-2 rounded-lg overflow-hidden border border-white/10 h-32">
                <img src={coverImage} alt="Cover preview" className="w-full h-full object-cover" />
                <button
                  onClick={() => setCoverImage("")}
                  className="absolute top-2 right-2 p-1 rounded-full bg-black/60 text-white/60 hover:text-white transition-colors"
                >
                  <X size={12} />
                </button>
              </div>
            )}
          </div>
        </div>

        {/* Content with Tabs */}
        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <button
                onClick={() => setTab("write")}
                className={`text-[10px] font-black uppercase tracking-widest px-3 py-1 rounded-lg transition-all ${
                  tab === "write" ? "bg-white/10 text-white" : "text-white/40 hover:text-white/60"
                }`}
              >
                Write
              </button>
              <button
                onClick={() => setTab("preview")}
                className={`flex items-center gap-1 text-[10px] font-black uppercase tracking-widest px-3 py-1 rounded-lg transition-all ${
                  tab === "preview" ? "bg-white/10 text-white" : "text-white/40 hover:text-white/60"
                }`}
              >
                <Eye size={10} />
                Preview
              </button>
            </div>
            <label className="flex items-center gap-1 text-[10px] text-white/30 hover:text-memzent-glow cursor-pointer transition-colors">
              <Image size={10} />
              Insert Image
              <input type="file" accept="image/*" onChange={handleInlineImageUpload} className="hidden" />
            </label>
          </div>

          {tab === "write" ? (
            <textarea
              value={content}
              onChange={(e) => setContent(e.target.value)}
              placeholder="Write your blog post in Markdown..."
              rows={24}
              className="w-full px-4 py-3 rounded-xl bg-white/[0.03] border border-white/10 text-sm font-mono text-white/70 placeholder:text-white/20 focus:outline-none focus:border-memzent-glow/40 resize-y leading-relaxed"
            />
          ) : (
            <div className="min-h-[500px] px-6 py-4 rounded-xl bg-white/[0.02] border border-white/10 overflow-y-auto prose-invert max-w-none">
              {content ? (
                <div dangerouslySetInnerHTML={{ __html: renderPreviewMarkdown(content) }} />
              ) : (
                <p className="text-white/20 text-sm italic">Nothing to preview yet...</p>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

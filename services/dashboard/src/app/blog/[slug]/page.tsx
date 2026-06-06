import { getPostBySlug, getAllPosts } from "@/lib/blog";
import { BLOG_CATEGORIES } from "@/lib/blog-types";
import { notFound } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, Calendar, Clock, User } from "lucide-react";

export async function generateStaticParams() {
  const posts = await getAllPosts();
  return posts.map((post) => ({ slug: post.slug }));
}

export default async function BlogPostPage({ params }: { params: Promise<{ slug: string }> }) {
  const { slug } = await params;
  const post = await getPostBySlug(slug);

  if (!post) notFound();

  const cat = BLOG_CATEGORIES[post.category];

  return (
    <div className="min-h-screen bg-[#030507]">
      {/* Navigation */}
      <div className="border-b border-white/5 bg-black/40">
        <div className="max-w-4xl mx-auto px-6 py-4">
          <Link href="/blog" className="flex items-center gap-2 text-xs text-white/40 hover:text-memzent-glow transition-colors font-bold">
            <ArrowLeft size={12} />
            Back to Blog
          </Link>
        </div>
      </div>

      {/* Article */}
      <article className="max-w-4xl mx-auto px-6 py-12">
        {/* Header */}
        <header className="mb-10 space-y-4">
          <div className="flex items-center gap-3">
            <span className={`text-[10px] font-black uppercase tracking-widest px-2 py-0.5 rounded ${cat.bg} ${cat.color}`}>
              {cat.label}
            </span>
            {post.tags.map((tag) => (
              <span key={tag} className="text-[10px] font-bold text-white/20 bg-white/5 px-2 py-0.5 rounded">
                {tag}
              </span>
            ))}
          </div>

          <h1 className="text-3xl sm:text-4xl font-black tracking-tight">
            {post.title}
          </h1>

          {post.description && (
            <p className="text-lg text-white/50 leading-relaxed">
              {post.description}
            </p>
          )}

          <div className="flex items-center gap-4 pt-4 border-t border-white/5">
            <div className="flex items-center gap-2">
              {post.author_avatar ? (
                <img src={post.author_avatar} alt={post.author} className="w-6 h-6 rounded-full" />
              ) : (
                <div className="w-6 h-6 rounded-full bg-memzent-glow/10 flex items-center justify-center">
                  <User size={12} className="text-memzent-glow" />
                </div>
              )}
              <span className="text-xs font-bold text-white/60">{post.author}</span>
            </div>
            <span className="flex items-center gap-1 text-[10px] text-white/30">
              <Calendar size={10} />
              {new Date(post.published_at).toLocaleDateString("en-US", { month: "long", day: "numeric", year: "numeric" })}
            </span>
            <span className="flex items-center gap-1 text-[10px] text-white/30">
              <Clock size={10} />
              {post.reading_time} min read
            </span>
          </div>
        </header>

        {/* Cover image */}
        {post.cover_image && (
          <div className="aspect-video rounded-2xl overflow-hidden mb-10 border border-white/5">
            <img src={post.cover_image} alt={post.title} className="w-full h-full object-cover" />
          </div>
        )}

        {/* Content */}
        <div className="prose prose-invert prose-sm max-w-none
          prose-headings:font-black prose-headings:tracking-tight prose-headings:uppercase
          prose-h2:text-xl prose-h2:mt-10 prose-h2:mb-4
          prose-h3:text-base prose-h3:mt-8 prose-h3:mb-3
          prose-p:text-white/60 prose-p:leading-relaxed prose-p:text-sm
          prose-a:text-memzent-glow prose-a:no-underline hover:prose-a:underline
          prose-code:text-memzent-glow/80 prose-code:bg-memzent-glow/5 prose-code:px-1 prose-code:rounded prose-code:font-mono prose-code:text-xs
          prose-pre:bg-black/60 prose-pre:border prose-pre:border-white/5 prose-pre:rounded-xl
          prose-strong:text-white/80
          prose-li:text-white/50 prose-li:text-sm
          prose-blockquote:border-memzent-glow/30 prose-blockquote:text-white/40
        ">
          <div dangerouslySetInnerHTML={{ __html: renderMarkdown(post.content) }} />
        </div>
      </article>
    </div>
  );
}

/**
 * Simple markdown to HTML renderer (no external deps).
 * For production, replace with remark/rehype or MDX compilation.
 */
function renderMarkdown(md: string): string {
  let html = md
    // Code blocks
    .replace(/```(\w*)\n([\s\S]*?)```/g, '<pre><code class="language-$1">$2</code></pre>')
    // Inline code
    .replace(/`([^`]+)`/g, '<code>$1</code>')
    // Headers
    .replace(/^### (.+)$/gm, '<h3>$1</h3>')
    .replace(/^## (.+)$/gm, '<h2>$1</h2>')
    .replace(/^# (.+)$/gm, '<h1>$1</h1>')
    // Bold and italic
    .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
    .replace(/\*(.+?)\*/g, '<em>$1</em>')
    // Links
    .replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2">$1</a>')
    // Unordered lists
    .replace(/^- (.+)$/gm, '<li>$1</li>')
    // Paragraphs (simple: double newlines)
    .replace(/\n\n/g, '</p><p>')
    // Line breaks
    .replace(/\n/g, '<br/>');

  // Wrap list items
  html = html.replace(/(<li>[\s\S]*?<\/li>)/g, '<ul>$1</ul>');
  // Remove duplicate ul wrappers
  html = html.replace(/<\/ul>\s*<ul>/g, '');

  return `<p>${html}</p>`;
}

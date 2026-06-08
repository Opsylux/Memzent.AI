import { getPostBySlug, getAllPosts } from "@/lib/blog";
import { BLOG_CATEGORIES } from "@/lib/blog-types";
import { notFound } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, Calendar, Clock, User } from "lucide-react";

export const dynamic = "force-dynamic";

export default async function BlogPostPage({ params }: { params: Promise<{ slug: string }> }) {
  const { slug } = await params;
  const post = await getPostBySlug(slug);

  if (!post) notFound();

  // Find 2 related posts by shared tags or same category
  const allPosts = await getAllPosts();
  const related = allPosts
    .filter((p) => p.slug !== post.slug)
    .map((p) => {
      const sharedTags = p.tags.filter((t) => post.tags.includes(t)).length;
      const sameCategory = p.category === post.category ? 1 : 0;
      return { post: p, score: sharedTags * 2 + sameCategory };
    })
    .sort((a, b) => b.score - a.score)
    .slice(0, 2)
    .map((r) => r.post);

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

      {/* Suggested Posts */}
      {related.length > 0 && (
        <section className="max-w-4xl mx-auto px-6 pb-16">
          <div className="border-t border-white/5 pt-10">
            <h2 className="text-xs font-black uppercase tracking-widest text-white/30 mb-6">
              Continue Reading
            </h2>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              {related.map((r) => {
                const rCat = BLOG_CATEGORIES[r.category];
                return (
                  <Link
                    key={r.slug}
                    href={`/blog/${r.slug}`}
                    className="group block p-5 rounded-xl border border-white/5 bg-white/[0.02] hover:bg-white/[0.04] hover:border-white/10 transition-all"
                  >
                    <span className={`text-[9px] font-black uppercase tracking-widest ${rCat.color}`}>
                      {rCat.label}
                    </span>
                    <h3 className="text-sm font-bold text-white/80 mt-2 group-hover:text-memzent-glow transition-colors line-clamp-2">
                      {r.title}
                    </h3>
                    <p className="text-[11px] text-white/30 mt-2 line-clamp-2">
                      {r.description}
                    </p>
                    <div className="flex items-center gap-3 mt-3">
                      <span className="text-[10px] text-white/20">{r.reading_time} min read</span>
                      <span className="text-[10px] text-white/20">
                        {new Date(r.published_at).toLocaleDateString("en-US", { month: "short", day: "numeric" })}
                      </span>
                    </div>
                  </Link>
                );
              })}
            </div>
          </div>
        </section>
      )}
    </div>
  );
}

/**
 * Markdown to HTML renderer supporting:
 * - Code blocks (fenced with ```) with language class
 * - Inline code
 * - Images (![alt](src))
 * - Headers (h1-h3)
 * - Bold, italic
 * - Links
 * - Unordered and ordered lists
 * - Blockquotes
 * - Tables (GFM-style)
 * - Horizontal rules
 */
function renderMarkdown(md: string): string {
  // Protect code blocks from other transforms
  const codeBlocks: string[] = [];
  let html = md.replace(/```(\w*)\n([\s\S]*?)```/g, (_match, lang, code) => {
    const escaped = code.replace(/</g, '&lt;').replace(/>/g, '&gt;');
    const idx = codeBlocks.length;
    codeBlocks.push(
      `<div class="relative group my-6 rounded-xl overflow-hidden border border-white/5 bg-black/60">` +
      (lang ? `<div class="px-4 py-1.5 border-b border-white/5 bg-white/[0.02]"><span class="text-[10px] font-black uppercase tracking-widest text-white/20">${lang}</span></div>` : '') +
      `<pre class="p-4 overflow-x-auto text-[13px] font-mono leading-relaxed text-slate-300"><code class="language-${lang}">${escaped}</code></pre></div>`
    );
    return `%%CODEBLOCK_${idx}%%`;
  });

  // Images
  html = html.replace(/!\[([^\]]*)\]\(([^)]+)\)/g,
    '<figure class="my-6"><img src="$2" alt="$1" class="rounded-xl border border-white/5 w-full" /><figcaption class="text-[10px] text-white/30 text-center mt-2">$1</figcaption></figure>'
  );

  // Tables (GFM)
  html = html.replace(/^(\|.+\|)\n(\|[-| :]+\|)\n((?:\|.+\|\n?)+)/gm, (_match, header, _sep, body) => {
    const headers = header.split('|').filter((c: string) => c.trim()).map((c: string) =>
      `<th class="text-left px-4 py-2 font-black text-white/60 text-xs">${c.trim()}</th>`
    ).join('');
    const rows = body.trim().split('\n').map((row: string) => {
      const cells = row.split('|').filter((c: string) => c.trim()).map((c: string) =>
        `<td class="px-4 py-2 text-xs text-white/40">${c.trim()}</td>`
      ).join('');
      return `<tr class="border-b border-white/5">${cells}</tr>`;
    }).join('');
    return `<div class="overflow-x-auto my-6"><table class="w-full border border-white/5 rounded-lg overflow-hidden"><thead><tr class="bg-white/[0.03] border-b border-white/5">${headers}</tr></thead><tbody>${rows}</tbody></table></div>`;
  });

  // Blockquotes
  html = html.replace(/^> (.+)$/gm, '<blockquote class="border-l-2 border-memzent-glow/30 pl-4 my-4 text-white/40 italic text-sm">$1</blockquote>');

  // Horizontal rules
  html = html.replace(/^---$/gm, '<hr class="border-white/5 my-8" />');

  // Headers
  html = html.replace(/^### (.+)$/gm, '<h3>$1</h3>');
  html = html.replace(/^## (.+)$/gm, '<h2>$1</h2>');
  html = html.replace(/^# (.+)$/gm, '<h1>$1</h1>');

  // Inline code
  html = html.replace(/`([^`]+)`/g, '<code>$1</code>');

  // Bold and italic
  html = html.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>');
  html = html.replace(/\*(.+?)\*/g, '<em>$1</em>');

  // Links
  html = html.replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2">$1</a>');

  // Ordered lists
  html = html.replace(/^(\d+)\. (.+)$/gm, '<li class="list-decimal">$2</li>');

  // Unordered lists
  html = html.replace(/^- (.+)$/gm, '<li>$1</li>');

  // Wrap consecutive list items
  html = html.replace(/((?:<li[^>]*>.*?<\/li>\s*)+)/g, '<ul class="space-y-1 my-4 pl-5">$1</ul>');

  // Paragraphs (double newlines)
  html = html.replace(/\n\n/g, '</p><p>');

  // Line breaks within paragraphs
  html = html.replace(/\n/g, '<br/>');

  // Restore code blocks
  codeBlocks.forEach((block, idx) => {
    html = html.replace(`%%CODEBLOCK_${idx}%%`, block);
  });

  return `<p>${html}</p>`;
}


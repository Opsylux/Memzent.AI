import { useParams, Link } from "react-router-dom";
import { ArrowLeft, Calendar, Clock, User } from "lucide-react";
import { motion } from "framer-motion";
import { Helmet } from "react-helmet-async";
import { getPostBySlug, getAllPosts, renderMarkdown } from "../lib/blog";
import { BLOG_CATEGORIES } from "../lib/blog-types";

export default function BlogPostPage() {
  const { slug } = useParams<{ slug: string }>();
  const post = slug ? getPostBySlug(slug) : null;

  if (!post) {
    return (
      <div className="min-h-screen bg-memzent-dark pt-24 flex items-center justify-center">
        <div className="text-center space-y-4">
          <h1 className="text-2xl font-black">Post Not Found</h1>
          <p className="text-white/40 text-sm">The post you're looking for doesn't exist.</p>
          <Link to="/blog" className="inline-flex items-center gap-2 text-memzent-glow text-sm font-bold hover:underline">
            <ArrowLeft size={14} /> Back to Blog
          </Link>
        </div>
      </div>
    );
  }

  const cat = BLOG_CATEGORIES[post.category] || BLOG_CATEGORIES.engineering;

  // Find related posts
  const allPosts = getAllPosts();
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

  return (
    <div className="min-h-screen bg-memzent-dark pt-24">
      <Helmet>
        <title>{post.title} | Memzent Blog</title>
        <meta name="description" content={post.description || `${post.title} — technical deep-dive from the Memzent engineering team.`} />
        <meta property="og:title" content={post.title} />
        <meta property="og:description" content={post.description} />
        <meta property="og:type" content="article" />
        <meta property="og:url" content={`https://memzent.ai/blog/${post.slug}`} />
        {post.cover_image && <meta property="og:image" content={post.cover_image} />}
        <meta name="twitter:card" content="summary_large_image" />
        <meta name="twitter:title" content={post.title} />
        <meta name="twitter:description" content={post.description} />
        <meta property="article:published_time" content={post.published_at} />
        <meta property="article:author" content={post.author} />
        <meta name="keywords" content={post.tags.join(", ")} />
        <link rel="canonical" href={`https://memzent.ai/blog/${post.slug}`} />
      </Helmet>

      {/* Navigation */}
      <div className="border-b border-white/5 bg-black/40">
        <div className="max-w-4xl mx-auto px-6 py-4">
          <Link
            to="/blog"
            className="flex items-center gap-2 text-xs text-white/40 hover:text-memzent-glow transition-colors font-bold"
          >
            <ArrowLeft size={12} />
            Back to Blog
          </Link>
        </div>
      </div>

      {/* Article */}
      <motion.article
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5 }}
        className="max-w-4xl mx-auto px-6 py-12"
      >
        {/* Header */}
        <header className="mb-10 space-y-4">
          <div className="flex items-center gap-3 flex-wrap">
            <span className={`text-[10px] font-black uppercase tracking-widest px-2 py-0.5 rounded ${cat.bg} ${cat.color}`}>
              {cat.label}
            </span>
            {post.tags.map((tag) => (
              <span key={tag} className="text-[10px] font-bold text-white/20 bg-white/5 px-2 py-0.5 rounded">
                {tag}
              </span>
            ))}
          </div>

          <h1 className="text-3xl sm:text-4xl font-black tracking-tight">{post.title}</h1>

          {post.description && (
            <p className="text-lg text-white/50 leading-relaxed">{post.description}</p>
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
              {new Date(post.published_at).toLocaleDateString("en-US", {
                month: "long",
                day: "numeric",
                year: "numeric",
              })}
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
        <div
          className="prose prose-invert prose-sm max-w-none
            prose-headings:font-black prose-headings:tracking-tight prose-headings:uppercase
            prose-h2:text-xl prose-h2:mt-10 prose-h2:mb-4
            prose-h3:text-base prose-h3:mt-8 prose-h3:mb-3
            prose-p:text-white/60 prose-p:leading-relaxed prose-p:text-sm
            prose-a:text-memzent-glow prose-a:no-underline hover:prose-a:underline
            prose-code:text-memzent-glow/80 prose-code:bg-memzent-glow/5 prose-code:px-1 prose-code:rounded prose-code:font-mono prose-code:text-xs
            prose-pre:bg-black/60 prose-pre:border prose-pre:border-white/5 prose-pre:rounded-xl
            prose-strong:text-white/80
            prose-li:text-white/50 prose-li:text-sm
            prose-blockquote:border-memzent-glow/30 prose-blockquote:text-white/40"
          dangerouslySetInnerHTML={{ __html: renderMarkdown(post.content) }}
        />
      </motion.article>

      {/* Suggested Posts */}
      {related.length > 0 && (
        <section className="max-w-4xl mx-auto px-6 pb-16">
          <div className="border-t border-white/5 pt-10">
            <h2 className="text-xs font-black uppercase tracking-widest text-white/30 mb-6">
              Continue Reading
            </h2>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              {related.map((r) => {
                const rCat = BLOG_CATEGORIES[r.category] || BLOG_CATEGORIES.engineering;
                return (
                  <Link
                    key={r.slug}
                    to={`/blog/${r.slug}`}
                    className="group block p-5 rounded-xl border border-white/5 bg-white/[0.02] hover:bg-white/[0.04] hover:border-white/10 transition-all"
                  >
                    <span className={`text-[9px] font-black uppercase tracking-widest ${rCat.color}`}>
                      {rCat.label}
                    </span>
                    <h3 className="text-sm font-bold text-white/80 mt-2 group-hover:text-memzent-glow transition-colors line-clamp-2">
                      {r.title}
                    </h3>
                    <p className="text-[11px] text-white/30 mt-2 line-clamp-2">{r.description}</p>
                    <div className="flex items-center gap-3 mt-3">
                      <span className="text-[10px] text-white/20">{r.reading_time} min read</span>
                      <span className="text-[10px] text-white/20">
                        {new Date(r.published_at).toLocaleDateString("en-US", {
                          month: "short",
                          day: "numeric",
                        })}
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

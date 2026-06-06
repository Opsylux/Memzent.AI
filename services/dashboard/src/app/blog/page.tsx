import { getAllPosts } from "@/lib/blog";
import { BLOG_CATEGORIES } from "@/lib/blog-types";
import Link from "next/link";
import { Calendar, Clock, ArrowRight, BookOpen } from "lucide-react";

export default async function BlogPage() {
  const posts = await getAllPosts();

  return (
    <div className="min-h-screen bg-[#030507]">
      {/* Header */}
      <div className="border-b border-white/5 bg-black/40">
        <div className="max-w-6xl mx-auto px-6 py-16">
          <div className="flex items-center gap-2 px-3 py-1 rounded-full bg-memzent-glow/5 border border-memzent-glow/20 w-fit mb-4">
            <BookOpen size={12} className="text-memzent-glow" />
            <span className="text-[10px] font-black text-memzent-glow uppercase tracking-tighter">Blog</span>
          </div>
          <h1 className="text-4xl sm:text-5xl font-black tracking-tighter uppercase mb-4">
            Memzent Blog
          </h1>
          <p className="text-lg text-white/50 max-w-2xl">
            Engineering deep-dives, use cases, tutorials, and product updates from the Memzent team.
          </p>
        </div>
      </div>

      {/* Posts Grid */}
      <div className="max-w-6xl mx-auto px-6 py-12">
        {posts.length === 0 ? (
          <div className="text-center py-20">
            <p className="text-white/30 text-sm">No posts yet. Check back soon!</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {posts.map((post) => {
              const cat = BLOG_CATEGORIES[post.category];
              return (
                <Link
                  key={post.slug}
                  href={`/blog/${post.slug}`}
                  className="group p-6 rounded-2xl bg-white/[0.02] border border-white/5 hover:border-white/10 hover:bg-white/[0.04] transition-all space-y-4"
                >
                  {post.cover_image && (
                    <div className="aspect-video rounded-xl overflow-hidden bg-white/5">
                      <img
                        src={post.cover_image}
                        alt={post.title}
                        className="w-full h-full object-cover opacity-80 group-hover:opacity-100 transition-opacity"
                      />
                    </div>
                  )}

                  <div className="flex items-center gap-2">
                    <span className={`text-[10px] font-black uppercase tracking-widest px-2 py-0.5 rounded ${cat.bg} ${cat.color}`}>
                      {cat.label}
                    </span>
                  </div>

                  <h2 className="text-sm font-black uppercase tracking-tight group-hover:text-memzent-glow transition-colors line-clamp-2">
                    {post.title}
                  </h2>

                  <p className="text-xs text-white/40 leading-relaxed line-clamp-3">
                    {post.description}
                  </p>

                  <div className="flex items-center justify-between pt-2 border-t border-white/5">
                    <div className="flex items-center gap-3 text-[10px] text-white/30">
                      <span className="flex items-center gap-1">
                        <Calendar size={10} />
                        {new Date(post.published_at).toLocaleDateString("en-US", { month: "short", day: "numeric", year: "numeric" })}
                      </span>
                      <span className="flex items-center gap-1">
                        <Clock size={10} />
                        {post.reading_time} min read
                      </span>
                    </div>
                    <ArrowRight size={12} className="text-white/20 group-hover:text-memzent-glow transition-colors" />
                  </div>
                </Link>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}

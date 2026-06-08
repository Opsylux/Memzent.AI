import type { BlogPost } from "./blog-types";

function parseFrontmatter(raw: string): { meta: Record<string, string>; content: string } {
  const meta: Record<string, string> = {};
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

function estimateReadingTime(content: string): number {
  const words = content.split(/\s+/).length;
  return Math.max(1, Math.ceil(words / 200));
}

// Load all markdown files at build time via Vite's import.meta.glob
const mdModules = import.meta.glob("../content/blog/*.md", {
  query: "?raw",
  import: "default",
  eager: true,
}) as Record<string, string>;

function loadPosts(): BlogPost[] {
  const posts: BlogPost[] = [];

  for (const [filepath, raw] of Object.entries(mdModules)) {
    const filename = filepath.split("/").pop() || "";
    const slug = filename.replace(/\.md$/, "");
    const { meta, content } = parseFrontmatter(raw);

    if (!meta.title || !meta.published_at) continue;

    posts.push({
      slug,
      title: meta.title,
      description: meta.description || "",
      content,
      author: meta.author || "Memzent Team",
      author_avatar: meta.author_avatar,
      cover_image: meta.cover_image,
      category: (meta.category as BlogPost["category"]) || "engineering",
      tags: meta.tags ? meta.tags.split(",").map((t) => t.trim()) : [],
      published_at: meta.published_at,
      updated_at: meta.updated_at,
      reading_time: estimateReadingTime(content),
    });
  }

  posts.sort((a, b) => new Date(b.published_at).getTime() - new Date(a.published_at).getTime());
  return posts;
}

// Cached posts (loaded once at module init since they're build-time static)
const allPosts = loadPosts();

export function getAllPosts(): BlogPost[] {
  return allPosts;
}

export function getPostBySlug(slug: string): BlogPost | null {
  return allPosts.find((p) => p.slug === slug) || null;
}

/**
 * Markdown to HTML renderer
 */
export function renderMarkdown(md: string): string {
  const codeBlocks: string[] = [];
  let html = md.replace(/```(\w*)\n([\s\S]*?)```/g, (_match, lang, code) => {
    const escaped = code.replace(/</g, "&lt;").replace(/>/g, "&gt;");
    const idx = codeBlocks.length;
    codeBlocks.push(
      `<div class="relative group my-6 rounded-xl overflow-hidden border border-white/5 bg-black/60">` +
        (lang
          ? `<div class="px-4 py-1.5 border-b border-white/5 bg-white/[0.02]"><span class="text-[10px] font-black uppercase tracking-widest text-white/20">${lang}</span></div>`
          : "") +
        `<pre class="p-4 overflow-x-auto text-[13px] font-mono leading-relaxed text-slate-300"><code class="language-${lang}">${escaped}</code></pre></div>`
    );
    return `%%CODEBLOCK_${idx}%%`;
  });

  // Images
  html = html.replace(
    /!\[([^\]]*)\]\(([^)]+)\)/g,
    '<figure class="my-6"><img src="$2" alt="$1" class="rounded-xl border border-white/5 w-full" /><figcaption class="text-[10px] text-white/30 text-center mt-2">$1</figcaption></figure>'
  );

  // Tables (GFM)
  html = html.replace(
    /^(\|.+\|)\n(\|[-| :]+\|)\n((?:\|.+\|\n?)+)/gm,
    (_match: string, header: string, _sep: string, body: string) => {
      const headers = header
        .split("|")
        .filter((c: string) => c.trim())
        .map((c: string) => `<th class="text-left px-4 py-2 font-black text-white/60 text-xs">${c.trim()}</th>`)
        .join("");
      const rows = body
        .trim()
        .split("\n")
        .map((row: string) => {
          const cells = row
            .split("|")
            .filter((c: string) => c.trim())
            .map((c: string) => `<td class="px-4 py-2 text-xs text-white/40">${c.trim()}</td>`)
            .join("");
          return `<tr class="border-b border-white/5">${cells}</tr>`;
        })
        .join("");
      return `<div class="overflow-x-auto my-6"><table class="w-full border border-white/5 rounded-lg overflow-hidden"><thead><tr class="bg-white/[0.03] border-b border-white/5">${headers}</tr></thead><tbody>${rows}</tbody></table></div>`;
    }
  );

  // Blockquotes
  html = html.replace(
    /^> (.+)$/gm,
    '<blockquote class="border-l-2 border-memzent-glow/30 pl-4 my-4 text-white/40 italic text-sm">$1</blockquote>'
  );

  // Horizontal rules
  html = html.replace(/^---$/gm, '<hr class="border-white/5 my-8" />');

  // Headers
  html = html.replace(/^### (.+)$/gm, "<h3>$1</h3>");
  html = html.replace(/^## (.+)$/gm, "<h2>$1</h2>");
  html = html.replace(/^# (.+)$/gm, "<h1>$1</h1>");

  // Inline code
  html = html.replace(/`([^`]+)`/g, "<code>$1</code>");

  // Bold and italic
  html = html.replace(/\*\*(.+?)\*\*/g, "<strong>$1</strong>");
  html = html.replace(/\*(.+?)\*/g, "<em>$1</em>");

  // Links
  html = html.replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2">$1</a>');

  // Ordered lists
  html = html.replace(/^(\d+)\. (.+)$/gm, '<li class="list-decimal">$2</li>');

  // Unordered lists
  html = html.replace(/^- (.+)$/gm, "<li>$1</li>");

  // Wrap consecutive list items
  html = html.replace(/((?:<li[^>]*>.*?<\/li>\s*)+)/g, '<ul class="space-y-1 my-4 pl-5">$1</ul>');

  // Paragraphs
  html = html.replace(/\n\n/g, "</p><p>");
  html = html.replace(/\n/g, "<br/>");

  // Restore code blocks
  codeBlocks.forEach((block, idx) => {
    html = html.replace(`%%CODEBLOCK_${idx}%%`, block);
  });

  return `<p>${html}</p>`;
}

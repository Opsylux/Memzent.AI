import { redirect } from "next/navigation";

// Individual blog posts are viewed on the public website
// Edit links route to /blog/admin?slug=...
export default async function BlogPostPage({ params }: { params: Promise<{ slug: string }> }) {
  const { slug } = await params;
  redirect(`/blog/admin?slug=${encodeURIComponent(slug)}`);
}


import { redirect } from "next/navigation";

// Blog has moved to the public website (memzent.ai/blog)
export default async function BlogPostPage({ params }: { params: Promise<{ slug: string }> }) {
  const { slug } = await params;
  redirect(`https://memzent.ai/blog/${slug}`);
}


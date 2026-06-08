import { redirect } from "next/navigation";

// Blog has moved to the public website (memzent.ai/blog)
export default function BlogPage() {
  redirect("https://memzent.ai/blog");
}

import { Sidebar } from "@/components/sidebar";
import { MemzentTopNav } from "@/components/memzent-top-nav";
import { NeuralAssistant } from "@/components/neural-assistant";
import { getCurrentOrg, type OrgContext } from "@/lib/user-context";
import { redirect } from "next/navigation";

export const dynamic = "force-dynamic";

export default async function AppLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const org = await getCurrentOrg();

  // Middleware handles the primary auth redirect; this is a safety fallback
  if (!org) {
    redirect("/login");
  }

  return (
    <div className="flex min-h-screen w-full bg-memzent-dark relative">
      {/* Sidebar - responsive (hidden on mobile, visible on lg+) */}
      <Sidebar orgName={org.orgName} tier={org.tier} initials={org.initials} role={org.role} />

      {/* Main Content Area */}
      <div className="flex-1 flex flex-col min-h-screen relative overflow-hidden w-full">
        {/* Top Navigation */}
        <MemzentTopNav
          orgName={org.orgName}
          email={org.email}
          initials={org.initials}
          tier={org.tier}
          role={org.role}
        />

        {/* Page Content — responsive padding */}
        <main className="flex-1 overflow-auto px-4 py-6 sm:px-6 lg:p-8 mt-2">
          <div className="max-w-7xl mx-auto">
            {children}
          </div>
        </main>

        <NeuralAssistant orgId={org.orgId} />

        {/* Background flair — hidden on mobile for performance */}
        <div className="hidden sm:block fixed top-0 right-0 w-[500px] h-[500px] bg-memzent-glow/5 blur-[120px] -z-10 rounded-full" />
        <div className="hidden sm:block fixed bottom-0 right-0 w-[300px] h-[300px] bg-memzent-purple/5 blur-[100px] -z-10 rounded-full" />
      </div>
    </div>
  );
}

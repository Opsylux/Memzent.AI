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

  // We redirect to /login if no session exists to avoid 404s
  if (!org) {
    redirect("/login");
  }

  return (
    <div className="flex min-h-screen w-full bg-memzent-dark relative">
      {/* Sidebar - Fixed width */}
      <Sidebar orgName={org.orgName} tier={org.tier} initials={org.initials} role={org.role} />

      {/* Main Content Area */}
      <div className="flex-1 flex flex-col min-h-screen relative overflow-hidden">
        {/* Top Navigation - Sits within the main content area, offset by Sidebar */}
        <MemzentTopNav
          orgName={org.orgName}
          email={org.email}
          initials={org.initials}
          tier={org.tier}
          role={org.role}
        />

        {/* Page Content - Adjusted margin to fit perfectly with sticky header flow */}
        <main className="flex-1 overflow-auto p-8 mt-2">
          <div className="max-w-7xl mx-auto">
            {children}
          </div>
        </main>

        <NeuralAssistant orgId={org.orgId} />

        {/* Optional: Static background flair */}
        <div className="fixed top-0 right-0 w-[500px] h-[500px] bg-memzent-glow/5 blur-[120px] -z-10 rounded-full" />
        <div className="fixed bottom-0 right-0 w-[300px] h-[300px] bg-memzent-purple/5 blur-[100px] -z-10 rounded-full" />
      </div>
    </div>
  );
}

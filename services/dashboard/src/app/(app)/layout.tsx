import { Sidebar } from "@/components/sidebar";
import { AuraTopNav } from "@/components/aura-top-nav";

export default function AppLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="flex min-h-screen w-full bg-aura-dark relative">
      {/* Sidebar - Fixed width */}
      <Sidebar />

      {/* Main Content Area */}
      <div className="flex-1 flex flex-col min-h-screen relative overflow-hidden">
        {/* Top Navigation - Sits within the main content area, offset by Sidebar */}
        <AuraTopNav />
        
        {/* Page Content - Padded to avoid overlapping with TopNav */}
        <main className="flex-1 overflow-auto p-8 mt-40">
          <div className="max-w-7xl mx-auto">
            {children}
          </div>
        </main>
        
        {/* Optional: Static background flair */}
        <div className="fixed top-0 right-0 w-[500px] h-[500px] bg-aura-glow/5 blur-[120px] -z-10 rounded-full" />
        <div className="fixed bottom-0 right-0 w-[300px] h-[300px] bg-aura-purple/5 blur-[100px] -z-10 rounded-full" />
      </div>
    </div>
  );
}

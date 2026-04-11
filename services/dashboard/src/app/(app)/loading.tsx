import { Skeleton } from "@/components/ui/skeleton";

export default function Loading() {
  return (
    <div className="space-y-12 pb-20 animate-pulse">
      {/* Workspace Header Skeleton */}
      <div className="flex items-center gap-4 mb-4">
        <div className="w-2 h-8 rounded-full bg-white/10" />
        <div className="space-y-2">
          <div className="h-8 w-64 bg-white/5 rounded-lg" />
          <div className="h-3 w-32 bg-white/5 rounded-md" />
        </div>
      </div>

      {/* KPI Section Skeleton */}
      <section className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {[1, 2, 3, 4].map((i) => (
          <div key={i} className="h-32 rounded-3xl bg-white/[0.02] border border-white/5 p-6 space-y-4">
            <div className="flex justify-between items-start">
              <div className="w-10 h-10 rounded-xl bg-white/5" />
              <div className="w-12 h-4 bg-white/5 rounded" />
            </div>
            <div className="space-y-2">
              <div className="h-6 w-24 bg-white/5 rounded" />
              <div className="h-3 w-32 bg-white/5 rounded" />
            </div>
          </div>
        ))}
      </section>

      {/* Main Intelligence Grid Skeleton */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        <section className="lg:col-span-2 space-y-8">
          <div className="flex items-center justify-between">
            <div className="space-y-2">
              <div className="h-8 w-48 bg-white/5 rounded-lg" />
              <div className="h-3 w-64 bg-white/5 rounded-md" />
            </div>
          </div>
          <div className="h-[400px] rounded-3xl bg-white/[0.02] border border-white/5 shadow-2xl" />
          
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="h-[300px] rounded-3xl bg-white/[0.02] border border-white/5" />
            <div className="h-[300px] rounded-3xl bg-white/[0.02] border border-white/5" />
          </div>
        </section>

        <section className="space-y-8">
            <div className="h-96 rounded-3xl bg-white/[0.02] border border-white/5 p-8 space-y-6">
                <div className="h-4 w-32 bg-white/10 rounded" />
                <div className="space-y-4">
                    {[1,2,3,4,5].map(i => (
                        <div key={i} className="flex items-center justify-between">
                            <div className="flex items-center gap-3">
                                <div className="w-2 h-2 rounded-full bg-white/10" />
                                <div className="space-y-1">
                                    <div className="h-3 w-24 bg-white/5 rounded" />
                                    <div className="h-2 w-16 bg-white/5 rounded" />
                                </div>
                            </div>
                            <div className="h-4 w-12 bg-white/5 rounded" />
                        </div>
                    ))}
                </div>
            </div>
        </section>
      </div>
    </div>
  );
}

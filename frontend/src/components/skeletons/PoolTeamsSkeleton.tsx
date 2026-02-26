import { Skeleton } from '../ui/Skeleton';

export function PoolTeamsSkeleton() {
  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between mb-4">
        <Skeleton className="h-6 w-40" />
        <Skeleton className="h-5 w-24" />
      </div>
      {Array.from({ length: 10 }).map((_, i) => (
        <div key={i} className="bg-card p-3 rounded-lg shadow flex items-center justify-between">
          <div className="flex items-center gap-3">
            <Skeleton className="h-8 w-8 rounded-full" />
            <div className="space-y-1">
              <Skeleton className="h-5 w-32" />
              <Skeleton className="h-4 w-20" />
            </div>
          </div>
          <div className="text-right space-y-1">
            <Skeleton className="h-5 w-16 ml-auto" />
            <Skeleton className="h-4 w-12 ml-auto" />
          </div>
        </div>
      ))}
    </div>
  );
}

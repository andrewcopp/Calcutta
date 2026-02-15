import { Skeleton } from '../ui/Skeleton';

export function LeaderboardSkeleton() {
  return (
    <div className="grid gap-4">
      {Array.from({ length: 5 }).map((_, i) => (
        <div key={i} className="bg-white p-4 rounded-lg shadow">
          <div className="flex justify-between items-center">
            <div className="space-y-2">
              <Skeleton className="h-6 w-40" />
            </div>
            <div className="text-right space-y-2">
              <Skeleton className="h-7 w-24" />
              <Skeleton className="h-4 w-16 ml-auto" />
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}

import { Card } from '../ui/Card';
import { Skeleton } from '../ui/Skeleton';

export function BiddingSkeleton() {
  return (
    <div className="space-y-4">
      {/* Budget tracker skeleton */}
      <Card>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div>
            <Skeleton className="h-4 w-16 mb-2" />
            <Skeleton className="h-8 w-40" />
            <Skeleton className="h-2 w-full mt-2 rounded-full" />
          </div>
          <div>
            <Skeleton className="h-4 w-24 mb-2" />
            <Skeleton className="h-8 w-24" />
          </div>
          <div>
            <Skeleton className="h-4 w-12 mb-2" />
            <Skeleton className="h-8 w-32" />
          </div>
        </div>
      </Card>

      {/* Slot list skeleton */}
      <Card>
        <div className="mb-4">
          <Skeleton className="h-6 w-28 mb-2" />
          <Skeleton className="h-4 w-64" />
        </div>
        <div className="space-y-3">
          {Array.from({ length: 8 }).map((_, i) => (
            <div key={i} className="flex items-center gap-3 py-3">
              <Skeleton className="h-7 w-7 rounded-full shrink-0" />
              <Skeleton className="h-10 flex-1" />
            </div>
          ))}
        </div>
      </Card>
    </div>
  );
}

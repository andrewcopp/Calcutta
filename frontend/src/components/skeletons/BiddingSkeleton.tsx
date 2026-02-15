import { Card } from '../ui/Card';
import { Skeleton } from '../ui/Skeleton';

export function BiddingSkeleton() {
  return (
    <div className="space-y-4">
      <Card>
        <div className="flex items-center justify-between mb-4">
          <Skeleton className="h-6 w-32" />
          <Skeleton className="h-8 w-24" />
        </div>
        <div className="grid gap-3">
          {Array.from({ length: 8 }).map((_, i) => (
            <div key={i} className="flex items-center justify-between py-2">
              <div className="flex items-center gap-3">
                <Skeleton className="h-6 w-6 rounded" />
                <Skeleton className="h-5 w-40" />
              </div>
              <Skeleton className="h-8 w-20" />
            </div>
          ))}
        </div>
      </Card>
      <Skeleton className="h-10 w-full" />
    </div>
  );
}

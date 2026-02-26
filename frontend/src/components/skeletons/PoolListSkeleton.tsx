import { Card } from '../ui/Card';
import { Skeleton } from '../ui/Skeleton';

export function PoolListSkeleton() {
  return (
    <div className="grid gap-4">
      {Array.from({ length: 3 }).map((_, i) => (
        <Card key={i}>
          <div className="flex items-center justify-between">
            <div className="space-y-2">
              <Skeleton className="h-6 w-48" />
              <Skeleton className="h-4 w-32" />
            </div>
            <div className="text-right space-y-2">
              <Skeleton className="h-5 w-20 ml-auto" />
              <Skeleton className="h-4 w-16 ml-auto" />
            </div>
          </div>
        </Card>
      ))}
    </div>
  );
}

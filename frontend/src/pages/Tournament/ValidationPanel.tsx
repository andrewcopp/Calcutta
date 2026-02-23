import { Card } from '../../components/ui/Card';

interface ValidationStats {
  total: number;
  playIns: number;
  perRegion: Record<string, number>;
  duplicates: string[];
}

interface ValidationPanelProps {
  stats: ValidationStats;
  regionNames: string[];
}

export function ValidationPanel({ stats, regionNames }: ValidationPanelProps) {
  return (
    <Card className="mb-6">
      <div className="flex flex-wrap gap-6 text-sm">
        <div>
          <span className="text-muted-foreground">Total: </span>
          <span className={stats.total === 68 ? 'font-bold text-success' : 'font-bold text-amber-600'}>
            {stats.total}/68
          </span>
        </div>
        {regionNames.map((region) => {
          const count = stats.perRegion[region] || 0;
          return (
            <div key={region}>
              <span className="text-muted-foreground">{region}: </span>
              <span className={count >= 16 ? 'font-semibold text-success' : 'font-semibold text-amber-600'}>
                {count}
              </span>
            </div>
          );
        })}
        <div>
          <span className="text-muted-foreground">Play-ins: </span>
          <span className={stats.playIns === 4 ? 'font-bold text-success' : 'font-bold text-amber-600'}>
            {stats.playIns}/4
          </span>
        </div>
        {stats.duplicates.length > 0 && (
          <div className="text-destructive font-semibold">Duplicates: {stats.duplicates.join(', ')}</div>
        )}
      </div>
    </Card>
  );
}

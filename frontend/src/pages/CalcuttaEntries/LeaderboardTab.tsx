import { useMemo, useState } from 'react';
import { Link } from 'react-router-dom';
import { CalcuttaEntry } from '../../schemas/calcutta';
import { formatDollarsFromCents } from '../../utils/format';

type SortMode = 'actual' | 'projected';

interface LeaderboardTabProps {
  calcuttaId: string;
  entries: CalcuttaEntry[];
}

export function LeaderboardTab({ calcuttaId, entries }: LeaderboardTabProps) {
  const [sortMode, setSortMode] = useState<SortMode>('actual');

  const hasProjections = entries.some((e) => e.projectedEv != null);

  const sortedEntries = useMemo(() => {
    if (sortMode === 'projected') {
      return [...entries].sort((a, b) => {
        const aVal = a.projectedEv ?? a.totalPoints ?? 0;
        const bVal = b.projectedEv ?? b.totalPoints ?? 0;
        return bVal - aVal;
      });
    }
    return entries;
  }, [entries, sortMode]);

  return (
    <div className="grid gap-4">
      {hasProjections && (
        <div className="flex gap-1 rounded-lg bg-muted p-1">
          <button
            onClick={() => setSortMode('actual')}
            className={`flex-1 rounded-md px-3 py-1.5 text-sm font-medium transition-colors ${
              sortMode === 'actual'
                ? 'bg-primary text-primary-foreground shadow-sm'
                : 'text-muted-foreground hover:text-foreground'
            }`}
          >
            Actual Points
          </button>
          <button
            onClick={() => setSortMode('projected')}
            className={`flex-1 rounded-md px-3 py-1.5 text-sm font-medium transition-colors ${
              sortMode === 'projected'
                ? 'bg-primary text-primary-foreground shadow-sm'
                : 'text-muted-foreground hover:text-foreground'
            }`}
          >
            Projected Finish
          </button>
        </div>
      )}

      {sortedEntries.map((entry, index) => {
        const displayPosition =
          sortMode === 'actual'
            ? entry.finishPosition || index + 1
            : index + 1;
        const isInTheMoney = sortMode === 'actual' && Boolean(entry.inTheMoney);
        const payoutText = entry.payoutCents ? `(${formatDollarsFromCents(entry.payoutCents)})` : '';

        const rowClass = isInTheMoney
          ? displayPosition === 1
            ? 'bg-gradient-to-r from-yellow-50 to-yellow-100 border-2 border-yellow-400'
            : displayPosition === 2
              ? 'bg-gradient-to-r from-slate-50 to-slate-200 border-2 border-slate-400'
              : displayPosition === 3
                ? 'bg-gradient-to-r from-amber-50 to-amber-100 border-2 border-amber-500'
                : 'bg-gradient-to-r from-slate-50 to-blue-50 border-2 border-slate-300'
          : 'bg-card border border-gray-100';

        const pointsClass = isInTheMoney
          ? displayPosition === 1
            ? 'text-yellow-700'
            : displayPosition === 2
              ? 'text-slate-700'
              : displayPosition === 3
                ? 'text-amber-700'
                : 'text-slate-700'
          : 'text-primary';

        const displayValue =
          sortMode === 'projected' && entry.projectedEv != null
            ? entry.projectedEv.toFixed(2)
            : (entry.totalPoints ?? 0).toFixed(2);

        const displayLabel = sortMode === 'projected' ? 'projected' : 'points';

        return (
          <Link
            key={entry.id}
            to={`/calcuttas/${calcuttaId}/entries/${entry.id}`}
            className={`block p-4 rounded-lg shadow hover:shadow-md transition-shadow ${rowClass}`}
          >
            <div className="flex justify-between items-center">
              <div>
                <h2 className="text-xl font-semibold">
                  {displayPosition}. {entry.name}
                  {sortMode === 'actual' && isInTheMoney && payoutText && (
                    <span className="ml-2 text-sm text-foreground">{payoutText}</span>
                  )}
                </h2>
              </div>
              <div className="text-right">
                <p className={`text-2xl font-bold ${pointsClass}`}>
                  {displayValue} {displayLabel}
                </p>
              </div>
            </div>
          </Link>
        );
      })}
    </div>
  );
}

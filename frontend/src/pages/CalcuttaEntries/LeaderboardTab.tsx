import { useMemo, useState } from 'react';
import { Link } from 'react-router-dom';
import type { CalcuttaDashboard, CalcuttaEntry } from '../../schemas/calcutta';
import { formatDollarsFromCents } from '../../utils/format';
import { getRoundOptions } from '../../utils/roundLabels';
import { useLeaderboardStandings } from '../../hooks/useLeaderboardStandings';
import { Select } from '../../components/ui/Select';

type SortMode = 'actual' | 'projected';

interface LeaderboardTabProps {
  calcuttaId: string;
  entries: CalcuttaEntry[];
  dashboard: CalcuttaDashboard;
}

export function LeaderboardTab({ calcuttaId, entries, dashboard }: LeaderboardTabProps) {
  const [sortMode, setSortMode] = useState<SortMode>('actual');
  const [throughRound, setThroughRound] = useState<number | null>(null);

  const rewindEntries = useLeaderboardStandings(dashboard, throughRound);
  const displayEntries = rewindEntries ?? entries;

  const isHistorical = throughRound !== null;
  const effectiveSortMode = isHistorical ? 'actual' : sortMode;

  const hasProjections = entries.some((e) => e.projectedEv != null);

  const roundOptions = useMemo(
    () => getRoundOptions(dashboard.scoringRules),
    [dashboard.scoringRules],
  );

  const sortedEntries = useMemo(() => {
    if (effectiveSortMode === 'projected') {
      return [...displayEntries].sort((a, b) => {
        const aVal = a.projectedEv ?? a.totalPoints ?? 0;
        const bVal = b.projectedEv ?? b.totalPoints ?? 0;
        return bVal - aVal;
      });
    }
    return displayEntries;
  }, [displayEntries, effectiveSortMode]);

  return (
    <div className="grid gap-4">
      <div className="flex gap-3 items-center">
        {hasProjections && (
          <Select
            value={effectiveSortMode}
            onChange={(e) => setSortMode(e.target.value as SortMode)}
            disabled={isHistorical}
            className="w-auto"
          >
            <option value="actual">Actual Points</option>
            <option value="projected">Projected Finish</option>
          </Select>
        )}

        {roundOptions.length > 1 && (
          <Select
            value={throughRound === null ? 'current' : String(throughRound)}
            onChange={(e) => {
              const val = e.target.value;
              setThroughRound(val === 'current' ? null : Number(val));
            }}
            className="w-auto"
          >
            {roundOptions.map((opt) => (
              <option key={opt.value === null ? 'current' : opt.value} value={opt.value === null ? 'current' : opt.value}>
                {opt.label}
              </option>
            ))}
          </Select>
        )}
      </div>

      {sortedEntries.map((entry, index) => {
        const displayPosition =
          effectiveSortMode === 'actual'
            ? entry.finishPosition || index + 1
            : index + 1;
        const isInTheMoney = effectiveSortMode === 'actual' && Boolean(entry.inTheMoney);
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
          effectiveSortMode === 'projected' && entry.projectedEv != null
            ? entry.projectedEv.toFixed(2)
            : (entry.totalPoints ?? 0).toFixed(2);

        const displayLabel = effectiveSortMode === 'projected' ? 'projected' : 'points';

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
                  {effectiveSortMode === 'actual' && isInTheMoney && payoutText && (
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

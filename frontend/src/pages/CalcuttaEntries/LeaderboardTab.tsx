import { Link } from 'react-router-dom';
import { CalcuttaEntry } from '../../types/calcutta';
import { formatDollarsFromCents } from '../../utils/format';

interface LeaderboardTabProps {
  calcuttaId: string;
  entries: CalcuttaEntry[];
}

export function LeaderboardTab({ calcuttaId, entries }: LeaderboardTabProps) {
  return (
    <div className="grid gap-4">
      {entries.map((entry, index) => {
        const displayPosition = entry.finishPosition || index + 1;
        const isInTheMoney = Boolean(entry.inTheMoney);
        const payoutText = entry.payoutCents ? `(${formatDollarsFromCents(entry.payoutCents)})` : '';

        const rowClass = isInTheMoney
          ? displayPosition === 1
            ? 'bg-gradient-to-r from-yellow-50 to-yellow-100 border-2 border-yellow-400'
            : displayPosition === 2
              ? 'bg-gradient-to-r from-slate-50 to-slate-200 border-2 border-slate-400'
              : displayPosition === 3
                ? 'bg-gradient-to-r from-amber-50 to-amber-100 border-2 border-amber-500'
                : 'bg-gradient-to-r from-slate-50 to-blue-50 border-2 border-slate-300'
          : 'bg-white border border-gray-100';

        const pointsClass = isInTheMoney
          ? displayPosition === 1
            ? 'text-yellow-700'
            : displayPosition === 2
              ? 'text-slate-700'
              : displayPosition === 3
                ? 'text-amber-700'
                : 'text-slate-700'
          : 'text-blue-600';

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
                  {isInTheMoney && payoutText && <span className="ml-2 text-sm text-gray-700">{payoutText}</span>}
                </h2>
              </div>
              <div className="text-right">
                <p className={`text-2xl font-bold ${pointsClass}`}>
                  {entry.totalPoints ? entry.totalPoints.toFixed(2) : '0.00'} pts
                </p>
              </div>
            </div>
          </Link>
        );
      })}
    </div>
  );
}

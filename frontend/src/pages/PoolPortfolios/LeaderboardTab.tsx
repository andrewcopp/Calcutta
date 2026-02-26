import { useMemo, useState } from 'react';
import { Link } from 'react-router-dom';
import type { PoolDashboard, Portfolio } from '../../schemas/pool';
import { formatDollarsFromCents } from '../../utils/format';
import { getRoundOptions } from '../../utils/roundLabels';
import { Select } from '../../components/ui/Select';

type SortMode = 'actual' | 'projected' | 'favorites';

interface LeaderboardTabProps {
  poolId: string;
  portfolios: Portfolio[];
  dashboard: PoolDashboard;
}

export function LeaderboardTab({ poolId, portfolios, dashboard }: LeaderboardTabProps) {
  const [sortMode, setSortMode] = useState<SortMode>('actual');
  const [throughRound, setThroughRound] = useState<number | null>(null);

  const effectiveSortMode = sortMode;

  const roundOptions = useMemo(
    () => getRoundOptions(dashboard.roundStandings.map((g) => g.round)),
    [dashboard.roundStandings],
  );

  const displayPortfolios = useMemo(() => {
    if (throughRound === null) return portfolios;

    const roundGroup = dashboard.roundStandings.find((g) => g.round === throughRound);
    if (!roundGroup) return portfolios;

    const portfolioById = new Map(portfolios.map((e) => [e.id, e]));
    return roundGroup.entries.map((s) => {
      const portfolio = portfolioById.get(s.portfolioId);
      return {
        ...(portfolio ?? { id: s.portfolioId, name: s.portfolioId, poolId: '', createdAt: '', updatedAt: '' }),
        totalReturns: s.totalReturns,
        finishPosition: s.finishPosition,
        isTied: s.isTied,
        payoutCents: s.payoutCents,
        inTheMoney: s.inTheMoney,
        expectedValue: s.expectedValue,
        projectedFavorites: s.projectedFavorites,
      };
    });
  }, [portfolios, dashboard.roundStandings, throughRound]);

  const hasProjections = displayPortfolios.some((e) => e.expectedValue != null);
  const hasFavorites = displayPortfolios.some((e) => e.projectedFavorites != null);

  const sortedPortfolios = useMemo(() => {
    if (effectiveSortMode === 'projected') {
      return [...displayPortfolios].sort((a, b) => {
        const aVal = a.expectedValue ?? a.totalReturns ?? 0;
        const bVal = b.expectedValue ?? b.totalReturns ?? 0;
        return bVal - aVal;
      });
    }
    if (effectiveSortMode === 'favorites') {
      return [...displayPortfolios].sort((a, b) => {
        const aVal = a.projectedFavorites ?? a.totalReturns ?? 0;
        const bVal = b.projectedFavorites ?? b.totalReturns ?? 0;
        return bVal - aVal;
      });
    }
    return displayPortfolios;
  }, [displayPortfolios, effectiveSortMode]);

  return (
    <div className="grid gap-4">
      <div className="flex gap-3 items-center">
        {(hasProjections || hasFavorites) && (
          <Select
            value={effectiveSortMode}
            onChange={(e) => setSortMode(e.target.value as SortMode)}
            className="w-auto"
          >
            <option value="actual">Actual Returns</option>
            {hasProjections && <option value="projected">Projected EV</option>}
            {hasFavorites && <option value="favorites">Projected Favorites</option>}
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

      {sortedPortfolios.map((portfolio, index) => {
        const displayPosition =
          effectiveSortMode === 'actual'
            ? portfolio.finishPosition || index + 1
            : index + 1;
        const isInTheMoney = effectiveSortMode === 'actual' && Boolean(portfolio.inTheMoney);
        const payoutText = portfolio.payoutCents ? `(${formatDollarsFromCents(portfolio.payoutCents)})` : '';

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
          effectiveSortMode === 'projected' && portfolio.expectedValue != null
            ? portfolio.expectedValue.toFixed(2)
            : effectiveSortMode === 'favorites' && portfolio.projectedFavorites != null
              ? portfolio.projectedFavorites.toFixed(2)
              : (portfolio.totalReturns ?? 0).toFixed(2);

        const displayLabel =
          effectiveSortMode === 'projected'
            ? 'projected'
            : effectiveSortMode === 'favorites'
              ? 'favorites'
              : 'returns';

        return (
          <Link
            key={portfolio.id}
            to={`/pools/${poolId}/portfolios/${portfolio.id}`}
            className={`block p-4 rounded-lg shadow hover:shadow-md transition-shadow ${rowClass}`}
          >
            <div className="flex justify-between items-center">
              <div>
                <h2 className="text-xl font-semibold">
                  {displayPosition}. {portfolio.name}
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

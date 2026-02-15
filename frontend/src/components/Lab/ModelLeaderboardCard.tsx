import React from 'react';
import { useNavigate } from 'react-router-dom';
import { cn } from '../../lib/cn';
import { PipelineProgressBar } from './PipelineProgressBar';
import type { LeaderboardEntry } from '../../services/labService';

type ModelLeaderboardCardProps = {
  entry: LeaderboardEntry;
  rank: number;
  totalCalcuttas: number;
};

function formatPct(val?: number | null): string {
  if (val == null) return '-';
  return `${(val * 100).toFixed(1)}%`;
}

function formatPayoutX(val?: number | null): string {
  if (val == null) return '-';
  return `${val.toFixed(2)}x`;
}

function getPayoutColorClass(payout?: number | null): string {
  if (payout == null) return 'text-gray-500';
  if (payout >= 1.2) return 'text-green-700';
  if (payout < 0.9) return 'text-red-600';
  return 'text-gray-700';
}

function getTop1ColorClass(p?: number | null): string {
  if (p == null) return 'text-gray-500';
  if (p >= 0.15) return 'text-green-700';
  if (p >= 0.08) return 'text-blue-600';
  return 'text-gray-700';
}

export function ModelLeaderboardCard({ entry, rank, totalCalcuttas }: ModelLeaderboardCardProps) {
  const navigate = useNavigate();

  const hasPredictions = entry.n_entries_with_predictions > 0;
  const hasEntries = entry.n_calcuttas_with_entries > 0;
  const hasEvaluations = entry.n_calcuttas_with_evaluations > 0;

  const coverageText = totalCalcuttas > 0
    ? `${entry.n_calcuttas_with_evaluations}/${totalCalcuttas}`
    : `${entry.n_calcuttas_with_evaluations}`;

  return (
    <div
      role="button"
      tabIndex={0}
      onClick={() => navigate(`/lab/models/${encodeURIComponent(entry.investment_model_id)}`)}
      onKeyDown={(e) => {
        if (e.key === 'Enter' || e.key === ' ') {
          e.preventDefault();
          navigate(`/lab/models/${encodeURIComponent(entry.investment_model_id)}`);
        }
      }}
      className="bg-white border border-gray-200 rounded-lg p-4 hover:border-gray-300 hover:shadow-sm transition-all cursor-pointer"
    >
      {/* Desktop: single row */}
      <div className="flex items-center gap-4">
        {/* Rank */}
        <div className="flex-shrink-0 w-8 h-8 rounded-full bg-gray-100 flex items-center justify-center">
          <span className="text-sm font-semibold text-gray-700">{rank}</span>
        </div>

        {/* Model info */}
        <div className="flex-1 min-w-0">
          <h3 className="text-sm font-semibold text-gray-900 truncate">{entry.model_name}</h3>
          <p className="text-xs text-gray-500">{entry.model_kind}</p>
        </div>

        {/* Progress bar - hidden on mobile */}
        <div className="hidden md:block flex-shrink-0 w-48">
          <PipelineProgressBar hasPredictions={hasPredictions} hasEntries={hasEntries} hasEvaluations={hasEvaluations} />
        </div>

        {/* Coverage - hidden on mobile */}
        <div className="hidden md:block flex-shrink-0 w-16 text-center">
          <p className="text-xs text-gray-500">Coverage</p>
          <p className="text-sm font-medium text-gray-700">{coverageText}</p>
        </div>

        {/* P(Top 1) - always visible, primary metric */}
        <div className="flex-shrink-0 w-20 text-center">
          <p className="text-xs text-gray-500">P(Top 1)</p>
          <p className={cn('text-lg font-bold', getTop1ColorClass(entry.avg_p_top1))}>
            {formatPct(entry.avg_p_top1)}
          </p>
        </div>

        {/* Avg Payout - hidden on mobile */}
        <div className="hidden md:block flex-shrink-0 w-20 text-center">
          <p className="text-xs text-gray-500">Avg Payout</p>
          <p className={cn('text-sm font-medium', getPayoutColorClass(entry.avg_mean_payout))}>
            {formatPayoutX(entry.avg_mean_payout)}
          </p>
        </div>

        {/* P(In Money) - hidden on mobile */}
        <div className="hidden md:block flex-shrink-0 w-16 text-center">
          <p className="text-xs text-gray-500">P(In $)</p>
          <p className="text-sm font-medium text-gray-700">{formatPct(entry.avg_p_in_money)}</p>
        </div>
      </div>

      {/* Mobile: metrics row below */}
      <div className="flex items-center gap-3 mt-3 pt-3 border-t border-gray-100 md:hidden">
        <div className="flex-1">
          <PipelineProgressBar hasPredictions={hasPredictions} hasEntries={hasEntries} hasEvaluations={hasEvaluations} />
        </div>
        <div className="text-center">
          <p className="text-xs text-gray-500">Payout</p>
          <p className={cn('text-xs font-semibold', getPayoutColorClass(entry.avg_mean_payout))}>
            {formatPayoutX(entry.avg_mean_payout)}
          </p>
        </div>
        <div className="text-center">
          <p className="text-xs text-gray-500">P(In $)</p>
          <p className="text-xs font-medium text-gray-700">{formatPct(entry.avg_p_in_money)}</p>
        </div>
      </div>
    </div>
  );
}

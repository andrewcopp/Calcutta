import React from 'react';
import { cn } from '../../lib/cn';

export type PipelineStatus = 'complete' | 'pending' | 'missing';

type PipelineStatusCellProps = {
  status: PipelineStatus;
  label?: string;
  metric?: number | null;
  metricFormat?: 'payout' | 'percent';
  className?: string;
  onClick?: () => void;
};

const statusConfig: Record<PipelineStatus, { icon: string; colorClass: string }> = {
  complete: { icon: '✓', colorClass: 'text-green-600' },
  pending: { icon: '⏳', colorClass: 'text-amber-600' },
  missing: { icon: '−', colorClass: 'text-gray-400' },
};

function formatMetric(value: number | null | undefined, format: 'payout' | 'percent'): string {
  if (value == null) return '';
  if (format === 'payout') {
    return `${value.toFixed(2)}x`;
  }
  return `${(value * 100).toFixed(1)}%`;
}

function getPayoutColorClass(payout: number): string {
  if (payout >= 1.2) return 'text-green-700 font-semibold';
  if (payout < 0.9) return 'text-red-600';
  return 'text-gray-700';
}

export function PipelineStatusCell({
  status,
  label,
  metric,
  metricFormat = 'payout',
  className,
  onClick,
}: PipelineStatusCellProps) {
  const config = statusConfig[status];
  const isClickable = onClick != null;

  const content = (
    <div className="flex items-center gap-2">
      <span className={cn('text-base', config.colorClass)}>{config.icon}</span>
      {label ? <span className="text-sm text-gray-600">{label}</span> : null}
      {metric != null && status === 'complete' ? (
        <span
          className={cn(
            'text-sm ml-auto',
            metricFormat === 'payout' ? getPayoutColorClass(metric) : 'text-gray-700'
          )}
        >
          {formatMetric(metric, metricFormat)}
        </span>
      ) : null}
      {status === 'missing' ? <span className="text-sm text-gray-400 ml-auto">−</span> : null}
    </div>
  );

  if (isClickable) {
    return (
      <button
        type="button"
        onClick={onClick}
        className={cn(
          'w-full text-left px-3 py-2 rounded hover:bg-gray-50 transition-colors',
          className
        )}
      >
        {content}
      </button>
    );
  }

  return <div className={cn('px-3 py-2', className)}>{content}</div>;
}

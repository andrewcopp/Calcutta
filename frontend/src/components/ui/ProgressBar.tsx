import React from 'react';

import { cn } from '../../lib/cn';

type ProgressBarProps = {
  value: number; // 0 to 1
  className?: string;
  showLabel?: boolean;
  size?: 'sm' | 'md';
};

export function ProgressBar({ value, className, showLabel = false, size = 'md' }: ProgressBarProps) {
  const percent = Math.max(0, Math.min(100, Math.round(value * 100)));
  const heightClass = size === 'sm' ? 'h-1.5' : 'h-2.5';

  return (
    <div className={cn('w-full', className)}>
      <div className={cn('w-full bg-gray-200 rounded-full overflow-hidden', heightClass)}>
        <div
          className={cn('bg-blue-600 rounded-full transition-all duration-300', heightClass)}
          style={{ width: `${percent}%` }}
        />
      </div>
      {showLabel && (
        <div className="text-xs text-gray-600 mt-1">{percent}%</div>
      )}
    </div>
  );
}

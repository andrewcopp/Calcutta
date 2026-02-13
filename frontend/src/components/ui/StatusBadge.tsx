import React from 'react';

import { cn } from '../../lib/cn';

type StatusBadgeProps = {
  status: string;
  className?: string;
};

const statusStyles: Record<string, string> = {
  succeeded: 'bg-green-100 text-green-800',
  completed: 'bg-green-100 text-green-800',
  failed: 'bg-red-100 text-red-800',
  error: 'bg-red-100 text-red-800',
  pending: 'bg-gray-100 text-gray-800',
  queued: 'bg-gray-100 text-gray-800',
  running: 'bg-yellow-100 text-yellow-800',
  in_progress: 'bg-yellow-100 text-yellow-800',
  claimed: 'bg-blue-100 text-blue-800',
};

export function StatusBadge({ status, className }: StatusBadgeProps) {
  const normalizedStatus = status.toLowerCase().replace(/-/g, '_');
  const style = statusStyles[normalizedStatus] || 'bg-gray-100 text-gray-800';

  return (
    <span
      className={cn(
        'inline-flex items-center px-2 py-0.5 rounded text-xs font-medium',
        style,
        className
      )}
    >
      {status}
    </span>
  );
}

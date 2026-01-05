import React from 'react';

import { cn } from '../../lib/cn';

type AlertVariant = 'info' | 'success' | 'warning' | 'error';

type AlertProps = React.HTMLAttributes<HTMLDivElement> & {
  variant?: AlertVariant;
};

const variantClasses: Record<AlertVariant, string> = {
  info: 'bg-blue-50 border-blue-200 text-blue-900',
  success: 'bg-green-50 border-green-200 text-green-900',
  warning: 'bg-yellow-50 border-yellow-200 text-yellow-900',
  error: 'bg-red-100 border-red-400 text-red-700',
};

export function Alert({ className, variant = 'info', ...props }: AlertProps) {
  return (
    <div
      role={variant === 'error' ? 'alert' : undefined}
      className={cn('border px-4 py-3 rounded-lg text-sm', variantClasses[variant], className)}
      {...props}
    />
  );
}

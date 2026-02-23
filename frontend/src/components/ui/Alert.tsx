import React from 'react';

import { cn } from '../../lib/cn';

type AlertVariant = 'info' | 'success' | 'warning' | 'error';

type AlertProps = React.HTMLAttributes<HTMLDivElement> & {
  variant?: AlertVariant;
};

const variantClasses: Record<AlertVariant, string> = {
  info: 'bg-info/10 border-info/20 text-info-foreground',
  success: 'bg-success/10 border-success/20 text-success',
  warning: 'bg-warning/10 border-warning/20 text-warning',
  error: 'bg-destructive/10 border-destructive/20 text-destructive',
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

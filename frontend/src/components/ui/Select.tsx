import React from 'react';

import { cn } from '../../lib/cn';

type SelectProps = React.SelectHTMLAttributes<HTMLSelectElement>;

export const Select = React.forwardRef<HTMLSelectElement, SelectProps>(({ className, ...props }, ref) => {
  return (
    <select
      ref={ref}
      className={cn(
        'h-10 w-full rounded-lg border border-border bg-card px-4 py-2 text-sm text-foreground outline-none focus:ring-2 focus:ring-primary focus:border-primary disabled:opacity-50',
        className,
      )}
      {...props}
    />
  );
});

Select.displayName = 'Select';

import React from 'react';

import { cn } from '../../lib/cn';
import { Spinner } from './Spinner';

type LoadingStateProps = {
  label?: string;
  size?: React.ComponentProps<typeof Spinner>['size'];
  className?: string;
  layout?: 'inline' | 'center';
};

export function LoadingState({ label = 'Loading...', size = 'md', className, layout = 'center' }: LoadingStateProps) {
  if (layout === 'inline') {
    return (
      <div className={cn('inline-flex items-center gap-2 text-sm text-muted-foreground', className)}>
        <Spinner size={size} />
        <span>{label}</span>
      </div>
    );
  }

  return (
    <div className={cn('flex flex-col items-center justify-center gap-3 py-8 text-muted-foreground', className)}>
      <Spinner size={size === 'sm' ? 'md' : size} />
      {label ? <div className="text-sm">{label}</div> : null}
    </div>
  );
}

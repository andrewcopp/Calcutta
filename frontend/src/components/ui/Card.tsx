import React from 'react';

import { cn } from '../../lib/cn';

type DivProps = React.HTMLAttributes<HTMLDivElement>;

export function Card({ className, ...props }: DivProps) {
  return <div className={cn('bg-surface rounded-lg shadow p-6', className)} {...props} />;
}

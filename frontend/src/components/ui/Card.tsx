import React from 'react';

import { cn } from '../../lib/cn';

type DivProps = React.HTMLAttributes<HTMLDivElement>;

export function Card({ className, ...props }: DivProps) {
  return <div className={cn('bg-surface rounded-lg shadow p-6', className)} {...props} />;
}

export function CardHeader({ className, ...props }: DivProps) {
  return <div className={cn('mb-4', className)} {...props} />;
}

export function CardBody({ className, ...props }: DivProps) {
  return <div className={cn('', className)} {...props} />;
}

export function CardFooter({ className, ...props }: DivProps) {
  return <div className={cn('mt-4', className)} {...props} />;
}

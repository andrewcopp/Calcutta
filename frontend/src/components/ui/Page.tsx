import React from 'react';

import { cn } from '../../lib/cn';

type PageContainerProps = {
  children: React.ReactNode;
  className?: string;
};

export function PageContainer({ children, className }: PageContainerProps) {
  return <div className={cn('container mx-auto px-4 py-8', className)}>{children}</div>;
}

type PageHeaderProps = {
  title: React.ReactNode;
  subtitle?: React.ReactNode;
  leftActions?: React.ReactNode;
  actions?: React.ReactNode;
  className?: string;
};

export function PageHeader({ title, subtitle, leftActions, actions, className }: PageHeaderProps) {
  return (
    <div className={cn('mb-8 flex items-start justify-between gap-6', className)}>
      <div className="min-w-0">
        {leftActions ? <div className="mb-2 flex items-center gap-2">{leftActions}</div> : null}
        <h1 className="text-3xl font-bold tracking-tight mb-2">{title}</h1>
        {subtitle ? <p className="text-gray-500">{subtitle}</p> : null}
      </div>
      {actions ? <div className="flex items-center gap-2">{actions}</div> : null}
    </div>
  );
}

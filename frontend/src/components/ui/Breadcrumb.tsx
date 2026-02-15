import React from 'react';
import { Link } from 'react-router-dom';

import { cn } from '../../lib/cn';

export type BreadcrumbItem = {
  label: string;
  href?: string;
};

type BreadcrumbProps = {
  items: BreadcrumbItem[];
  className?: string;
};

export function Breadcrumb({ items, className }: BreadcrumbProps) {
  if (items.length === 0) return null;

  return (
    <nav aria-label="Breadcrumb" className={cn('mb-4', className)}>
      <ol className="flex flex-wrap items-center gap-1 text-sm text-gray-600">
        {items.map((item, index) => {
          const isLast = index === items.length - 1;

          return (
            <li key={index} className="flex items-center">
              {index > 0 && (
                <span className="mx-2 text-gray-400" aria-hidden="true">
                  /
                </span>
              )}
              {isLast || !item.href ? (
                <span
                  className={cn(
                    'max-w-[200px] truncate',
                    isLast ? 'font-medium text-gray-900' : ''
                  )}
                  aria-current={isLast ? 'page' : undefined}
                  title={item.label}
                >
                  {item.label}
                </span>
              ) : (
                <Link
                  to={item.href}
                  className="max-w-[200px] truncate text-blue-600 hover:text-blue-800 hover:underline"
                  title={item.label}
                >
                  {item.label}
                </Link>
              )}
            </li>
          );
        })}
      </ol>
    </nav>
  );
}

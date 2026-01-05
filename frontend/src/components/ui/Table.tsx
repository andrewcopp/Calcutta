import React from 'react';

import { cn } from '../../lib/cn';

type TableProps = React.TableHTMLAttributes<HTMLTableElement> & {
  containerClassName?: string;
};

export function Table({ className, containerClassName, ...props }: TableProps) {
  return (
    <div className={cn('overflow-x-auto', containerClassName)}>
      <table className={cn('min-w-full divide-y divide-border', className)} {...props} />
    </div>
  );
}

type TheadProps = React.HTMLAttributes<HTMLTableSectionElement>;

type TbodyProps = React.HTMLAttributes<HTMLTableSectionElement>;

type TrProps = React.HTMLAttributes<HTMLTableRowElement>;

type ThProps = React.ThHTMLAttributes<HTMLTableCellElement>;

type TdProps = React.TdHTMLAttributes<HTMLTableCellElement>;

export function TableHead({ className, ...props }: TheadProps) {
  return <thead className={cn('bg-gray-50', className)} {...props} />;
}

export function TableBody({ className, ...props }: TbodyProps) {
  return <tbody className={cn('bg-surface divide-y divide-border', className)} {...props} />;
}

export function TableRow({ className, ...props }: TrProps) {
  return <tr className={cn('hover:bg-gray-50', className)} {...props} />;
}

export function TableHeaderCell({ className, ...props }: ThProps) {
  return (
    <th
      className={cn('px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider', className)}
      {...props}
    />
  );
}

export function TableCell({ className, ...props }: TdProps) {
  return <td className={cn('px-4 py-3 text-sm text-gray-700', className)} {...props} />;
}

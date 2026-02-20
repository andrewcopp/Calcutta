import { useState } from 'react';

import { cn } from '../../lib/cn';

interface CollapsibleProps {
  title: string;
  count: number;
  defaultOpen?: boolean;
  children: React.ReactNode;
}

export function Collapsible({ title, count, defaultOpen = false, children }: CollapsibleProps) {
  const [open, setOpen] = useState(defaultOpen);

  return (
    <div className="border-t border-gray-200 pt-4">
      <button
        type="button"
        onClick={() => setOpen((prev) => !prev)}
        className="flex w-full items-center gap-2 text-left text-gray-500 hover:text-gray-700"
      >
        <svg
          className={cn('h-4 w-4 shrink-0 transition-transform', open && 'rotate-90')}
          viewBox="0 0 20 20"
          fill="currentColor"
        >
          <path
            fillRule="evenodd"
            d="M7.21 14.77a.75.75 0 01.02-1.06L11.168 10 7.23 6.29a.75.75 0 111.04-1.08l4.5 4.25a.75.75 0 010 1.08l-4.5 4.25a.75.75 0 01-1.06-.02z"
            clipRule="evenodd"
          />
        </svg>
        <span className="text-sm font-medium">{title}</span>
        <span className="rounded-full bg-gray-100 px-2 py-0.5 text-xs text-gray-500">
          {count}
        </span>
      </button>
      {open && <div className="mt-4">{children}</div>}
    </div>
  );
}

import { useEffect, useState } from 'react';

import { cn } from '../../lib/cn';
import { toast as toastStore, type Toast as ToastType } from '../../lib/toast';

const variantClasses: Record<ToastType['variant'], string> = {
  success: 'bg-green-50 border-green-200 text-green-900',
  error: 'bg-red-100 border-red-400 text-red-700',
  info: 'bg-blue-50 border-blue-200 text-blue-900',
};

const EXIT_MS = 200;

export function Toast({ id, message, variant }: ToastType) {
  const [exiting, setExiting] = useState(false);

  useEffect(() => {
    // no-op: kept for future animation hooks
  }, []);

  const handleDismiss = () => {
    setExiting(true);
    setTimeout(() => toastStore.dismiss(id), EXIT_MS);
  };

  return (
    <div
      role={variant === 'error' ? 'alert' : 'status'}
      aria-live={variant === 'error' ? 'assertive' : 'polite'}
      className={cn(
        'pointer-events-auto flex items-center gap-2 border px-4 py-3 rounded-lg text-sm shadow-lg transition-all duration-200',
        exiting ? 'opacity-0 translate-x-4' : 'opacity-100 translate-x-0',
        variantClasses[variant],
      )}
    >
      <span className="flex-1">{message}</span>
      <button
        onClick={handleDismiss}
        className="shrink-0 opacity-60 hover:opacity-100"
        aria-label="Dismiss"
      >
        <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" strokeWidth="2" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    </div>
  );
}

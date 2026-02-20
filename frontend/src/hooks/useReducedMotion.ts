import { useMemo } from 'react';

/**
 * Returns true when the user has enabled the "prefers-reduced-motion: reduce"
 * media query. Safe for SSR -- returns false when `window` is unavailable.
 */
export function useReducedMotion(): boolean {
  return useMemo(() => {
    return (
      typeof window !== 'undefined' &&
      typeof window.matchMedia === 'function' &&
      window.matchMedia('(prefers-reduced-motion: reduce)').matches
    );
  }, []);
}

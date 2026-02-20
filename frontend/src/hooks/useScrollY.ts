import { useEffect, useRef, useState } from 'react';

/**
 * Tracks `window.scrollY` throttled via `requestAnimationFrame`.
 *
 * @param enabled - When false the listener is not attached and the
 *   returned value stays at 0. Pass `!prefersReducedMotion` to skip
 *   scroll tracking when the user has opted out of motion.
 * @returns The current vertical scroll offset in pixels.
 */
export function useScrollY(enabled = true): number {
  const [scrollY, setScrollY] = useState(0);
  const rafRef = useRef<number | null>(null);

  useEffect(() => {
    if (!enabled) return;

    const onScroll = () => {
      if (rafRef.current != null) return;
      rafRef.current = window.requestAnimationFrame(() => {
        rafRef.current = null;
        setScrollY(window.scrollY || 0);
      });
    };

    onScroll();
    window.addEventListener('scroll', onScroll, { passive: true });
    return () => {
      window.removeEventListener('scroll', onScroll);
      if (rafRef.current != null) {
        window.cancelAnimationFrame(rafRef.current);
        rafRef.current = null;
      }
    };
  }, [enabled]);

  return scrollY;
}

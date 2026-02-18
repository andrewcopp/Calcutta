import { useCallback, useEffect, useRef, useState } from 'react';

/**
 * Hook for auto-clearing flash messages.
 *
 * @param duration - milliseconds before the message auto-clears (default 3000)
 * @returns [message, flash] where flash(msg) sets the message and schedules auto-clear.
 */
export function useFlashMessage(duration = 3000): [string | null, (msg: string) => void] {
  const [message, setMessage] = useState<string | null>(null);
  const timerRef = useRef<ReturnType<typeof setTimeout>>();

  useEffect(() => {
    return () => { clearTimeout(timerRef.current); };
  }, []);

  const flash = useCallback((msg: string) => {
    clearTimeout(timerRef.current);
    setMessage(msg);
    timerRef.current = setTimeout(() => setMessage(null), duration);
  }, [duration]);

  return [message, flash];
}

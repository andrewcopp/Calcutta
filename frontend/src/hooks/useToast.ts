import { useSyncExternalStore } from 'react';

import { toast, subscribe, getSnapshot } from '../lib/toast';

export function useToast() {
  const toasts = useSyncExternalStore(subscribe, getSnapshot);
  return { toasts, toast };
}

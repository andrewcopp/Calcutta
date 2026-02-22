type ToastVariant = 'success' | 'error' | 'info';

export interface Toast {
  id: number;
  message: string;
  variant: ToastVariant;
}

const MAX_TOASTS = 3;
const AUTO_DISMISS_MS = 4000;

let nextId = 1;
let toasts: Toast[] = [];
let listeners: Array<() => void> = [];

function emit() {
  for (const listener of listeners) {
    listener();
  }
}

function add(message: string, variant: ToastVariant) {
  const id = nextId++;
  toasts = [{ id, message, variant }, ...toasts].slice(0, MAX_TOASTS);
  emit();
  setTimeout(() => dismiss(id), AUTO_DISMISS_MS);
}

function dismiss(id: number) {
  const prev = toasts;
  toasts = toasts.filter((t) => t.id !== id);
  if (toasts !== prev) emit();
}

export const toast = {
  success: (message: string) => add(message, 'success'),
  error: (message: string) => add(message, 'error'),
  info: (message: string) => add(message, 'info'),
  dismiss,
};

export function subscribe(listener: () => void) {
  listeners = [...listeners, listener];
  return () => {
    listeners = listeners.filter((l) => l !== listener);
  };
}

export function getSnapshot(): Toast[] {
  return toasts;
}

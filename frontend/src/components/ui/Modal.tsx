import React, { useEffect, useRef } from 'react';

import { cn } from '../../lib/cn';

type ModalProps = {
  open: boolean;
  onClose: () => void;
  title?: string;
  children: React.ReactNode;
  className?: string;
};

export function Modal({ open, onClose, title, children, className }: ModalProps) {
  const dialogRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;

    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };

    document.addEventListener('keydown', handleEscape);
    document.body.style.overflow = 'hidden';

    return () => {
      document.removeEventListener('keydown', handleEscape);
      document.body.style.overflow = '';
    };
  }, [open, onClose]);

  useEffect(() => {
    if (open && dialogRef.current) {
      const firstInput = dialogRef.current.querySelector<HTMLElement>('input, button, select, textarea');
      firstInput?.focus();
    }
  }, [open]);

  if (!open) return null;

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center"
      onClick={onClose}
      role="dialog"
      aria-modal="true"
      aria-labelledby={title ? 'modal-title' : undefined}
    >
      <div className="fixed inset-0 bg-black/50" aria-hidden="true" />
      <div
        ref={dialogRef}
        className={cn(
          'relative z-10 w-full max-w-md rounded-lg bg-white p-6 shadow-xl',
          className
        )}
        onClick={(e) => e.stopPropagation()}
      >
        {title && (
          <h2 id="modal-title" className="mb-4 text-lg font-semibold text-gray-900">
            {title}
          </h2>
        )}
        {children}
      </div>
    </div>
  );
}

type ModalActionsProps = {
  children: React.ReactNode;
  className?: string;
};

export function ModalActions({ children, className }: ModalActionsProps) {
  return (
    <div className={cn('mt-6 flex justify-end gap-3', className)}>
      {children}
    </div>
  );
}

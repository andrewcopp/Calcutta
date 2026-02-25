import React, { useState } from 'react';
import { Modal, ModalActions } from '../ui/Modal';
import { Button } from '../ui/Button';
import { Alert } from '../ui/Alert';

type ResetPasswordConfirmModalProps = {
  open: boolean;
  onClose: () => void;
  userId: string;
  userName: string;
  userEmail: string;
  onConfirm: (userId: string) => Promise<void>;
};

export function ResetPasswordConfirmModal({
  open,
  onClose,
  userId,
  userName,
  userEmail,
  onConfirm,
}: ResetPasswordConfirmModalProps) {
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleConfirm = async () => {
    setError('');
    setLoading(true);
    try {
      await onConfirm(userId);
      onClose();
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Failed to send reset email. Please try again.';
      if (message.toLowerCase().includes('rate') || message.toLowerCase().includes('too many')) {
        setError('Reset email was sent recently. Please wait before sending again.');
      } else {
        setError(message);
      }
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    setError('');
    onClose();
  };

  return (
    <Modal open={open} onClose={handleClose} title="Reset Password">
      <p className="text-sm text-gray-600 mb-4">
        Send a password reset email to <strong>{userName}</strong> at <strong>{userEmail}</strong>?
      </p>
      <p className="text-sm text-gray-500 mb-4">
        This will generate a reset link that expires in 30 minutes. Any existing sessions will be revoked when the
        password is changed.
      </p>

      {error && (
        <Alert variant="error" className="mb-4">
          {error}
        </Alert>
      )}

      <ModalActions>
        <Button type="button" variant="secondary" onClick={handleClose} disabled={loading}>
          Cancel
        </Button>
        <Button type="button" onClick={handleConfirm} disabled={loading}>
          {loading ? 'Sending...' : 'Send Reset Email'}
        </Button>
      </ModalActions>
    </Modal>
  );
}

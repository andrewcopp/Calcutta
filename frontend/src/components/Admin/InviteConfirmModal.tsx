import React, { useState } from 'react';
import { Modal, ModalActions } from '../ui/Modal';
import { Button } from '../ui/Button';
import { Alert } from '../ui/Alert';

type InviteConfirmModalProps = {
  open: boolean;
  onClose: () => void;
  userId: string;
  userName: string;
  userEmail: string;
  lastInviteSentAt: string | null;
  onConfirm: (userId: string) => Promise<void>;
};

function formatRelativeTime(dateStr: string): string {
  const date = new Date(dateStr);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMs / 3600000);
  const diffDays = Math.floor(diffMs / 86400000);

  if (diffMins < 1) return 'just now';
  if (diffMins < 60) return `${diffMins} minute${diffMins === 1 ? '' : 's'} ago`;
  if (diffHours < 24) return `${diffHours} hour${diffHours === 1 ? '' : 's'} ago`;
  return `${diffDays} day${diffDays === 1 ? '' : 's'} ago`;
}

export function InviteConfirmModal({
  open,
  onClose,
  userId,
  userName,
  userEmail,
  lastInviteSentAt,
  onConfirm,
}: InviteConfirmModalProps) {
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const recentlySent = lastInviteSentAt && Date.now() - new Date(lastInviteSentAt).getTime() < 60000;

  const handleConfirm = async () => {
    setError('');
    setLoading(true);
    try {
      await onConfirm(userId);
      onClose();
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Failed to send invite. Please try again.';
      if (message.toLowerCase().includes('rate') || message.toLowerCase().includes('too many')) {
        setError('Invite was sent recently. Please wait before sending again.');
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
    <Modal open={open} onClose={handleClose} title="Send Invite">
      <p className="text-sm text-gray-600 mb-4">
        Send an invite email to <strong>{userName}</strong> at <strong>{userEmail}</strong>?
      </p>

      {lastInviteSentAt && (
        <p className="text-sm text-gray-500 mb-4">Last invite sent: {formatRelativeTime(lastInviteSentAt)}</p>
      )}

      {recentlySent && (
        <Alert variant="warning" className="mb-4">
          An invite was sent very recently. You may want to wait before sending again.
        </Alert>
      )}

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
          {loading ? 'Sending...' : lastInviteSentAt ? 'Resend Invite' : 'Send Invite'}
        </Button>
      </ModalActions>
    </Modal>
  );
}

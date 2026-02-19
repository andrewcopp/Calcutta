import React, { useState } from 'react';
import { Modal, ModalActions } from '../ui/Modal';
import { Input } from '../ui/Input';
import { Button } from '../ui/Button';
import { Alert } from '../ui/Alert';

type SetEmailModalProps = {
  open: boolean;
  onClose: () => void;
  userId: string;
  userName: string;
  onSubmit: (userId: string, email: string) => Promise<void>;
};

export function SetEmailModal({ open, onClose, userId, userName, onSubmit }: SetEmailModalProps) {
  const [email, setEmail] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    const trimmed = email.trim();
    if (!trimmed) {
      setError('Email is required.');
      return;
    }
    if (!trimmed.includes('@')) {
      setError('Please enter a valid email address.');
      return;
    }

    setLoading(true);
    try {
      await onSubmit(userId, trimmed);
      setEmail('');
      onClose();
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Failed to set email. Please try again.';
      setError(message);
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    setEmail('');
    setError('');
    onClose();
  };

  return (
    <Modal open={open} onClose={handleClose} title="Set Email Address">
      <form onSubmit={handleSubmit}>
        <p className="text-sm text-gray-600 mb-4">
          Set the email address for <strong>{userName}</strong>. This will transition the user to "invited" status.
        </p>

        <div className="mb-4">
          <label htmlFor="email" className="block text-sm font-medium text-gray-700 mb-1">
            Email Address
          </label>
          <Input
            type="email"
            id="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="user@example.com"
            autoFocus
          />
        </div>

        {error && (
          <Alert variant="error" className="mb-4">{error}</Alert>
        )}

        <ModalActions>
          <Button type="button" variant="secondary" onClick={handleClose} disabled={loading}>
            Cancel
          </Button>
          <Button type="submit" disabled={loading || !email.trim()}>
            {loading ? 'Saving...' : 'Set Email'}
          </Button>
        </ModalActions>
      </form>
    </Modal>
  );
}

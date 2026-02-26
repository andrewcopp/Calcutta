import React, { useState, useEffect } from 'react';
import { Link, useNavigate, useSearchParams, useLocation } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { useUser } from '../contexts/UserContext';
import { userService } from '../services/userService';
import { toast } from '../lib/toast';
import { formatDate } from '../utils/format';
import { Input } from '../components/ui/Input';
import { Button } from '../components/ui/Button';
import { Alert } from '../components/ui/Alert';

export function AcceptInvitePage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const location = useLocation();
  const token = searchParams.get('token') ?? '';
  const from = (location.state as { from?: string })?.from ?? '/pools';
  const { user, acceptInvite, logout } = useUser();

  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [showLogoutConfirm, setShowLogoutConfirm] = useState(false);
  const [inviteConsumed, setInviteConsumed] = useState(false);

  const previewQuery = useQuery({
    queryKey: ['invite-preview', token],
    queryFn: () => userService.previewInvite(token),
    enabled: !!token && !user,
    retry: false,
  });
  const preview = previewQuery.data;

  useEffect(() => {
    if (user && token && !inviteConsumed) {
      setShowLogoutConfirm(true);
    }
  }, [user, token, inviteConsumed]);

  const passwordValid = password.length >= 8;
  const passwordsMatch = password === confirmPassword;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!token) {
      setError('Invalid invite link. Please contact the person who invited you for a new link.');
      return;
    }

    if (!passwordValid) {
      setError('Password must be at least 8 characters.');
      return;
    }

    if (!passwordsMatch) {
      setError('Passwords do not match.');
      return;
    }

    setLoading(true);
    try {
      await acceptInvite(token, password);
      setInviteConsumed(true);
      toast.success("Welcome! You're all set.");
      navigate(from, { replace: true });
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Failed to accept invite. Please try again.';
      if (message.toLowerCase().includes('expired') || message.toLowerCase().includes('invalid')) {
        setError(
          'This invite link has expired or is invalid. Please contact the person who invited you for a new link.',
        );
      } else {
        setError(message);
      }
    } finally {
      setLoading(false);
    }
  };

  if (!token) {
    return (
      <div className="min-h-screen bg-muted">
        <div className="container mx-auto px-4 py-10">
          <div className="mb-6">
            <Link to="/" className="text-primary hover:text-primary">
              ← Back to Home
            </Link>
          </div>
          <div className="max-w-md mx-auto mt-8 p-6 bg-card rounded-lg shadow-md">
            <Alert variant="error">
              Invalid invite link. Please contact the person who invited you for a new link.
            </Alert>
          </div>
        </div>
      </div>
    );
  }

  if (showLogoutConfirm && user) {
    return (
      <div className="min-h-screen bg-muted">
        <div className="container mx-auto px-4 py-10">
          <div className="mb-6">
            <Link to="/" className="text-primary hover:text-primary">
              ← Back to Home
            </Link>
          </div>
          <div className="max-w-md mx-auto mt-8 p-6 bg-card rounded-lg shadow-md">
            <h2 className="text-xl font-bold mb-4 text-center">You're Already Logged In</h2>
            <p className="text-muted-foreground mb-6 text-center">
              You're currently logged in as <strong>{user.email}</strong>. To accept this invite for a different
              account, you'll need to log out first.
            </p>
            <div className="flex gap-3">
              <Button variant="secondary" className="flex-1" onClick={() => navigate(from)}>
                Cancel
              </Button>
              <Button
                className="flex-1"
                onClick={() => {
                  logout();
                  setShowLogoutConfirm(false);
                }}
              >
                Log Out & Continue
              </Button>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-muted">
      <div className="container mx-auto px-4 py-10">
        <div className="mb-6">
          <Link to="/" className="text-primary hover:text-primary">
            ← Back to Home
          </Link>
        </div>

        <div className="max-w-md mx-auto mt-8 p-6 bg-card rounded-lg shadow-md">
          {preview ? (
            <>
              <h2 className="text-2xl font-bold mb-2 text-center">Welcome, {preview.firstName}!</h2>
              <div className="text-center mb-6">
                <p className="text-muted-foreground">
                  You've been invited to <span className="font-semibold">"{preview.calcuttaName}"</span>
                </p>
                <p className="text-muted-foreground text-sm">organized by {preview.commissionerName}</p>
                {preview.tournamentStartingAt && (
                  <p className="text-muted-foreground text-sm mt-1">
                    Tournament starts {formatDate(preview.tournamentStartingAt)}
                  </p>
                )}
                <p className="text-muted-foreground mt-3">Choose a password and you're in.</p>
              </div>
            </>
          ) : (
            <>
              <h2 className="text-2xl font-bold mb-2 text-center">Welcome!</h2>
              <p className="text-muted-foreground mb-6 text-center">Choose a password and you're in.</p>
            </>
          )}

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label htmlFor="password" className="block text-sm font-medium text-foreground mb-1">
                Password
              </label>
              <Input
                type="password"
                id="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                autoFocus
              />
              <p className="mt-1 text-xs text-muted-foreground">Must be at least 8 characters</p>
            </div>

            <div>
              <label htmlFor="confirmPassword" className="block text-sm font-medium text-foreground mb-1">
                Confirm Password
              </label>
              <Input
                type="password"
                id="confirmPassword"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                required
              />
              {confirmPassword && !passwordsMatch && (
                <p className="mt-1 text-xs text-red-600">Passwords do not match</p>
              )}
            </div>

            {error && <Alert variant="error">{error}</Alert>}

            <Button type="submit" className="w-full" disabled={loading || !passwordValid || !passwordsMatch}>
              {loading ? 'Getting you in...' : 'Set Password & Continue'}
            </Button>
          </form>
        </div>
      </div>
    </div>
  );
}

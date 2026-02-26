import React, { useState, useEffect } from 'react';
import { Link, useNavigate, useSearchParams } from 'react-router-dom';
import { useUser } from '../contexts/UserContext';
import { Input } from '../components/ui/Input';
import { Button } from '../components/ui/Button';
import { Alert } from '../components/ui/Alert';

export function ResetPasswordPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const token = searchParams.get('token') ?? '';
  const { user, resetPassword, logout } = useUser();

  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [showLogoutConfirm, setShowLogoutConfirm] = useState(false);
  const [resetConsumed, setResetConsumed] = useState(false);

  useEffect(() => {
    if (user && token && !resetConsumed) {
      setShowLogoutConfirm(true);
    }
  }, [user, token, resetConsumed]);

  const passwordValid = password.length >= 8;
  const passwordsMatch = password === confirmPassword;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!token) {
      setError('Invalid reset link. Please request a new one.');
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
      await resetPassword(token, password);
      setResetConsumed(true);
      navigate('/pools', { replace: true });
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Failed to reset password. Please try again.';
      if (message.toLowerCase().includes('expired') || message.toLowerCase().includes('invalid')) {
        setError('This reset link has expired or is invalid. Please request a new one.');
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
            <Link to="/login" className="text-primary hover:text-primary">
              &larr; Back to Login
            </Link>
          </div>
          <div className="max-w-md mx-auto mt-8 p-6 bg-card rounded-lg shadow-md">
            <Alert variant="error">
              Invalid reset link. Please request a new password reset from the login page.
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
            <Link to="/login" className="text-primary hover:text-primary">
              &larr; Back to Login
            </Link>
          </div>
          <div className="max-w-md mx-auto mt-8 p-6 bg-card rounded-lg shadow-md">
            <h2 className="text-xl font-bold mb-4 text-center">You're Already Logged In</h2>
            <p className="text-muted-foreground mb-6 text-center">
              You're currently logged in as <strong>{user.email}</strong>. To reset your password, you'll need to log
              out first.
            </p>
            <div className="flex gap-3">
              <Button variant="secondary" className="flex-1" onClick={() => navigate('/pools')}>
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
          <Link to="/login" className="text-primary hover:text-primary">
            &larr; Back to Login
          </Link>
        </div>

        <div className="max-w-md mx-auto mt-8 p-6 bg-card rounded-lg shadow-md">
          <h2 className="text-2xl font-bold mb-2 text-center">Reset Password</h2>
          <p className="text-muted-foreground mb-6 text-center">Choose a new password for your account.</p>

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label htmlFor="password" className="block text-sm font-medium text-foreground mb-1">
                New Password
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
                Confirm New Password
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
              {loading ? 'Resetting...' : 'Reset Password'}
            </Button>
          </form>
        </div>
      </div>
    </div>
  );
}

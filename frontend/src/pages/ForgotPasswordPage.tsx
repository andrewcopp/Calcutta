import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import { userService } from '../services/userService';
import { Input } from '../components/ui/Input';
import { Button } from '../components/ui/Button';
import { Alert } from '../components/ui/Alert';

export function ForgotPasswordPage() {
  const [email, setEmail] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [submitted, setSubmitted] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!email.trim()) {
      setError('Please enter your email address.');
      return;
    }

    setLoading(true);
    try {
      await userService.forgotPassword(email.trim());
      setSubmitted(true);
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Something went wrong. Please try again.';
      setError(message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-muted">
      <div className="container mx-auto px-4 py-10">
        <div className="mb-6">
          <Link to="/login" className="text-primary hover:text-primary">
            &larr; Back to Login
          </Link>
        </div>

        <div className="max-w-md mx-auto mt-8 p-6 bg-card rounded-lg shadow-md">
          <h2 className="text-2xl font-bold mb-2 text-center">Forgot Password</h2>

          {submitted ? (
            <div className="text-center">
              <p className="text-muted-foreground mb-4">
                If that email is registered, a reset link has been sent. Check your inbox.
              </p>
              <Link to="/login" className="text-primary hover:text-primary text-sm">
                Back to Login
              </Link>
            </div>
          ) : (
            <>
              <p className="text-muted-foreground mb-6 text-center">
                Enter your email and we'll send you a link to reset your password.
              </p>

              <form onSubmit={handleSubmit} className="space-y-4">
                <div>
                  <label htmlFor="email" className="block text-sm font-medium text-foreground mb-1">
                    Email
                  </label>
                  <Input
                    type="email"
                    id="email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    required
                    autoFocus
                  />
                </div>

                {error && <Alert variant="error">{error}</Alert>}

                <Button type="submit" className="w-full" disabled={loading}>
                  {loading ? 'Sending...' : 'Send Reset Link'}
                </Button>

                <div className="text-center">
                  <Link to="/login" className="text-sm text-primary hover:text-primary">
                    Back to Login
                  </Link>
                </div>
              </form>
            </>
          )}
        </div>
      </div>
    </div>
  );
}

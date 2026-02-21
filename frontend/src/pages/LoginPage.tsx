import React, { useEffect } from 'react';
import { Link, useNavigate, useLocation, useSearchParams } from 'react-router-dom';
import { AuthForm } from '../components/Auth/AuthForm';
import { useUser } from '../contexts/useUser';
import { Alert } from '../components/ui/Alert';

export function LoginPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const [searchParams] = useSearchParams();
  const { user } = useUser();
  const from = (location.state as { from?: string })?.from ?? '/calcuttas';
  const expired = searchParams.get('expired') === 'true';

  useEffect(() => {
    if (user) navigate(from);
  }, [navigate, user, from]);

  return (
    <div className="min-h-screen bg-gray-100">
      <div className="container mx-auto px-4 py-10">
        <div className="mb-6">
          <Link to="/" className="text-blue-600 hover:text-blue-800">
            ← Back to Home
          </Link>
        </div>

        <div className="max-w-2xl mx-auto">
          {expired && (
            <Alert variant="warning" className="mb-4">
              Your session has expired. Please log in again.
            </Alert>
          )}
          <h1 className="text-3xl font-bold text-gray-900 mb-2">Log in</h1>
          <p className="text-gray-600 mb-6">Sign in to view your pools and entries.</p>
          <AuthForm />
        </div>
      </div>
    </div>
  );
}

import { ReactNode } from 'react';
import { useLocation } from 'react-router-dom';
import { ErrorBoundary } from './ErrorBoundary';

function RouteErrorFallback() {
  return (
  <div className="flex items-center justify-center py-24">
    <div className="max-w-md w-full bg-white rounded-lg shadow-md p-8 text-center">
      <h2 className="text-xl font-bold text-gray-900 mb-2">Something went wrong</h2>
      <p className="text-gray-600 mb-6">
        This section encountered an error. The rest of the app should still work.
      </p>
      <button
        onClick={() => window.location.reload()}
        className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors"
      >
        Refresh Page
      </button>
    </div>
  </div>
  );
}

export function RouteErrorBoundary({ children }: { children: ReactNode }) {
  const location = useLocation();
  return (
    <ErrorBoundary key={location.pathname} fallback={<RouteErrorFallback />}>
      {children}
    </ErrorBoundary>
  );
}

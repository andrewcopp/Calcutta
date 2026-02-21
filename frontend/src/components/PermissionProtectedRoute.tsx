import { ReactNode } from 'react';
import { Navigate } from 'react-router-dom';
import { useUser } from '../contexts/UserContext';
import { ADMIN_PERMISSIONS } from '../constants/permissions';

interface Props {
  permission: string;
  children: ReactNode;
  redirectTo?: string;
}

export function PermissionProtectedRoute({
  permission,
  children,
  redirectTo = '/',
}: Props) {
  const { user, hasPermission, permissionsLoading } = useUser();

  if (!user) {
    return <Navigate to="/login" replace />;
  }

  if (permissionsLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600" />
      </div>
    );
  }

  // "admin" is a meta-permission: grant access if user has any admin permission
  if (permission === 'admin') {
    if (!ADMIN_PERMISSIONS.some((p) => hasPermission(p))) {
      return <Navigate to={redirectTo} replace />;
    }
    return <>{children}</>;
  }

  if (!hasPermission(permission)) {
    return <Navigate to={redirectTo} replace />;
  }

  return <>{children}</>;
}

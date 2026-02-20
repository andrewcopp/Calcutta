import React from 'react';
import { Navigate } from 'react-router-dom';
import { useUser } from '../contexts/useUser';
import { ADMIN_PERMISSIONS } from '../constants/permissions';

interface Props {
  permission: string;
  children: React.ReactNode;
  redirectTo?: string;
}

export const PermissionProtectedRoute: React.FC<Props> = ({
  permission,
  children,
  redirectTo = '/',
}) => {
  const { user, hasPermission, permissionsLoading } = useUser();

  if (!user) {
    return <Navigate to="/login" replace />;
  }

  if (permissionsLoading) {
    return null;
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
};

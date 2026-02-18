import React from 'react';
import { Navigate } from 'react-router-dom';
import { useUser } from '../contexts/useUser';

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

  if (!hasPermission(permission)) {
    return <Navigate to={redirectTo} replace />;
  }

  return <>{children}</>;
};

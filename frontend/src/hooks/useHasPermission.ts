import { useUser } from '../contexts/useUser';

export function useHasPermission(permission: string): boolean {
  const { hasPermission } = useUser();
  return hasPermission(permission);
}

export function usePermissions(): {
  permissions: string[];
  loading: boolean;
  hasPermission: (permission: string) => boolean;
} {
  const { permissions, permissionsLoading, hasPermission } = useUser();
  return { permissions, loading: permissionsLoading, hasPermission };
}

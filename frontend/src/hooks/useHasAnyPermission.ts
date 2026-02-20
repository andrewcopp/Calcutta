import { useUser } from '../contexts/useUser';

export function useHasAnyPermission(permissions: readonly string[]): boolean {
  const { hasPermission } = useUser();
  return permissions.some((p) => hasPermission(p));
}

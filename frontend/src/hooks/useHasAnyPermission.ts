import { useUser } from '../contexts/UserContext';

export function useHasAnyPermission(permissions: readonly string[]): boolean {
  const { hasPermission } = useUser();
  return permissions.some((p) => hasPermission(p));
}

import { useUser } from '../contexts/useUser';

export function useHasPermission(permission: string): boolean {
  const { hasPermission } = useUser();
  return hasPermission(permission);
}

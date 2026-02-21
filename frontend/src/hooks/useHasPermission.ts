import { useUser } from '../contexts/UserContext';

export function useHasPermission(permission: string): boolean {
  const { hasPermission } = useUser();
  return hasPermission(permission);
}

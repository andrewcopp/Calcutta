import React, { useCallback, useEffect, useState } from 'react';
import type { User } from '../types/user';
import { userService } from '../services/userService';
import { UserContext } from './userContextInternal';

export const UserProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [permissions, setPermissions] = useState<string[]>([]);
  const [permissionsLoading, setPermissionsLoading] = useState(false);

  const fetchPermissions = useCallback(async () => {
    setPermissionsLoading(true);
    try {
      const perms = await userService.fetchPermissions();
      setPermissions(perms);
    } finally {
      setPermissionsLoading(false);
    }
  }, []);

  useEffect(() => {
    const currentUser = userService.getCurrentUser();
    if (currentUser) {
      setUser(currentUser);
      const storedPerms = userService.getStoredPermissions();
      setPermissions(storedPerms);
      void fetchPermissions();
    }
  }, [fetchPermissions]);

  const login = async (email: string, password: string) => {
    const loggedInUser = await userService.login({ email, password });
    setUser(loggedInUser);
    await fetchPermissions();
  };

  const signup = async (email: string, firstName: string, lastName: string, password: string) => {
    const signedUpUser = await userService.signup({ email, firstName, lastName, password });
    setUser(signedUpUser);
    await fetchPermissions();
  };

  const logout = () => {
    userService.logout();
    setUser(null);
    setPermissions([]);
  };

  const hasPermission = useCallback(
    (permission: string) => permissions.includes(permission),
    [permissions]
  );

  return (
    <UserContext.Provider value={{ user, permissions, permissionsLoading, login, signup, logout, hasPermission }}>
      {children}
    </UserContext.Provider>
  );
};
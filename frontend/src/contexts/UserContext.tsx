import React, { createContext, useCallback, useContext, useEffect, useState } from 'react';
import type { User } from '../schemas/user';
import { userService } from '../services/userService';

export interface UserContextType {
  user: User | null;
  permissions: string[];
  permissionsLoading: boolean;
  login: (email: string, password: string) => Promise<void>;
  acceptInvite: (token: string, password: string) => Promise<void>;
  resetPassword: (token: string, password: string) => Promise<void>;
  logout: () => void;
  hasPermission: (permission: string) => boolean;
}

const UserContext = createContext<UserContextType | undefined>(undefined);

// eslint-disable-next-line react-refresh/only-export-components
export const useUser = () => {
  const context = useContext(UserContext);
  if (context === undefined) {
    throw new Error('useUser must be used within a UserProvider');
  }
  return context;
};

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

  const acceptInvite = async (token: string, password: string) => {
    const acceptedUser = await userService.acceptInvite(token, password);
    setUser(acceptedUser);
    await fetchPermissions();
  };

  const resetPassword = async (token: string, password: string) => {
    const resetUser = await userService.resetPassword(token, password);
    setUser(resetUser);
    await fetchPermissions();
  };

  const logout = () => {
    userService.logout();
    setUser(null);
    setPermissions([]);
  };

  const hasPermission = useCallback((permission: string) => permissions.includes(permission), [permissions]);

  return (
    <UserContext.Provider
      value={{ user, permissions, permissionsLoading, login, acceptInvite, resetPassword, logout, hasPermission }}
    >
      {children}
    </UserContext.Provider>
  );
};

import React, { useEffect, useState } from 'react';
import type { User } from '../types/user';
import { userService } from '../services/userService';
import { UserContext } from './userContextInternal';

export const UserProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);

  useEffect(() => {
    const currentUser = userService.getCurrentUser();
    if (currentUser) {
      setUser(currentUser);
    }
  }, []);

  const login = async (email: string, password: string) => {
    const user = await userService.login({ email, password });
    setUser(user);
  };

  const signup = async (email: string, firstName: string, lastName: string, password: string) => {
    const user = await userService.signup({ email, firstName, lastName, password });
    setUser(user);
  };

  const logout = () => {
    userService.logout();
    setUser(null);
  };

  return (
    <UserContext.Provider value={{ user, login, signup, logout }}>
      {children}
    </UserContext.Provider>
  );
};
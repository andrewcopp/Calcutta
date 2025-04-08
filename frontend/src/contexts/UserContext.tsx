import React, { createContext, useContext, useState, useEffect } from 'react';
import { User } from '../types/user';
import { userService } from '../services/userService';

interface UserContextType {
  user: User | null;
  login: (email: string) => Promise<void>;
  signup: (email: string, firstName: string, lastName: string) => Promise<void>;
  logout: () => void;
}

const UserContext = createContext<UserContextType | undefined>(undefined);

export const UserProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);

  useEffect(() => {
    const currentUser = userService.getCurrentUser();
    if (currentUser) {
      setUser(currentUser);
    }
  }, []);

  const login = async (email: string) => {
    const user = await userService.login({ email });
    setUser(user);
  };

  const signup = async (email: string, firstName: string, lastName: string) => {
    const user = await userService.signup({ email, firstName, lastName });
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

export const useUser = () => {
  const context = useContext(UserContext);
  if (context === undefined) {
    throw new Error('useUser must be used within a UserProvider');
  }
  return context;
}; 
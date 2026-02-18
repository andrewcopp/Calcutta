import { createContext } from 'react';

import type { User } from '../types/user';

export interface UserContextType {
  user: User | null;
  permissions: string[];
  permissionsLoading: boolean;
  login: (email: string, password: string) => Promise<void>;
  signup: (email: string, firstName: string, lastName: string, password: string) => Promise<void>;
  logout: () => void;
  hasPermission: (permission: string) => boolean;
}

export const UserContext = createContext<UserContextType | undefined>(undefined);

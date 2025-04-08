import React from 'react';
import { useUser } from '../../contexts/UserContext';

export const UserMenu: React.FC = () => {
  const { user, logout } = useUser();

  if (!user) {
    return null;
  }

  return (
    <div className="relative">
      <div className="flex items-center space-x-4">
        <span className="text-sm text-gray-700">
          {user.firstName} {user.lastName}
        </span>
        <button
          onClick={logout}
          className="text-sm text-red-600 hover:text-red-800"
        >
          Logout
        </button>
      </div>
    </div>
  );
}; 
import React from 'react';
import { Link, useLocation } from 'react-router-dom';

export const Header: React.FC = () => {
  const location = useLocation();
  const isAdminPage = location.pathname.startsWith('/admin');

  return (
    <header className="bg-white shadow-md">
      <div className="container mx-auto px-4 py-4">
        <div className="flex justify-between items-center">
          <div className="flex items-center space-x-6">
            <Link to="/" className="text-xl font-bold text-gray-800">
              Calcutta
            </Link>
            {isAdminPage && (
              <Link to="/admin" className="text-gray-600 hover:text-gray-800">
                Admin Console
              </Link>
            )}
          </div>
          
          {!isAdminPage && (
            <Link
              to="/admin"
              className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors"
            >
              Admin Console
            </Link>
          )}
        </div>
      </div>
    </header>
  );
};

export default Header; 
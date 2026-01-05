import React from 'react';
import { Link } from 'react-router-dom';
import { UserMenu } from './Header/UserMenu';

export const Header: React.FC = () => {
  return (
    <header className="bg-white shadow-md">
      <div className="container mx-auto px-4 py-4">
        <div className="flex justify-between items-center">
          <div className="flex items-center space-x-6">
            <Link to="/" className="text-xl font-bold text-gray-800">
              Calcutta
            </Link>
            <nav className="hidden md:flex space-x-4">
              <Link to="/calcuttas" className="text-gray-600 hover:text-gray-800">
                Calcuttas
              </Link>
              <Link to="/lab" className="text-gray-600 hover:text-gray-800">
                Lab
              </Link>
              <Link to="/sandbox" className="text-gray-600 hover:text-gray-800">
                Sandbox
              </Link>
              <Link to="/rules" className="text-gray-600 hover:text-gray-800">
                How It Works
              </Link>
            </nav>
          </div>
          
          <div className="flex items-center space-x-4">
            <UserMenu />
          </div>
        </div>
      </div>
    </header>
  );
};

export default Header;
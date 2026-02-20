import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import { useHasPermission } from '../hooks/useHasPermission';
import { useHasAnyPermission } from '../hooks/useHasAnyPermission';
import { PERMISSIONS, ADMIN_PERMISSIONS } from '../constants/permissions';
import { UserMenu } from './Header/UserMenu';

export const Header: React.FC = () => {
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const canAccessLab = useHasPermission(PERMISSIONS.LAB_READ);
  const canAccessAdmin = useHasAnyPermission(ADMIN_PERMISSIONS);

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
              {canAccessLab && (
                <Link to="/lab" className="text-gray-600 hover:text-gray-800">
                  Lab
                </Link>
              )}
              <Link to="/rules" className="text-gray-600 hover:text-gray-800">
                How It Works
              </Link>
              {canAccessAdmin && (
                <Link to="/admin" className="text-gray-600 hover:text-gray-800">
                  Admin
                </Link>
              )}
            </nav>
          </div>

          <div className="flex items-center space-x-4">
            <div className="hidden md:block">
              <UserMenu />
            </div>
            {/* Mobile hamburger */}
            <button
              className="md:hidden p-2 rounded-md text-gray-600 hover:text-gray-800 hover:bg-gray-100"
              onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
              aria-label="Toggle menu"
              aria-expanded={mobileMenuOpen}
            >
              <svg className="h-6 w-6" fill="none" viewBox="0 0 24 24" strokeWidth="1.5" stroke="currentColor">
                {mobileMenuOpen ? (
                  <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
                ) : (
                  <path strokeLinecap="round" strokeLinejoin="round" d="M3.75 6.75h16.5M3.75 12h16.5m-16.5 5.25h16.5" />
                )}
              </svg>
            </button>
          </div>
        </div>

        {/* Mobile menu */}
        {mobileMenuOpen && (
          <nav className="md:hidden mt-4 pb-2 border-t border-gray-200 pt-4 space-y-2">
            <Link
              to="/calcuttas"
              className="block px-3 py-2 rounded-md text-gray-600 hover:text-gray-800 hover:bg-gray-100"
              onClick={() => setMobileMenuOpen(false)}
            >
              Calcuttas
            </Link>
            {canAccessLab && (
              <Link
                to="/lab"
                className="block px-3 py-2 rounded-md text-gray-600 hover:text-gray-800 hover:bg-gray-100"
                onClick={() => setMobileMenuOpen(false)}
              >
                Lab
              </Link>
            )}
            <Link
              to="/rules"
              className="block px-3 py-2 rounded-md text-gray-600 hover:text-gray-800 hover:bg-gray-100"
              onClick={() => setMobileMenuOpen(false)}
            >
              How It Works
            </Link>
            {canAccessAdmin && (
              <Link
                to="/admin"
                className="block px-3 py-2 rounded-md text-gray-600 hover:text-gray-800 hover:bg-gray-100"
                onClick={() => setMobileMenuOpen(false)}
              >
                Admin
              </Link>
            )}
            <div className="pt-2 border-t border-gray-200">
              <UserMenu />
            </div>
          </nav>
        )}
      </div>
    </header>
  );
};

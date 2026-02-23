import { useState } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { useHasPermission } from '../hooks/useHasPermission';
import { useHasAnyPermission } from '../hooks/useHasAnyPermission';
import { PERMISSIONS, ADMIN_PERMISSIONS } from '../constants/permissions';
import { UserMenu } from './Header/UserMenu';
import { cn } from '../lib/cn';

export function Header() {
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const canAccessLab = useHasPermission(PERMISSIONS.LAB_READ);
  const canAccessAdmin = useHasAnyPermission(ADMIN_PERMISSIONS);
  const location = useLocation();

  const navLinks = [
    { to: '/calcuttas', label: 'My Pools', show: true },
    { to: '/lab', label: 'Lab', show: canAccessLab },
    { to: '/rules', label: 'How It Works', show: true },
    { to: '/admin', label: 'Admin', show: canAccessAdmin },
  ];

  const isActive = (path: string) => location.pathname.startsWith(path);

  return (
    <header className="bg-header shadow-lg shadow-black/10">
      <a
        href="#main-content"
        className="sr-only focus:not-sr-only focus:absolute focus:z-50 focus:bg-card focus:px-4 focus:py-2 focus:text-sm focus:font-medium focus:text-primary"
      >
        Skip to content
      </a>
      <div className="container mx-auto px-4 py-4">
        <div className="flex justify-between items-center">
          <div className="flex items-center space-x-6">
            <Link to="/" className="text-xl font-bold text-header-text">
              <span className="text-blue-400">C</span>alcutta
            </Link>
            <nav className="hidden md:flex space-x-4">
              {navLinks
                .filter((l) => l.show)
                .map((link) => (
                  <Link
                    key={link.to}
                    to={link.to}
                    className={cn(
                      'transition-colors',
                      isActive(link.to) ? 'text-header-text' : 'text-header-muted hover:text-header-text',
                    )}
                  >
                    {link.label}
                  </Link>
                ))}
            </nav>
          </div>

          <div className="flex items-center space-x-4">
            <div className="hidden md:block">
              <UserMenu />
            </div>
            {/* Mobile hamburger */}
            <button
              className="md:hidden p-2 rounded-md text-header-muted hover:text-header-text hover:bg-white/10"
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
          <nav className="md:hidden mt-4 pb-2 border-t border-white/10 pt-4 space-y-2">
            {navLinks
              .filter((l) => l.show)
              .map((link) => (
                <Link
                  key={link.to}
                  to={link.to}
                  className={cn(
                    'block px-3 py-2 rounded-md transition-colors',
                    isActive(link.to)
                      ? 'text-header-text bg-white/10'
                      : 'text-header-muted hover:text-header-text hover:bg-white/10',
                  )}
                  onClick={() => setMobileMenuOpen(false)}
                >
                  {link.label}
                </Link>
              ))}
            <div className="pt-2 border-t border-white/10">
              <UserMenu />
            </div>
          </nav>
        )}
      </div>
    </header>
  );
}

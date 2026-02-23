import { Link } from 'react-router-dom';
import { useUser } from '../../contexts/UserContext';
import { useHasAnyPermission } from '../../hooks/useHasAnyPermission';
import { ADMIN_PERMISSIONS } from '../../constants/permissions';
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuLabel,
} from '../ui/DropdownMenu';

export function UserMenu() {
  const { user, logout } = useUser();
  const canAccessAdmin = useHasAnyPermission(ADMIN_PERMISSIONS);

  if (!user) {
    return null;
  }

  const initials = `${user.firstName.charAt(0)}${user.lastName.charAt(0)}`.toUpperCase();

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <button
          className="flex items-center justify-center w-8 h-8 rounded-full bg-blue-500 text-white text-sm font-semibold hover:bg-blue-400 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-400"
          aria-label="User menu"
        >
          {initials}
        </button>
      </DropdownMenuTrigger>

      <DropdownMenuContent align="end">
        <DropdownMenuLabel>
          {user.firstName} {user.lastName}
        </DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DropdownMenuItem asChild>
          <Link to="/profile">Profile</Link>
        </DropdownMenuItem>
        {canAccessAdmin && (
          <DropdownMenuItem asChild>
            <Link to="/admin">Admin Console</Link>
          </DropdownMenuItem>
        )}
        <DropdownMenuItem className="text-destructive focus:text-destructive" onSelect={() => logout()}>
          Logout
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

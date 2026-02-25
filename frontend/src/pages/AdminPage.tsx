import { Link } from 'react-router-dom';
import { Card } from '../components/ui/Card';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { useHasPermission } from '../hooks/useHasPermission';
import { PERMISSIONS } from '../constants/permissions';

export function AdminPage() {
  const canTournaments = useHasPermission(PERMISSIONS.TOURNAMENT_GAME_WRITE);
  const canTournamentImports = useHasPermission(PERMISSIONS.ADMIN_BUNDLES_EXPORT);
  const canApiKeys = useHasPermission(PERMISSIONS.ADMIN_API_KEYS_WRITE);
  const canUsers = useHasPermission(PERMISSIONS.ADMIN_USERS_READ);
  const canUserMerge = useHasPermission(PERMISSIONS.ADMIN_USERS_WRITE);
  const canHof = useHasPermission(PERMISSIONS.ADMIN_HOF_READ);

  return (
    <PageContainer>
      <PageHeader title="Admin Console" />

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {canTournaments && (
          <Link to="/admin/tournaments" className="block">
            <Card className="hover:shadow-md transition-shadow">
              <h2 className="text-xl font-semibold mb-2">Tournaments</h2>
              <p className="text-muted-foreground">Manage tournaments, teams, and brackets</p>
            </Card>
          </Link>
        )}

        {canTournaments && (
          <Link to="/admin/kenpom" className="block">
            <Card className="hover:shadow-md transition-shadow">
              <h2 className="text-xl font-semibold mb-2">KenPom Ratings</h2>
              <p className="text-muted-foreground">Enter or edit KenPom ratings for tournament teams</p>
            </Card>
          </Link>
        )}

        {canTournaments && (
          <Link to="/admin/predictions" className="block">
            <Card className="hover:shadow-md transition-shadow">
              <h2 className="text-xl font-semibold mb-2">Predictions</h2>
              <p className="text-muted-foreground">View pre-tournament advancement probabilities by round</p>
            </Card>
          </Link>
        )}

        {canTournamentImports && (
          <Link to="/admin/tournament-imports" className="block">
            <Card className="hover:shadow-md transition-shadow">
              <h2 className="text-xl font-semibold mb-2">Tournament Data</h2>
              <p className="text-muted-foreground">Export or import tournament data archives</p>
            </Card>
          </Link>
        )}

        {canApiKeys && (
          <Link to="/admin/api-keys" className="block">
            <Card className="hover:shadow-md transition-shadow">
              <h2 className="text-xl font-semibold mb-2">API Keys</h2>
              <p className="text-muted-foreground">Create, view, and revoke server-to-server API keys</p>
            </Card>
          </Link>
        )}

        {canUsers && (
          <Link to="/admin/users" className="block">
            <Card className="hover:shadow-md transition-shadow">
              <h2 className="text-xl font-semibold mb-2">Users</h2>
              <p className="text-muted-foreground">View users and their roles/permissions</p>
            </Card>
          </Link>
        )}

        {canUserMerge && (
          <Link to="/admin/user-merges" className="block">
            <Card className="hover:shadow-md transition-shadow">
              <h2 className="text-xl font-semibold mb-2">User Merge</h2>
              <p className="text-muted-foreground">Consolidate stub users from historical imports with real accounts</p>
            </Card>
          </Link>
        )}

        {canHof && (
          <Link to="/admin/hall-of-fame" className="block">
            <Card className="hover:shadow-md transition-shadow">
              <h2 className="text-xl font-semibold mb-2">Hall of Fame</h2>
              <p className="text-muted-foreground">
                Leaderboards for best teams, investments, and entries across all years
              </p>
            </Card>
          </Link>
        )}
      </div>
    </PageContainer>
  );
}

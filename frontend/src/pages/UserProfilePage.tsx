import { useQuery } from '@tanstack/react-query';
import { userService } from '../services/userService';
import { queryKeys } from '../queryKeys';
import { ErrorState } from '../components/ui/ErrorState';
import { Breadcrumb } from '../components/ui/Breadcrumb';
import { Card } from '../components/ui/Card';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { Badge } from '../components/ui/Badge';
import { formatDate } from '../utils/format';

function getStatusBadgeVariant(status: string): 'success' | 'warning' | 'secondary' | 'default' {
  switch (status) {
    case 'active':
      return 'success';
    case 'invited':
    case 'requires_password_setup':
      return 'warning';
    case 'stub':
      return 'secondary';
    default:
      return 'default';
  }
}

export function UserProfilePage() {
  const profileQuery = useQuery({
    queryKey: queryKeys.profile.me(),
    queryFn: () => userService.fetchProfile(),
  });

  const profile = profileQuery.data;

  return (
    <PageContainer>
      <Breadcrumb items={[{ label: 'Profile' }]} />
      <PageHeader title="My Profile" />

      {profileQuery.isLoading && <LoadingState label="Loading profile..." layout="inline" />}
      {profileQuery.isError && <ErrorState error={profileQuery.error} onRetry={() => profileQuery.refetch()} />}

      {profile && (
        <Card>
          <div className="space-y-4">
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div>
                <div className="text-sm text-muted-foreground">Name</div>
                <div className="font-medium">
                  {profile.firstName} {profile.lastName}
                </div>
              </div>
              <div>
                <div className="text-sm text-muted-foreground">Email</div>
                <div className="font-medium">
                  {profile.email ?? <span className="text-muted-foreground/60 italic">Not set</span>}
                </div>
              </div>
              <div>
                <div className="text-sm text-muted-foreground">Status</div>
                <Badge variant={getStatusBadgeVariant(profile.status)}>{profile.status}</Badge>
              </div>
              <div>
                <div className="text-sm text-muted-foreground">Member Since</div>
                <div className="font-medium">{formatDate(profile.createdAt)}</div>
              </div>
            </div>

            <div>
              <div className="text-sm text-muted-foreground mb-1">Roles</div>
              {profile.roles.length > 0 ? (
                <div className="flex flex-wrap gap-1">
                  {profile.roles.map((l) => (
                    <Badge key={l} variant="default">
                      {l}
                    </Badge>
                  ))}
                </div>
              ) : (
                <span className="text-muted-foreground/60">None</span>
              )}
            </div>

            <div>
              <div className="text-sm text-muted-foreground mb-1">Permissions</div>
              {profile.permissions.length > 0 ? (
                <div className="flex flex-wrap gap-1">
                  {profile.permissions.map((p) => (
                    <Badge key={p} variant="secondary">
                      {p}
                    </Badge>
                  ))}
                </div>
              ) : (
                <span className="text-muted-foreground/60">None</span>
              )}
            </div>
          </div>
        </Card>
      )}
    </PageContainer>
  );
}

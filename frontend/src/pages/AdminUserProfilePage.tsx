import { useState } from 'react';
import { useParams } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { adminService } from '../services/adminService';
import { queryKeys } from '../queryKeys';
import { useHasPermission } from '../hooks/useHasPermission';
import { PERMISSIONS } from '../constants/permissions';
import { ErrorState } from '../components/ui/ErrorState';
import { Breadcrumb } from '../components/ui/Breadcrumb';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { Badge } from '../components/ui/Badge';
import { Select } from '../components/ui/Select';
import { formatDate } from '../utils/format';

const ALL_LABELS = ['site_admin', 'tournament_admin', 'calcutta_admin', 'player', 'user_manager'];

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

export function AdminUserProfilePage() {
  const { userId } = useParams<{ userId: string }>();
  const queryClient = useQueryClient();
  const canWrite = useHasPermission(PERMISSIONS.ADMIN_USERS_WRITE);
  const [selectedLabel, setSelectedLabel] = useState('');

  const userQuery = useQuery({
    queryKey: queryKeys.admin.userDetail(userId),
    queryFn: () => (userId ? adminService.getUser(userId) : Promise.reject('No userId')),
    enabled: Boolean(userId),
  });

  const grantMutation = useMutation({
    mutationFn: (labelKey: string) => adminService.grantLabel(userId!, labelKey),
    onSuccess: () => {
      setSelectedLabel('');
      void queryClient.invalidateQueries({ queryKey: queryKeys.admin.userDetail(userId) });
    },
  });

  const revokeMutation = useMutation({
    mutationFn: (labelKey: string) => adminService.revokeLabel(userId!, labelKey),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: queryKeys.admin.userDetail(userId) });
    },
  });

  const profile = userQuery.data;
  const availableLabels = profile ? ALL_LABELS.filter((l) => !profile.labels.includes(l)) : [];

  return (
    <PageContainer>
      <Breadcrumb
        items={[
          { label: 'Admin', href: '/admin' },
          { label: 'Users', href: '/admin/users' },
          { label: profile ? `${profile.firstName} ${profile.lastName}` : 'User' },
        ]}
      />
      <PageHeader title={profile ? `${profile.firstName} ${profile.lastName}` : 'User Detail'} />

      {userQuery.isLoading && <LoadingState label="Loading user..." layout="inline" />}
      {userQuery.isError && <ErrorState error={userQuery.error} onRetry={() => userQuery.refetch()} />}

      {profile && (
        <div className="space-y-6">
          <Card>
            <h2 className="text-lg font-semibold mb-4">User Info</h2>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div>
                <div className="text-sm text-gray-500">Name</div>
                <div className="font-medium">{profile.firstName} {profile.lastName}</div>
              </div>
              <div>
                <div className="text-sm text-gray-500">Email</div>
                <div className="font-medium">{profile.email ?? <span className="text-gray-400 italic">Not set</span>}</div>
              </div>
              <div>
                <div className="text-sm text-gray-500">Status</div>
                <Badge variant={getStatusBadgeVariant(profile.status)}>{profile.status}</Badge>
              </div>
              <div>
                <div className="text-sm text-gray-500">Member Since</div>
                <div className="font-medium">{formatDate(profile.createdAt)}</div>
              </div>
            </div>
          </Card>

          <Card>
            <h2 className="text-lg font-semibold mb-4">Labels</h2>
            {profile.labels.length > 0 ? (
              <div className="flex flex-wrap gap-2 mb-4">
                {profile.labels.map((l) => (
                  <span key={l} className="inline-flex items-center gap-1">
                    <Badge variant="default">{l}</Badge>
                    {canWrite && (
                      <button
                        className="text-gray-400 hover:text-red-600 text-sm"
                        onClick={() => revokeMutation.mutate(l)}
                        disabled={revokeMutation.isPending}
                        aria-label={`Remove ${l}`}
                      >
                        x
                      </button>
                    )}
                  </span>
                ))}
              </div>
            ) : (
              <p className="text-gray-400 mb-4">No labels assigned.</p>
            )}

            {canWrite && availableLabels.length > 0 && (
              <div className="flex items-center gap-2">
                <Select
                  value={selectedLabel}
                  onChange={(e) => setSelectedLabel(e.target.value)}
                  className="w-48"
                >
                  <option value="">Select label...</option>
                  {availableLabels.map((l) => (
                    <option key={l} value={l}>{l}</option>
                  ))}
                </Select>
                <Button
                  size="sm"
                  disabled={!selectedLabel || grantMutation.isPending}
                  onClick={() => {
                    if (selectedLabel) grantMutation.mutate(selectedLabel);
                  }}
                >
                  Grant
                </Button>
              </div>
            )}
          </Card>

          <Card>
            <h2 className="text-lg font-semibold mb-4">Permissions</h2>
            {profile.permissions.length > 0 ? (
              <div className="flex flex-wrap gap-1">
                {profile.permissions.map((p) => (
                  <Badge key={p} variant="secondary">{p}</Badge>
                ))}
              </div>
            ) : (
              <span className="text-gray-400">None</span>
            )}
          </Card>
        </div>
      )}
    </PageContainer>
  );
}

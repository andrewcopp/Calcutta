import { useState } from 'react';
import { useParams } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { adminService } from '../services/adminService';
import { calcuttaService } from '../services/calcuttaService';
import { tournamentService } from '../services/tournamentService';
import { queryKeys } from '../queryKeys';
import { toast } from '../lib/toast';
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
import type { LabelGrant } from '../types/admin';

const ALL_LABELS = ['site_admin', 'tournament_admin', 'calcutta_admin', 'player', 'user_manager'];

const LABEL_SCOPES: Record<string, string[]> = {
  site_admin: ['global'],
  user_manager: ['global'],
  calcutta_admin: ['global', 'calcutta'],
  tournament_admin: ['global', 'tournament'],
  player: ['global', 'calcutta'],
};

function labelGrantKey(g: LabelGrant): string {
  return g.scopeId ? `${g.key}:${g.scopeType}:${g.scopeId}` : `${g.key}:global`;
}

function labelGrantDisplay(g: LabelGrant): string {
  if (g.scopeType === 'global') return g.key;
  return `${g.key} (${g.scopeName ?? g.scopeId})`;
}

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
  const [selectedScopeType, setSelectedScopeType] = useState<'global' | 'calcutta' | 'tournament'>('global');
  const [selectedScopeId, setSelectedScopeId] = useState('');

  const userQuery = useQuery({
    queryKey: queryKeys.admin.userDetail(userId),
    queryFn: () => (userId ? adminService.getUser(userId) : Promise.reject('No userId')),
    enabled: Boolean(userId),
  });

  const calcuttasQuery = useQuery({
    queryKey: queryKeys.calcuttas.all(),
    queryFn: () => calcuttaService.getAllCalcuttas(),
    enabled: canWrite,
  });

  const tournamentsQuery = useQuery({
    queryKey: queryKeys.tournaments.all(),
    queryFn: () => tournamentService.getAllTournaments(),
    enabled: canWrite,
  });

  const grantMutation = useMutation({
    mutationFn: ({ labelKey, scopeType, scopeId }: { labelKey: string; scopeType: string; scopeId?: string }) =>
      adminService.grantLabel(userId!, labelKey, scopeType, scopeId),
    onSuccess: () => {
      setSelectedLabel('');
      setSelectedScopeType('global');
      setSelectedScopeId('');
      toast.success('Label granted');
      void queryClient.invalidateQueries({ queryKey: queryKeys.admin.userDetail(userId) });
    },
  });

  const revokeMutation = useMutation({
    mutationFn: (grant: LabelGrant) =>
      adminService.revokeLabel(userId!, grant.key, grant.scopeType, grant.scopeId),
    onSuccess: () => {
      toast.success('Label revoked');
      void queryClient.invalidateQueries({ queryKey: queryKeys.admin.userDetail(userId) });
    },
  });

  const profile = userQuery.data;

  // Only exclude a label from the dropdown if it already has a global grant
  const globalLabelKeys = new Set(
    profile?.labels.filter((g) => g.scopeType === 'global').map((g) => g.key) ?? [],
  );
  const availableLabels = ALL_LABELS.filter((l) => !globalLabelKeys.has(l));

  const allowedScopes = selectedLabel ? (LABEL_SCOPES[selectedLabel] ?? ['global']) : ['global'];
  const needsScopeSelection = allowedScopes.length > 1;
  const needsScopeId = selectedScopeType !== 'global';

  const canGrant =
    selectedLabel &&
    (!needsScopeId || selectedScopeId) &&
    !grantMutation.isPending;

  function handleLabelChange(label: string) {
    setSelectedLabel(label);
    const scopes = LABEL_SCOPES[label] ?? ['global'];
    setSelectedScopeType(scopes.length === 1 ? (scopes[0] as 'global' | 'calcutta' | 'tournament') : 'global');
    setSelectedScopeId('');
  }

  function handleGrant() {
    if (!selectedLabel) return;
    grantMutation.mutate({
      labelKey: selectedLabel,
      scopeType: selectedScopeType,
      scopeId: needsScopeId ? selectedScopeId : undefined,
    });
  }

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
                {profile.labels.map((g) => (
                  <span key={labelGrantKey(g)} className="inline-flex items-center gap-1">
                    <Badge variant="default">{labelGrantDisplay(g)}</Badge>
                    {canWrite && (
                      <button
                        className="text-gray-400 hover:text-red-600 text-sm"
                        onClick={() => revokeMutation.mutate(g)}
                        disabled={revokeMutation.isPending}
                        aria-label={`Remove ${labelGrantDisplay(g)}`}
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
              <div className="flex flex-wrap items-center gap-2">
                <Select
                  value={selectedLabel}
                  onChange={(e) => handleLabelChange(e.target.value)}
                  className="w-48"
                >
                  <option value="">Select label...</option>
                  {availableLabels.map((l) => (
                    <option key={l} value={l}>{l}</option>
                  ))}
                </Select>

                {selectedLabel && needsScopeSelection && (
                  <Select
                    value={selectedScopeType}
                    onChange={(e) => {
                      setSelectedScopeType(e.target.value as 'global' | 'calcutta' | 'tournament');
                      setSelectedScopeId('');
                    }}
                    className="w-40"
                  >
                    {allowedScopes.map((s) => (
                      <option key={s} value={s}>{s === 'global' ? 'Global' : `Specific ${s}`}</option>
                    ))}
                  </Select>
                )}

                {selectedLabel && selectedScopeType === 'calcutta' && (
                  <Select
                    value={selectedScopeId}
                    onChange={(e) => setSelectedScopeId(e.target.value)}
                    className="w-48"
                  >
                    <option value="">Select calcutta...</option>
                    {(calcuttasQuery.data ?? []).map((c) => (
                      <option key={c.id} value={c.id}>{c.name}</option>
                    ))}
                  </Select>
                )}

                {selectedLabel && selectedScopeType === 'tournament' && (
                  <Select
                    value={selectedScopeId}
                    onChange={(e) => setSelectedScopeId(e.target.value)}
                    className="w-48"
                  >
                    <option value="">Select tournament...</option>
                    {(tournamentsQuery.data ?? []).map((t) => (
                      <option key={t.id} value={t.id}>{t.name}</option>
                    ))}
                  </Select>
                )}

                <Button
                  size="sm"
                  disabled={!canGrant}
                  onClick={handleGrant}
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

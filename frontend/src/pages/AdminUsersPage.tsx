import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { adminService, AdminUserListItem } from '../services/adminService';
import { queryKeys } from '../queryKeys';
import { ErrorState } from '../components/ui/ErrorState';
import { Breadcrumb } from '../components/ui/Breadcrumb';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { Table, TableBody, TableCell, TableHead, TableHeaderCell, TableRow } from '../components/ui/Table';
import { Select } from '../components/ui/Select';
import { Badge } from '../components/ui/Badge';
import { Alert } from '../components/ui/Alert';
import { SetEmailModal } from '../components/Admin/SetEmailModal';
import { InviteConfirmModal } from '../components/Admin/InviteConfirmModal';
import { formatDate } from '../utils/format';

type StatusFilter = '' | 'stub' | 'invited' | 'requires_password_setup' | 'active';

const STATUS_OPTIONS: { value: StatusFilter; label: string }[] = [
  { value: '', label: 'All Statuses' },
  { value: 'stub', label: 'Stub' },
  { value: 'invited', label: 'Invited' },
  { value: 'requires_password_setup', label: 'Requires Password Setup' },
  { value: 'active', label: 'Active' },
];

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

function formatStatusLabel(status: string): string {
  switch (status) {
    case 'requires_password_setup':
      return 'Pending Setup';
    case 'stub':
      return 'Stub';
    case 'invited':
      return 'Invited';
    case 'active':
      return 'Active';
    default:
      return status;
  }
}

export const AdminUsersPage: React.FC = () => {
  const queryClient = useQueryClient();
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('');
  const [successMessage, setSuccessMessage] = useState('');

  const [emailModalUser, setEmailModalUser] = useState<AdminUserListItem | null>(null);
  const [inviteModalUser, setInviteModalUser] = useState<AdminUserListItem | null>(null);

  const usersQuery = useQuery({
    queryKey: queryKeys.admin.users(statusFilter || undefined),
    queryFn: () => adminService.listUsers(statusFilter || undefined),
  });

  const users = usersQuery.data?.items ?? [];

  const handleSetEmail = async (userId: string, email: string) => {
    await adminService.setUserEmail(userId, email);
    setSuccessMessage('Email set successfully.');
    await queryClient.invalidateQueries({ queryKey: queryKeys.admin.users() });
    setTimeout(() => setSuccessMessage(''), 3000);
  };

  const handleSendInvite = async (userId: string) => {
    await adminService.sendInvite(userId);
    setSuccessMessage('Invite sent successfully.');
    await queryClient.invalidateQueries({ queryKey: queryKeys.admin.users() });
    setTimeout(() => setSuccessMessage(''), 3000);
  };

  return (
    <PageContainer>
      <Breadcrumb
        items={[
          { label: 'Admin', href: '/admin' },
          { label: 'Users' },
        ]}
      />
      <PageHeader
        title="Admin: Users"
        subtitle="Manage users, set emails, and send invites."
        actions={
          <Link to="/admin" className="text-blue-600 hover:text-blue-800">
            Back to Admin Console
          </Link>
        }
      />

      <Card>
        <div className="flex items-center justify-between mb-4 gap-4">
          <div className="flex items-center gap-4">
            <h2 className="text-xl font-semibold">Users</h2>
            <Select
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value as StatusFilter)}
              className="w-48"
            >
              {STATUS_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </Select>
          </div>
          <Button onClick={() => usersQuery.refetch()} disabled={usersQuery.isFetching} variant="secondary">
            Refresh
          </Button>
        </div>

        {successMessage && (
          <Alert variant="success" className="mb-4">{successMessage}</Alert>
        )}

        {usersQuery.isError && (
          <ErrorState error={usersQuery.error} onRetry={() => usersQuery.refetch()} />
        )}

        {usersQuery.isLoading ? <LoadingState label="Loading users..." layout="inline" /> : null}

        <Table className="text-sm">
          <TableHead>
            <TableRow>
              <TableHeaderCell>Email</TableHeaderCell>
              <TableHeaderCell>Name</TableHeaderCell>
              <TableHeaderCell>Status</TableHeaderCell>
              <TableHeaderCell>Created</TableHeaderCell>
              <TableHeaderCell>Labels</TableHeaderCell>
              <TableHeaderCell>Actions</TableHeaderCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {users.map((u) => (
              <TableRow key={u.id} className="align-top">
                <TableCell className="whitespace-nowrap">
                  {u.email ?? <span className="text-gray-400 italic">No email</span>}
                </TableCell>
                <TableCell className="whitespace-nowrap">{u.first_name} {u.last_name}</TableCell>
                <TableCell>
                  <Badge variant={getStatusBadgeVariant(u.status)}>
                    {formatStatusLabel(u.status)}
                  </Badge>
                  {u.last_invite_sent_at && (
                    <div className="text-xs text-gray-500 mt-1">
                      Invited: {formatDate(u.last_invite_sent_at)}
                    </div>
                  )}
                </TableCell>
                <TableCell className="whitespace-nowrap">{formatDate(u.created_at)}</TableCell>
                <TableCell>
                  {(u.labels ?? []).length > 0 ? (
                    <div className="flex flex-wrap gap-1">
                      {(u.labels ?? []).map((l) => (
                        <Badge key={l} variant="default">{l}</Badge>
                      ))}
                    </div>
                  ) : (
                    <span className="text-gray-500">-</span>
                  )}
                </TableCell>
                <TableCell>
                  <div className="flex gap-2">
                    {u.status === 'stub' && (
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => setEmailModalUser(u)}
                      >
                        Set Email
                      </Button>
                    )}
                    {(u.status === 'invited' || u.status === 'requires_password_setup') && u.email && (
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => setInviteModalUser(u)}
                      >
                        {u.last_invite_sent_at ? 'Resend Invite' : 'Send Invite'}
                      </Button>
                    )}
                    {u.status === 'active' && (
                      <span className="text-gray-400 text-sm">-</span>
                    )}
                  </div>
                </TableCell>
              </TableRow>
            ))}

            {users.length === 0 && !usersQuery.isLoading ? (
              <TableRow>
                <TableCell className="text-gray-500" colSpan={6}>
                  No users found.
                </TableCell>
              </TableRow>
            ) : null}
          </TableBody>
        </Table>
      </Card>

      {emailModalUser && (
        <SetEmailModal
          open={!!emailModalUser}
          onClose={() => setEmailModalUser(null)}
          userId={emailModalUser.id}
          userName={`${emailModalUser.first_name} ${emailModalUser.last_name}`}
          onSubmit={handleSetEmail}
        />
      )}

      {inviteModalUser && (
        <InviteConfirmModal
          open={!!inviteModalUser}
          onClose={() => setInviteModalUser(null)}
          userId={inviteModalUser.id}
          userName={`${inviteModalUser.first_name} ${inviteModalUser.last_name}`}
          userEmail={inviteModalUser.email ?? ''}
          lastInviteSentAt={inviteModalUser.last_invite_sent_at}
          onConfirm={handleSendInvite}
        />
      )}
    </PageContainer>
  );
};

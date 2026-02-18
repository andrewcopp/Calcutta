import React from 'react';
import { Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { adminService } from '../services/adminService';
import { queryKeys } from '../queryKeys';
import { Alert } from '../components/ui/Alert';
import { Breadcrumb } from '../components/ui/Breadcrumb';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { Table, TableBody, TableCell, TableHead, TableHeaderCell, TableRow } from '../components/ui/Table';
import { formatDate } from '../utils/format';

export const AdminUsersPage: React.FC = () => {
  const usersQuery = useQuery({
    queryKey: queryKeys.admin.users(),
    queryFn: () => adminService.listUsers(),
  });

  const users = usersQuery.data?.items ?? [];

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
        subtitle="List of all users and their effective global labels/permissions."
        actions={
          <Link to="/admin" className="text-blue-600 hover:text-blue-800">
            Back to Admin Console
          </Link>
        }
      />

      <Card>
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-xl font-semibold">Users</h2>
          <Button onClick={() => usersQuery.refetch()} disabled={usersQuery.isFetching} variant="secondary">
            Refresh
          </Button>
        </div>

        {usersQuery.isError && (
          <Alert variant="error" className="mb-4">
            {usersQuery.error instanceof Error ? usersQuery.error.message : String(usersQuery.error)}
          </Alert>
        )}

        {usersQuery.isLoading ? <LoadingState label="Loading users..." layout="inline" /> : null}

        <Table className="text-sm">
          <TableHead>
            <TableRow>
              <TableHeaderCell>Email</TableHeaderCell>
              <TableHeaderCell>Name</TableHeaderCell>
              <TableHeaderCell>Created</TableHeaderCell>
              <TableHeaderCell>Labels</TableHeaderCell>
              <TableHeaderCell>Permissions</TableHeaderCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {users.map((u) => (
              <TableRow key={u.id} className="align-top">
                <TableCell className="whitespace-nowrap">{u.email}</TableCell>
                <TableCell className="whitespace-nowrap">{u.first_name} {u.last_name}</TableCell>
                <TableCell className="whitespace-nowrap">{formatDate(u.created_at)}</TableCell>
                <TableCell>
                  {(u.labels ?? []).length > 0 ? (
                    <div className="flex flex-wrap gap-2">
                      {(u.labels ?? []).map((l) => (
                        <span key={l} className="px-2 py-1 bg-blue-50 border border-blue-200 rounded">
                          {l}
                        </span>
                      ))}
                    </div>
                  ) : (
                    <span className="text-gray-500">-</span>
                  )}
                </TableCell>
                <TableCell>
                  {(u.permissions ?? []).length > 0 ? (
                    <div className="flex flex-wrap gap-2">
                      {(u.permissions ?? []).map((p) => (
                        <span key={p} className="px-2 py-1 bg-gray-50 border border-gray-200 rounded">
                          {p}
                        </span>
                      ))}
                    </div>
                  ) : (
                    <span className="text-gray-500">-</span>
                  )}
                </TableCell>
              </TableRow>
            ))}

            {users.length === 0 && !usersQuery.isLoading ? (
              <TableRow>
                <TableCell className="text-gray-500" colSpan={5}>
                  No users found.
                </TableCell>
              </TableRow>
            ) : null}
          </TableBody>
        </Table>
      </Card>
    </PageContainer>
  );
};

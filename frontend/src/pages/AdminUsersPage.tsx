import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { apiClient } from '../api/apiClient';
import { Alert } from '../components/ui/Alert';
import { Breadcrumb } from '../components/ui/Breadcrumb';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { Table, TableBody, TableCell, TableHead, TableHeaderCell, TableRow } from '../components/ui/Table';

type AdminUserListItem = {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  created_at: string;
  updated_at: string;
  labels: string[];
  permissions: string[];
};

type AdminUsersListResponse = {
  items: AdminUserListItem[];
};

export const AdminUsersPage: React.FC = () => {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [users, setUsers] = useState<AdminUserListItem[]>([]);

  const load = async () => {
    setError(null);
    setBusy(true);
    try {
      const res = await apiClient.get<AdminUsersListResponse>('/admin/users');
      setUsers(res.items ?? []);
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
    } finally {
      setBusy(false);
    }
  };

  useEffect(() => {
    void load();
  }, []);

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
          <Button onClick={load} disabled={busy} variant="secondary">
            Refresh
          </Button>
        </div>

        {error && (
          <Alert variant="error" className="mb-4">
            {error}
          </Alert>
        )}

        {busy && users.length === 0 ? <LoadingState label="Loading users..." layout="inline" /> : null}

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
                <TableCell className="whitespace-nowrap">{u.created_at}</TableCell>
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

            {users.length === 0 && !busy ? (
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

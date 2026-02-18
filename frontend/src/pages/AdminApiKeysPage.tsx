import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { adminService, CreateAPIKeyResponse } from '../services/adminService';
import { queryKeys } from '../queryKeys';
import { Alert } from '../components/ui/Alert';
import { Breadcrumb } from '../components/ui/Breadcrumb';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { Input } from '../components/ui/Input';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { Table, TableBody, TableCell, TableHead, TableHeaderCell, TableRow } from '../components/ui/Table';

export const AdminApiKeysPage: React.FC = () => {
  const queryClient = useQueryClient();

  const [label, setLabel] = useState('python-ds');
  const [created, setCreated] = useState<CreateAPIKeyResponse | null>(null);

  const keysQuery = useQuery({
    queryKey: queryKeys.admin.apiKeys(),
    queryFn: () => adminService.listApiKeys(),
  });

  const createMutation = useMutation({
    mutationFn: (trimmedLabel: string) => adminService.createApiKey(trimmedLabel),
    onSuccess: (data) => {
      setCreated(data);
      queryClient.invalidateQueries({ queryKey: queryKeys.admin.apiKeys() });
    },
  });

  const revokeMutation = useMutation({
    mutationFn: (id: string) => adminService.revokeApiKey(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.admin.apiKeys() });
    },
  });

  const busy = createMutation.isPending || revokeMutation.isPending;
  const error = keysQuery.error || createMutation.error || revokeMutation.error;
  const keys = keysQuery.data?.items ?? [];

  const copy = async (value: string) => {
    try {
      await navigator.clipboard.writeText(value);
    } catch (e) {
      // Clipboard errors are non-critical; silently ignored
    }
  };

  const handleRevoke = (id: string) => {
    if (window.confirm('Are you sure you want to revoke this API key? This cannot be undone.')) {
      revokeMutation.mutate(id);
    }
  };

  return (
    <PageContainer>
      <Breadcrumb
        items={[
          { label: 'Admin', href: '/admin' },
          { label: 'API Keys' },
        ]}
      />
      <PageHeader
        title="Admin: API Keys"
        subtitle="Create API keys for server-to-server access (e.g. the Python data science client)."
        actions={
          <Link to="/admin" className="text-blue-600 hover:text-blue-800">
            Back to Admin Console
          </Link>
        }
      />

      <Card className="mb-6">
        <h2 className="text-xl font-semibold mb-2">Mint a new key</h2>
        <div className="flex flex-col gap-3">
          <div className="flex flex-col sm:flex-row gap-3">
            <Input
              type="text"
              value={label}
              onChange={(e) => setLabel(e.target.value)}
              disabled={busy}
              placeholder="Label (optional)"
              className="flex-1"
            />
            <Button onClick={() => createMutation.mutate(label.trim())} disabled={busy}>
              Create API key
            </Button>
          </div>
        </div>

        {error && (
          <Alert variant="error" className="mt-4">
            {error instanceof Error ? error.message : String(error)}
          </Alert>
        )}

        {created && (
          <div className="mt-6">
            <h3 className="text-lg font-semibold mb-2">New key (copy this now)</h3>
            <div className="bg-yellow-50 border border-yellow-200 rounded p-4">
              <div className="mb-2 text-sm text-gray-700">This value will only be shown once.</div>
              <pre className="bg-white rounded p-3 overflow-auto text-sm">{created.key}</pre>
              <div className="mt-3">
                <Button onClick={() => copy(created.key)} variant="secondary">
                  Copy key
                </Button>
              </div>
            </div>
          </div>
        )}
      </Card>

      <Card>
        <h2 className="text-xl font-semibold mb-2">Your keys</h2>
        <p className="text-gray-600 mb-4">These are key records (not the raw secret).</p>

        {keysQuery.isLoading ? (
          <LoadingState label="Loading API keys..." layout="inline" />
        ) : (
          <Table className="text-sm">
            <TableHead>
              <TableRow>
                <TableHeaderCell>Label</TableHeaderCell>
                <TableHeaderCell>Created</TableHeaderCell>
                <TableHeaderCell>Last used</TableHeaderCell>
                <TableHeaderCell>Revoked</TableHeaderCell>
                <TableHeaderCell>Actions</TableHeaderCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {keys.map((k) => (
                <TableRow key={k.id}>
                  <TableCell>{k.label || '-'}</TableCell>
                  <TableCell>{k.created_at}</TableCell>
                  <TableCell>{k.last_used_at || '-'}</TableCell>
                  <TableCell>{k.revoked_at || '-'}</TableCell>
                  <TableCell>
                    <Button
                      onClick={() => handleRevoke(k.id)}
                      disabled={busy || Boolean(k.revoked_at)}
                      variant="secondary"
                      size="sm"
                      className="bg-red-600 text-white hover:bg-red-700 focus-visible:ring-red-600"
                    >
                      Revoke
                    </Button>
                  </TableCell>
                </TableRow>
              ))}

              {keys.length === 0 ? (
                <TableRow>
                  <TableCell className="text-gray-500" colSpan={5}>
                    No keys yet.
                  </TableCell>
                </TableRow>
              ) : null}
            </TableBody>
          </Table>
        )}

        <div className="mt-6">
          <Button onClick={() => keysQuery.refetch()} disabled={busy || keysQuery.isFetching} variant="secondary">
            Refresh list
          </Button>
        </div>

        <div className="mt-6 text-sm text-gray-600">
          <div>Use in Python as: <code>Authorization: Bearer &lt;api_key&gt;</code></div>
        </div>
      </Card>
    </PageContainer>
  );
};

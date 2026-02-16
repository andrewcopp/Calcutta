import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { apiClient } from '../api/apiClient';
import { Alert } from '../components/ui/Alert';
import { Breadcrumb } from '../components/ui/Breadcrumb';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { Input } from '../components/ui/Input';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { Table, TableBody, TableCell, TableHead, TableHeaderCell, TableRow } from '../components/ui/Table';

type CreateAPIKeyRequest = {
  label?: string;
};

type CreateAPIKeyResponse = {
  id: string;
  key: string;
  label?: string;
  created_at: string;
};

type APIKeyListItem = {
  id: string;
  label?: string;
  created_at: string;
  revoked_at?: string;
  last_used_at?: string;
};

type ListAPIKeysResponse = {
  items: APIKeyListItem[];
};

export const AdminApiKeysPage: React.FC = () => {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [label, setLabel] = useState('python-ds');
  const [created, setCreated] = useState<CreateAPIKeyResponse | null>(null);

  const [keys, setKeys] = useState<APIKeyListItem[]>([]);

  const load = async () => {
    setError(null);
    try {
      const res = await apiClient.get<ListAPIKeysResponse>('/admin/api-keys');
      setKeys(res.items ?? []);
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
    }
  };

  useEffect(() => {
    void load();
  }, []);

  const create = async () => {
    setError(null);
    setCreated(null);
    setBusy(true);
    try {
      const body: CreateAPIKeyRequest = {};
      const trimmed = label.trim();
      if (trimmed) body.label = trimmed;

      const res = await apiClient.post<CreateAPIKeyResponse>('/admin/api-keys', body);
      setCreated(res);
      await load();
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
    } finally {
      setBusy(false);
    }
  };

  const revoke = async (id: string) => {
    setError(null);
    setBusy(true);
    try {
      await apiClient.delete<void>(`/admin/api-keys/${id}`);
      await load();
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
    } finally {
      setBusy(false);
    }
  };

  const copy = async (value: string) => {
    setError(null);
    try {
      await navigator.clipboard.writeText(value);
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
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
            <Button onClick={create} disabled={busy}>
              Create API key
            </Button>
          </div>
        </div>

        {busy ? <LoadingState label="Working..." layout="inline" /> : null}

        {error && (
          <Alert variant="error" className="mt-4">
            {error}
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
                    onClick={() => revoke(k.id)}
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

        <div className="mt-6">
          <Button onClick={load} disabled={busy} variant="secondary">
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

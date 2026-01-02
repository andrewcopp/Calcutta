import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { apiClient } from '../api/apiClient';

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
    <div className="container mx-auto px-4 py-8">
      <div className="mb-6">
        <Link to="/admin" className="text-blue-600 hover:text-blue-800 mb-4 inline-block">
          Back to Admin Console
        </Link>
        <h1 className="text-3xl font-bold">Admin: API Keys</h1>
        <p className="text-gray-600 mt-2">Create API keys for server-to-server access (e.g. the Python data science client).</p>
      </div>

      <div className="bg-white rounded-lg shadow-md p-6 mb-6">
        <h2 className="text-xl font-semibold mb-2">Mint a new key</h2>
        <div className="flex flex-col gap-3">
          <div className="flex flex-col sm:flex-row gap-3">
            <input
              type="text"
              value={label}
              onChange={(e) => setLabel(e.target.value)}
              disabled={busy}
              placeholder="Label (optional)"
              className="border rounded px-3 py-2 flex-1"
            />
            <button
              onClick={create}
              disabled={busy}
              className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
            >
              Create API key
            </button>
          </div>
        </div>

        {created && (
          <div className="mt-6">
            <h3 className="text-lg font-semibold mb-2">New key (copy this now)</h3>
            <div className="bg-yellow-50 border border-yellow-200 rounded p-4">
              <div className="mb-2 text-sm text-gray-700">This value will only be shown once.</div>
              <pre className="bg-white rounded p-3 overflow-auto text-sm">{created.key}</pre>
              <div className="mt-3">
                <button
                  onClick={() => copy(created.key)}
                  className="px-4 py-2 bg-gray-800 text-white rounded hover:bg-gray-900"
                >
                  Copy key
                </button>
              </div>
            </div>
          </div>
        )}
      </div>

      <div className="bg-white rounded-lg shadow-md p-6">
        <h2 className="text-xl font-semibold mb-2">Your keys</h2>
        <p className="text-gray-600 mb-4">These are key records (not the raw secret).</p>

        <div className="overflow-auto">
          <table className="min-w-full text-sm">
            <thead>
              <tr className="text-left border-b">
                <th className="py-2 pr-4">Label</th>
                <th className="py-2 pr-4">Created</th>
                <th className="py-2 pr-4">Last used</th>
                <th className="py-2 pr-4">Revoked</th>
                <th className="py-2 pr-4">Actions</th>
              </tr>
            </thead>
            <tbody>
              {keys.map((k) => (
                <tr key={k.id} className="border-b">
                  <td className="py-2 pr-4">{k.label || '-'}</td>
                  <td className="py-2 pr-4">{k.created_at}</td>
                  <td className="py-2 pr-4">{k.last_used_at || '-'}</td>
                  <td className="py-2 pr-4">{k.revoked_at || '-'}</td>
                  <td className="py-2 pr-4">
                    <button
                      onClick={() => revoke(k.id)}
                      disabled={busy || Boolean(k.revoked_at)}
                      className="px-3 py-1 bg-red-600 text-white rounded hover:bg-red-700 disabled:opacity-50"
                    >
                      Revoke
                    </button>
                  </td>
                </tr>
              ))}
              {keys.length === 0 && (
                <tr>
                  <td className="py-3 text-gray-500" colSpan={5}>
                    No keys yet.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>

        {error && <div className="mt-4 text-red-600">{error}</div>}

        <div className="mt-6">
          <button
            onClick={load}
            disabled={busy}
            className="px-4 py-2 bg-gray-200 text-gray-800 rounded hover:bg-gray-300 disabled:opacity-50"
          >
            Refresh list
          </button>
        </div>

        <div className="mt-6 text-sm text-gray-600">
          <div>Use in Python as: <code>Authorization: Bearer &lt;api_key&gt;</code></div>
        </div>
      </div>
    </div>
  );
};

export default AdminApiKeysPage;

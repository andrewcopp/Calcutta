import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import { apiClient } from '../api/apiClient';

export const AdminPage: React.FC = () => {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [label, setLabel] = useState('python-ds');
  const [createdKey, setCreatedKey] = useState<string | null>(null);

  const mint = async () => {
    setError(null);
    setCreatedKey(null);
    setBusy(true);
    try {
      const trimmed = label.trim();
      const body: { label?: string } = {};
      if (trimmed) body.label = trimmed;

      const res = await apiClient.post<{ key: string }>('/admin/api-keys', body);
      setCreatedKey(res.key);
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
      <h1 className="text-3xl font-bold mb-8">Admin Console</h1>

      <div className="bg-white rounded-lg shadow-md p-6 mb-6">
        <h2 className="text-xl font-semibold mb-2">Mint API key</h2>
        <p className="text-gray-600 mb-4">
          Generate a key for server-to-server access (Python env var: <code>CALCUTTA_API_KEY</code>). The raw key is only shown once.
        </p>

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
            onClick={mint}
            disabled={busy}
            className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
          >
            Generate
          </button>
        </div>

        {createdKey && (
          <div className="mt-4">
            <pre className="bg-gray-100 rounded p-3 overflow-auto text-sm">{createdKey}</pre>
            <div className="mt-3">
              <button
                onClick={() => copy(createdKey)}
                className="px-4 py-2 bg-gray-800 text-white rounded hover:bg-gray-900"
              >
                Copy
              </button>
            </div>
          </div>
        )}

        {error && <div className="mt-4 text-red-600">{error}</div>}
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        <Link 
          to="/admin/tournaments" 
          className="p-6 bg-white rounded-lg shadow-md hover:shadow-lg transition-shadow"
        >
          <h2 className="text-xl font-semibold mb-2">Tournaments</h2>
          <p className="text-gray-600">Manage tournaments, teams, and brackets</p>
        </Link>

        <Link 
          to="/admin/bundles" 
          className="p-6 bg-white rounded-lg shadow-md hover:shadow-lg transition-shadow"
        >
          <h2 className="text-xl font-semibold mb-2">Bundles</h2>
          <p className="text-gray-600">Export or import bundle archives</p>
        </Link>

        <Link 
          to="/admin/analytics" 
          className="p-6 bg-white rounded-lg shadow-md hover:shadow-lg transition-shadow"
        >
          <h2 className="text-xl font-semibold mb-2">Analytics</h2>
          <p className="text-gray-600">View historical trends and patterns across all calcuttas</p>
        </Link>

        <Link 
          to="/admin/hall-of-fame" 
          className="p-6 bg-white rounded-lg shadow-md hover:shadow-lg transition-shadow"
        >
          <h2 className="text-xl font-semibold mb-2">Hall of Fame</h2>
          <p className="text-gray-600">Leaderboards for best teams, investments, and entries across all years</p>
        </Link>
      </div>
    </div>
  );
};

export default AdminPage; 
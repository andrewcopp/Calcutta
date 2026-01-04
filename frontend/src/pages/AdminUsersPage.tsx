import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { apiClient } from '../api/apiClient';

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
    <div className="container mx-auto px-4 py-8">
      <div className="mb-6">
        <Link to="/admin" className="text-blue-600 hover:text-blue-800 mb-4 inline-block">
          Back to Admin Console
        </Link>
        <h1 className="text-3xl font-bold">Admin: Users</h1>
        <p className="text-gray-600 mt-2">List of all users and their effective global labels/permissions.</p>
      </div>

      <div className="bg-white rounded-lg shadow-md p-6">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-xl font-semibold">Users</h2>
          <button
            onClick={load}
            disabled={busy}
            className="px-4 py-2 bg-gray-200 text-gray-800 rounded hover:bg-gray-300 disabled:opacity-50"
          >
            Refresh
          </button>
        </div>

        {error && <div className="mb-4 text-red-600">{error}</div>}

        <div className="overflow-auto">
          <table className="min-w-full text-sm">
            <thead>
              <tr className="text-left border-b">
                <th className="py-2 pr-4">Email</th>
                <th className="py-2 pr-4">Name</th>
                <th className="py-2 pr-4">Created</th>
                <th className="py-2 pr-4">Labels</th>
                <th className="py-2 pr-4">Permissions</th>
              </tr>
            </thead>
            <tbody>
              {users.map((u) => (
                <tr key={u.id} className="border-b align-top">
                  <td className="py-2 pr-4 whitespace-nowrap">{u.email}</td>
                  <td className="py-2 pr-4 whitespace-nowrap">{u.first_name} {u.last_name}</td>
                  <td className="py-2 pr-4 whitespace-nowrap">{u.created_at}</td>
                  <td className="py-2 pr-4">
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
                  </td>
                  <td className="py-2 pr-4">
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
                  </td>
                </tr>
              ))}
              {users.length === 0 && !busy && (
                <tr>
                  <td className="py-3 text-gray-500" colSpan={5}>
                    No users found.
                  </td>
                </tr>
              )}
              {busy && users.length === 0 && (
                <tr>
                  <td className="py-3 text-gray-500" colSpan={5}>
                    Loading...
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};

export default AdminUsersPage;

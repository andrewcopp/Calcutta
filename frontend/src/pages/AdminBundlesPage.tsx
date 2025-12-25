import React, { useMemo, useState } from 'react';

type ImportResponse = {
  upload_id: string;
  filename: string;
  sha256: string;
  size_bytes: number;
  import_report: {
    started_at: string;
    finished_at: string;
    dry_run: boolean;
    schools: number;
    tournaments: number;
    tournament_teams: number;
    calcuttas: number;
    entries: number;
    bids: number;
    payouts: number;
    rounds: number;
  };
  verify_report: {
    ok: boolean;
    mismatch_count: number;
    mismatches?: { where: string; what: string }[];
  };
};

export const AdminBundlesPage: React.FC = () => {
  const API_URL = useMemo(() => import.meta.env.VITE_API_URL || import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080', []);

  const [file, setFile] = useState<File | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<ImportResponse | null>(null);

  const download = async () => {
    setError(null);
    setResult(null);
    setBusy(true);
    try {
      const res = await fetch(`${API_URL}/api/admin/bundles/export`, { credentials: 'include' });
      if (!res.ok) {
        const txt = await res.text().catch(() => '');
        throw new Error(txt || `Export failed (${res.status})`);
      }
      const blob = await res.blob();
      const cd = res.headers.get('content-disposition') || '';
      const match = /filename="([^"]+)"/i.exec(cd);
      const filename = match?.[1] || 'bundles.zip';

      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = filename;
      document.body.appendChild(a);
      a.click();
      a.remove();
      window.URL.revokeObjectURL(url);
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
    } finally {
      setBusy(false);
    }
  };

  const upload = async () => {
    if (!file) {
      setError('Please choose a zip file');
      return;
    }
    setError(null);
    setResult(null);
    setBusy(true);
    try {
      const form = new FormData();
      form.append('file', file);

      const res = await fetch(`${API_URL}/api/admin/bundles/import`, { method: 'POST', body: form, credentials: 'include' });
      const body = await res.json().catch(() => undefined);
      if (!res.ok) {
        const msg = body?.error?.message || `Import failed (${res.status})`;
        throw new Error(msg);
      }
      setResult(body as ImportResponse);
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="container mx-auto px-4 py-8">
      <h1 className="text-3xl font-bold mb-6">Admin: Bundles</h1>

      <div className="bg-white rounded-lg shadow-md p-6 mb-6">
        <h2 className="text-xl font-semibold mb-2">Export</h2>
        <p className="text-gray-600 mb-4">Download a zip of the current DB state in bundle format.</p>
        <button
          onClick={download}
          disabled={busy}
          className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
        >
          Download bundles (.zip)
        </button>
      </div>

      <div className="bg-white rounded-lg shadow-md p-6">
        <h2 className="text-xl font-semibold mb-2">Import</h2>
        <p className="text-gray-600 mb-4">Upload a bundles zip and import it into the database.</p>

        <div className="flex flex-col gap-3">
          <input
            type="file"
            accept="application/zip,.zip"
            disabled={busy}
            onChange={(e) => setFile(e.target.files?.[0] ?? null)}
          />
          <div>
            <button
              onClick={upload}
              disabled={busy || !file}
              className="px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700 disabled:opacity-50"
            >
              Upload + Import
            </button>
          </div>
        </div>

        {error && <div className="mt-4 text-red-600">{error}</div>}

        {result && (
          <div className="mt-6">
            <h3 className="text-lg font-semibold mb-2">Result</h3>
            <pre className="bg-gray-100 rounded p-4 overflow-auto text-sm">{JSON.stringify(result, null, 2)}</pre>
          </div>
        )}
      </div>
    </div>
  );
};

export default AdminBundlesPage;

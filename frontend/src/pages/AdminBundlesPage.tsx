import React, { useEffect, useState } from 'react';
import { API_URL, apiClient } from '../api/apiClient';
import { Alert } from '../components/ui/Alert';
import { Breadcrumb } from '../components/ui/Breadcrumb';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';

type ImportStartResponse = {
  upload_id: string;
  status: 'pending' | 'running' | 'succeeded' | 'failed';
  filename: string;
  sha256: string;
  size_bytes: number;
};

type ImportStatusResponse = {
  upload_id: string;
  filename: string;
  sha256: string;
  size_bytes: number;
  status: 'pending' | 'running' | 'succeeded' | 'failed';
  started_at?: string;
  finished_at?: string;
  error_message?: string;
  import_report?: {
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
  verify_report?: {
    ok: boolean;
    mismatch_count: number;
    mismatches?: { where: string; what: string }[];
  };
};

export const AdminBundlesPage: React.FC = () => {
  const [file, setFile] = useState<File | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<ImportStatusResponse | null>(null);
  const [uploadId, setUploadId] = useState<string | null>(null);

  useEffect(() => {
    if (!uploadId) return;
    if (!busy) return;

    let cancelled = false;
    let timeoutId: number | undefined;

    const poll = async () => {
      try {
        const res = await apiClient.fetch(`${API_URL}/api/admin/bundles/import/${uploadId}`, { credentials: 'include' });
        const body = (await res.json().catch(() => undefined)) as ImportStatusResponse | undefined;
        if (!res.ok) {
          const maybeError = body && typeof body === 'object' ? (body as Record<string, unknown>).error : undefined;
          const rawMsg =
            (maybeError && typeof maybeError === 'object' ? (maybeError as Record<string, unknown>).message : undefined) ||
            `Status check failed (${res.status})`;
          throw new Error(typeof rawMsg === 'string' ? rawMsg : String(rawMsg));
        }
        if (cancelled) return;

        setResult(body ?? null);

        if (!body) {
          timeoutId = window.setTimeout(poll, 1000);
          return;
        }
        if (body.status === 'succeeded') {
          setBusy(false);
          return;
        }
        if (body.status === 'failed') {
          setBusy(false);
          setError(body.error_message || 'Import failed');
          return;
        }

        timeoutId = window.setTimeout(poll, 1000);
      } catch (e) {
        if (cancelled) return;
        setError(e instanceof Error ? e.message : String(e));
        setBusy(false);
      }
    };

    void poll();
    return () => {
      cancelled = true;
      if (timeoutId !== undefined) window.clearTimeout(timeoutId);
    };
  }, [busy, uploadId]);

  const download = async () => {
    setError(null);
    setResult(null);
    setBusy(true);
    try {
      const res = await apiClient.fetch(`${API_URL}/api/admin/bundles/export`, { credentials: 'include' });
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
    setUploadId(null);
    setBusy(true);
    try {
      const form = new FormData();
      form.append('file', file);

      const res = await apiClient.fetch(`${API_URL}/api/admin/bundles/import`, { method: 'POST', body: form, credentials: 'include' });
      const body = (await res.json().catch(() => undefined)) as unknown;
      if (!res.ok) {
        const record = body && typeof body === 'object' ? (body as Record<string, unknown>) : undefined;
        const maybeError = record?.error;
        const rawMsg =
          (maybeError && typeof maybeError === 'object' ? (maybeError as Record<string, unknown>).message : undefined) ||
          `Import failed (${res.status})`;
        throw new Error(typeof rawMsg === 'string' ? rawMsg : String(rawMsg));
      }
      const started = body as ImportStartResponse;
      setUploadId(started.upload_id);
      setResult({
        upload_id: started.upload_id,
        filename: started.filename,
        sha256: started.sha256,
        size_bytes: started.size_bytes,
        status: started.status,
      });
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
      setBusy(false);
    }
  };

  return (
    <PageContainer>
      <Breadcrumb
        items={[
          { label: 'Admin', href: '/admin' },
          { label: 'Bundles' },
        ]}
      />
      <PageHeader title="Admin: Bundles" />

      <Card className="mb-6">
        <h2 className="text-xl font-semibold mb-2">Export</h2>
        <p className="text-gray-600 mb-4">Download a zip of the current DB state in bundle format.</p>
        <Button onClick={download} disabled={busy}>
          Download bundles (.zip)
        </Button>
      </Card>

      <Card>
        <h2 className="text-xl font-semibold mb-2">Import</h2>
        <p className="text-gray-600 mb-4">Upload a bundles zip and import it into the database.</p>

        <div className="flex flex-col gap-3">
          <label className="flex items-center gap-3 cursor-pointer">
            <span className="inline-flex items-center px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 transition-colors">
              Choose File
            </span>
            <input
              type="file"
              accept="application/zip,.zip"
              disabled={busy}
              onChange={(e) => setFile(e.target.files?.[0] ?? null)}
              className="sr-only"
            />
            <span className="text-sm text-gray-500">{file ? file.name : 'No file selected'}</span>
          </label>
          <div>
            <Button onClick={upload} disabled={busy || !file}>
              Upload + Import
            </Button>
          </div>
        </div>

        {busy ? <LoadingState label="Working..." layout="inline" /> : null}

        {error && (
          <Alert variant="error" className="mt-4">
            {error}
          </Alert>
        )}

        {result && (
          <div className="mt-6">
            <h3 className="text-lg font-semibold mb-3">Result</h3>
            <div className="grid grid-cols-2 gap-x-6 gap-y-2 text-sm">
              <div className="text-gray-500">Status</div>
              <div className="font-medium">{result.status}</div>
              <div className="text-gray-500">Filename</div>
              <div className="font-medium">{result.filename}</div>
              <div className="text-gray-500">Size</div>
              <div className="font-medium">{(result.size_bytes / 1024).toFixed(1)} KB</div>
              {result.import_report && (
                <>
                  <div className="col-span-2 border-t border-gray-200 mt-2 pt-2 font-semibold">Import Report</div>
                  <div className="text-gray-500">Schools</div>
                  <div className="font-medium">{result.import_report.schools}</div>
                  <div className="text-gray-500">Tournaments</div>
                  <div className="font-medium">{result.import_report.tournaments}</div>
                  <div className="text-gray-500">Teams</div>
                  <div className="font-medium">{result.import_report.tournament_teams}</div>
                  <div className="text-gray-500">Calcuttas</div>
                  <div className="font-medium">{result.import_report.calcuttas}</div>
                  <div className="text-gray-500">Entries</div>
                  <div className="font-medium">{result.import_report.entries}</div>
                  <div className="text-gray-500">Bids</div>
                  <div className="font-medium">{result.import_report.bids}</div>
                  <div className="text-gray-500">Payouts</div>
                  <div className="font-medium">{result.import_report.payouts}</div>
                </>
              )}
              {result.verify_report && (
                <>
                  <div className="col-span-2 border-t border-gray-200 mt-2 pt-2 font-semibold">Verify Report</div>
                  <div className="text-gray-500">OK</div>
                  <div className={`font-medium ${result.verify_report.ok ? 'text-green-600' : 'text-red-600'}`}>
                    {result.verify_report.ok ? 'Yes' : 'No'}
                  </div>
                  {result.verify_report.mismatch_count > 0 && (
                    <>
                      <div className="text-gray-500">Mismatches</div>
                      <div className="font-medium text-red-600">{result.verify_report.mismatch_count}</div>
                    </>
                  )}
                </>
              )}
            </div>
          </div>
        )}
      </Card>
    </PageContainer>
  );
};

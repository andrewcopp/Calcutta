import { useEffect, useRef, useState } from 'react';
import { Alert } from '../components/ui/Alert';
import { Breadcrumb } from '../components/ui/Breadcrumb';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { adminService } from '../services/adminService';
import type { TournamentImportStatusResponse } from '../types/admin';

const MAX_POLL_ATTEMPTS = 120;

export function AdminTournamentImportsPage() {
  const [file, setFile] = useState<File | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<TournamentImportStatusResponse | null>(null);
  const [uploadId, setUploadId] = useState<string | null>(null);
  const pollAttemptsRef = useRef(0);

  useEffect(() => {
    if (!uploadId) return;
    if (!busy) return;

    let cancelled = false;
    let timeoutId: number | undefined;
    pollAttemptsRef.current = 0;

    const poll = async () => {
      pollAttemptsRef.current += 1;
      if (pollAttemptsRef.current > MAX_POLL_ATTEMPTS) {
        setError('Import timed out after too many poll attempts');
        setBusy(false);
        return;
      }
      try {
        const body = await adminService.getTournamentImportStatus(uploadId);
        if (cancelled) return;

        setResult(body);

        if (body.status === 'succeeded') {
          setBusy(false);
          return;
        }
        if (body.status === 'failed') {
          setBusy(false);
          setError(body.errorMessage || 'Import failed');
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
      const { blob, filename } = await adminService.exportTournamentData();

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
      const started = await adminService.startTournamentImport(file);
      setUploadId(started.uploadId);
      setResult({
        uploadId: started.uploadId,
        filename: started.filename,
        sha256: started.sha256,
        sizeBytes: started.sizeBytes,
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
          { label: 'Tournament Data' },
        ]}
      />
      <PageHeader title="Admin: Tournament Data" />

      <Card className="mb-6">
        <h2 className="text-xl font-semibold mb-2">Export</h2>
        <p className="text-gray-600 mb-4">Download a zip of the current DB state in bundle format.</p>
        <Button onClick={download} disabled={busy}>
          Download tournament data (.zip)
        </Button>
      </Card>

      <Card>
        <h2 className="text-xl font-semibold mb-2">Import</h2>
        <p className="text-gray-600 mb-4">Upload a tournament data zip and import it into the database.</p>

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
              <div className="font-medium">{(result.sizeBytes / 1024).toFixed(1)} KB</div>
              {result.importReport && (
                <>
                  <div className="col-span-2 border-t border-gray-200 mt-2 pt-2 font-semibold">Import Report</div>
                  <div className="text-gray-500">Schools</div>
                  <div className="font-medium">{result.importReport.schools}</div>
                  <div className="text-gray-500">Tournaments</div>
                  <div className="font-medium">{result.importReport.tournaments}</div>
                  <div className="text-gray-500">Teams</div>
                  <div className="font-medium">{result.importReport.tournamentTeams}</div>
                  <div className="text-gray-500">Calcuttas</div>
                  <div className="font-medium">{result.importReport.calcuttas}</div>
                  <div className="text-gray-500">Entries</div>
                  <div className="font-medium">{result.importReport.entries}</div>
                  <div className="text-gray-500">Bids</div>
                  <div className="font-medium">{result.importReport.bids}</div>
                  <div className="text-gray-500">Payouts</div>
                  <div className="font-medium">{result.importReport.payouts}</div>
                </>
              )}
              {result.verifyReport && (
                <>
                  <div className="col-span-2 border-t border-gray-200 mt-2 pt-2 font-semibold">Verify Report</div>
                  <div className="text-gray-500">OK</div>
                  <div className={`font-medium ${result.verifyReport.ok ? 'text-green-600' : 'text-red-600'}`}>
                    {result.verifyReport.ok ? 'Yes' : 'No'}
                  </div>
                  {result.verifyReport.mismatchCount > 0 && (
                    <>
                      <div className="text-gray-500">Mismatches</div>
                      <div className="font-medium text-red-600">{result.verifyReport.mismatchCount}</div>
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
}

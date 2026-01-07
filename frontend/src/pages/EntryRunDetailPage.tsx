import React, { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link, useParams } from 'react-router-dom';

import { ApiError } from '../api/apiClient';
import { Alert } from '../components/ui/Alert';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { entryRunsService, type EntryRunArtifactListItem } from '../services/entryRunsService';

export function EntryRunDetailPage() {
  const { runId } = useParams<{ runId: string }>();

  const showError = (err: unknown) => {
    if (err instanceof ApiError) {
      if (err.status === 403) return 'You do not have permission to view entry runs (403).';
      return `Request failed (${err.status}): ${err.message}`;
    }
    return err instanceof Error ? err.message : 'Unknown error';
  };

  const formatDateTime = (v: string | null | undefined) => {
    if (!v) return '—';
    const d = new Date(v);
    if (Number.isNaN(d.getTime())) return v;
    return d.toLocaleString();
  };

  const runQuery = useQuery({
    queryKey: ['entry-runs', 'get', runId],
    queryFn: () => entryRunsService.get(runId || ''),
    enabled: Boolean(runId),
  });

  const artifactsQuery = useQuery({
    queryKey: ['entry-runs', 'artifacts', 'list', runId],
    queryFn: () => entryRunsService.listArtifacts(runId || ''),
    enabled: Boolean(runId),
  });

  const artifacts: EntryRunArtifactListItem[] = useMemo(() => artifactsQuery.data?.items ?? [], [artifactsQuery.data?.items]);

  const metricsArtifactId = useMemo(() => {
    return artifacts.find((a) => a.artifact_kind === 'metrics')?.id ?? '';
  }, [artifacts]);

  return (
    <PageContainer className="max-w-none">
      <PageHeader
        title="Entry Run"
        subtitle={runQuery.data?.name || runId}
        leftActions={
          <Link to="/lab" className="text-blue-600 hover:text-blue-800">
            ← Back to Lab
          </Link>
        }
      />

      {!runId ? <Alert variant="error">Missing entry run ID.</Alert> : null}

      {runId && runQuery.isLoading ? <LoadingState label="Loading entry run..." /> : null}
      {runId && runQuery.isError ? (
        <Alert variant="error">
          <div className="font-semibold mb-1">Failed to load entry run</div>
          <div className="mb-3">{showError(runQuery.error)}</div>
          <Button size="sm" onClick={() => runQuery.refetch()}>
            Retry
          </Button>
        </Alert>
      ) : null}

      {runQuery.data ? (
        <div className="space-y-6">
          <Card>
            <h2 className="text-xl font-semibold mb-4">Run</h2>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
              <div>
                <div className="text-gray-500">ID</div>
                <div className="text-gray-900 font-medium break-words">{runQuery.data.id}</div>
              </div>
              <div>
                <div className="text-gray-500">Name</div>
                <div className="text-gray-900">{runQuery.data.name || '—'}</div>
              </div>
              <div>
                <div className="text-gray-500">Calcutta</div>
                <div className="text-gray-900 break-words">
                  {runQuery.data.calcutta_id ? (
                    <Link to={`/calcuttas/${encodeURIComponent(runQuery.data.calcutta_id)}`} className="text-blue-600 hover:text-blue-800">
                      {runQuery.data.calcutta_id}
                    </Link>
                  ) : (
                    '—'
                  )}
                </div>
              </div>
              <div>
                <div className="text-gray-500">Optimizer</div>
                <div className="text-gray-900">{runQuery.data.optimizer_key}</div>
              </div>
              <div>
                <div className="text-gray-500">Purpose</div>
                <div className="text-gray-900">{runQuery.data.purpose}</div>
              </div>
              <div>
                <div className="text-gray-500">Created</div>
                <div className="text-gray-900">{formatDateTime(runQuery.data.created_at)}</div>
              </div>
            </div>
          </Card>

          <Card>
            <div className="flex items-end justify-between gap-4 mb-4">
              <h2 className="text-xl font-semibold">Artifacts</h2>
              {metricsArtifactId ? (
                <Link
                  to={`/lab/entry-artifacts/${encodeURIComponent(metricsArtifactId)}`}
                  className="text-blue-600 hover:text-blue-800 text-sm"
                >
                  View metrics
                </Link>
              ) : null}
            </div>

            {artifactsQuery.isLoading ? <LoadingState label="Loading artifacts..." layout="inline" /> : null}
            {artifactsQuery.isError ? (
              <Alert variant="error" className="mt-3">
                <div className="font-semibold mb-1">Failed to load artifacts</div>
                <div className="mb-3">{showError(artifactsQuery.error)}</div>
                <Button size="sm" onClick={() => artifactsQuery.refetch()}>
                  Retry
                </Button>
              </Alert>
            ) : null}

            {!artifactsQuery.isLoading && !artifactsQuery.isError && artifacts.length === 0 ? (
              <Alert variant="info" className="mt-3">
                No artifacts found for this entry run.
              </Alert>
            ) : null}

            {!artifactsQuery.isLoading && !artifactsQuery.isError && artifacts.length > 0 ? (
              <div className="overflow-x-auto">
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Kind</th>
                      <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Schema</th>
                      <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Created</th>
                      <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Lineage</th>
                      <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Open</th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-200">
                    {artifacts.map((a) => (
                      <tr key={a.id} className="hover:bg-gray-50">
                        <td className="px-3 py-2 text-sm text-gray-900 font-medium">{a.artifact_kind}</td>
                        <td className="px-3 py-2 text-sm text-gray-700">{a.schema_version}</td>
                        <td className="px-3 py-2 text-sm text-gray-700">{formatDateTime(a.created_at)}</td>
                        <td className="px-3 py-2 text-xs text-gray-600 break-words">
                          <div>market_share={a.input_market_share_artifact_id || '—'}</div>
                          <div>advancement={a.input_advancement_artifact_id || '—'}</div>
                        </td>
                        <td className="px-3 py-2 text-sm text-right">
                          <Link
                            to={`/lab/entry-artifacts/${encodeURIComponent(a.id)}`}
                            className="text-blue-600 hover:text-blue-800"
                          >
                            View
                          </Link>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            ) : null}
          </Card>
        </div>
      ) : null}
    </PageContainer>
  );
}
